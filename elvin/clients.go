// Copyright 2018 Cobaro Pty Ltd. All Rights Reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package elvin

import (
	"encoding/binary"
	_ "errors"
	"fmt"
	"io"
	"log"
	"sync/atomic"
)

// Transaction IDs on packets
func XID() uint32 {
	return atomic.AddUint32(&xID, 1)
}

// private
var xID uint32 = 0

// Client connection states used for sanity and to enforce protocol rules
const (
	StateClosed = iota
	StateConnecting
	StateConnected
	StateDisconnecting
)

// Return state (synchronized)
func (client *Client) State() uint32 {
	return atomic.LoadUint32(&client.state)
}

// Get state (synchronized)
func (client *Client) SetState(val uint32) {
	atomic.StoreUint32(&client.state, val)
}

// This closes a client's sockets/endpoints and cleans state
// returning things to where they were following a NewClient()
// with the exception that the subscription list is maintained
// so it can be re-established on re-connection
func (client *Client) Close() {
	client.mu.Lock()
	client.SetState(StateClosed)
	select {
	case client.writeTerminate <- 1:
	default:
	}
	// client.readTerminate <- 1
	client.closer.Close()
	client.subReplies = make(map[uint32]*Subscription)
	client.connXID = 0
	client.disconnXID = 0
	client.wg.Wait() // Wait for reader and writer to finish
	client.mu.Unlock()
}

// Read n bytes from reader into buffer which must be big enough
func readBytes(reader io.Reader, buffer []byte, numToRead int) (int, error) {
	offset := 0
	for offset < numToRead {
		//log.Printf("offset = %d, numToRead = %d", offset, numToRead)
		length, err := reader.Read(buffer[offset:numToRead])
		if err != nil {
			return offset + length, err
		}
		offset += length
	}
	return offset, nil
}

// Handle reading for now run as a goroutine
func (client *Client) readHandler() {
	header := make([]byte, 4)
	buffer := make([]byte, 2048)

	for {
		// Read frame header
		length, err := readBytes(client.reader, header, 4)
		if length != 4 || err != nil {
			break // We're done
		}

		// Read the protocol packet, starting with it's length
		packetSize := int32(binary.BigEndian.Uint32(header))
		// Grow our buffer if needed
		if int(packetSize) > len(buffer) {
			buffer = make([]byte, packetSize)
		}

		length, err = readBytes(client.reader, buffer, int(packetSize))
		if length != int(packetSize) || err != nil {
			break // We're done
		}

		// Deal with the packet
		if err = client.HandlePacket(buffer); err != nil {
			log.Printf("Read Handler error: %v", err)
			// FIXME: protocol error
			break // We're done
		}

	}

	// Tell the client we lost the connection if we're supposed to be open
	// otherwise this can socket closure on shutdown or redirect etc
	if client.State() == StateConnected {
		client.Close()
		disconn := new(Disconn)
		disconn.Reason = DisconnReasonClientConnectionLost
		select {
		case client.DisconnChannel <- disconn:
		default:
		}
	}

	client.wg.Done()
	//log.Printf("read handler exiting")
}

// Handle writing for now run as a goroutine
func (client *Client) writeHandler() {
	header := make([]byte, 4)

	for {
		select {
		case buffer := <-client.writeChannel:

			// Write the frame header (packetsize)
			//log.Printf("write header")
			binary.BigEndian.PutUint32(header, uint32(buffer.Len()))
			_, err := client.writer.Write(header)
			if err != nil {
				// Deal with more errors
				if err != io.EOF {
					log.Printf("Unexpected write error: %v", err)
				}
				client.wg.Done()
				return
			}

			// Write the packet
			_, err = buffer.WriteTo(client.writer)
			if err != nil {
				// Deal with more errors
				if err != io.EOF {
					log.Printf("Unexpected write error: %v", err)
				}
				client.wg.Done()
				return
			}
		case <-client.writeTerminate:
			// log.Printf("writeHandler exiting")
			client.wg.Done()
			return
		}
	}
}

// Handle a protocol packet
func (client *Client) HandlePacket(buffer []byte) (err error) {

	// log.Printf("HandlePacket received %v (%d)", PacketIDString(PacketID(buffer)), client.State())

	// Packets accepted independent of Client's connection state
	switch PacketID(buffer) {
	case PacketDisconn:
		return client.HandleDisconn(buffer)
	}

	// Packets dependent upon Client's connection state
	switch client.State() {
	case StateConnecting:
		switch PacketID(buffer) {
		case PacketConnRply:
			return client.HandleConnRply(buffer)
		case PacketNack:
			return client.HandleNack(buffer)
		default:
			return fmt.Errorf("ProtocolError: %s received", PacketIDString(PacketID(buffer)))
		}

	case StateDisconnecting:
		switch PacketID(buffer) {
		case PacketDisconnRply:
			return client.HandleDisconnRply(buffer)
		}

	case StateConnected:
		switch PacketID(buffer) {
		case PacketSubRply:
			return client.HandleSubRply(buffer)
		case PacketNotifyDeliver:
			return client.HandleNotifyDeliver(buffer)
		case PacketNack:
			return client.HandleNack(buffer)
		case PacketDropWarn:
		case PacketReserved:
		case PacketQnchRply:
		case PacketSubAddNotify:
		case PacketSubModNotify:
		case PacketSubDelNotify:
			return fmt.Errorf("FIXME implement: %s received", PacketIDString(PacketID(buffer)))
		default:
			return fmt.Errorf("ProtocolError: %s received", PacketIDString(PacketID(buffer)))
		}

	case StateClosed:
		return fmt.Errorf("ProtocolError: %s received", PacketIDString(PacketID(buffer)))
	}

	return fmt.Errorf("Error: %s received and not handled", PacketIDString(PacketID(buffer)))
}

// On a protocol error we want to alert the client and reset the connection
func (client *Client) ProtocolError(err error) {

	// Log
	log.Println(err)

	// Kill the connection and clean up
	client.Close()

	// Tell the client (if they are listening)
	disconn := new(Disconn)
	disconn.Reason = DisconnReasonClientConnectionLost
	select {
	case client.DisconnChannel <- disconn:
	default:
	}

}

// Handle a Connection Reply
func (client *Client) HandleConnRply(buffer []byte) (err error) {
	connRply := new(ConnRply)
	if err = connRply.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}

	// We're now connected
	client.SetState(StateConnected)

	// FIXME; check options
	// connRply.Options

	// Signal the connection requestor
	client.connReplies <- connRply
	return nil
}

// Handle a Disconnection reply
func (client *Client) HandleDisconnRply(buffer []byte) (err error) {
	disconnRply := new(DisconnRply)
	if err = disconnRply.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}
	// Signal the disconnection requestor
	client.disconnReplies <- disconnRply
	return nil
}

// Handle a Disconn
func (client *Client) HandleDisconn(buffer []byte) (err error) {
	disconn := new(Disconn)
	if err = disconn.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}

	// Signal the disconect
	// If a client library isn't listening we just close the client
	select {
	case client.DisconnChannel <- disconn:
	default:
		// Reset the client
		client.Close()
	}

	return nil
}

// Handle a Nack
func (client *Client) HandleNack(buffer []byte) (err error) {
	nack := new(Nack)
	if err = nack.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}

	// A Nack can belong to multiple places so hunt it down
	client.mu.Lock()
	defer client.mu.Unlock()

	sub, ok := client.subReplies[nack.XID]
	if ok {
		delete(client.subReplies, nack.XID)
		sub.events <- Packet(nack)
		return nil
	}

	if client.connXID == nack.XID {
		client.connXID = 0
		client.connReplies <- Packet(nack)
		return nil
	}

	// FIXME: Quench packets

	return fmt.Errorf("Unhandled nack xid=%d, (conn:%d)\n", nack.XID, client.connXID)
}

// Handle a Subscription reply
func (client *Client) HandleSubRply(buffer []byte) (err error) {
	subRply := new(SubRply)
	if err = subRply.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}

	client.mu.Lock()
	sub, ok := client.subReplies[subRply.XID]
	delete(client.subReplies, subRply.XID)
	client.mu.Unlock()
	if !ok {
		client.Close()
		// FIXME: return error
	}

	// Signal the connection request
	sub.events <- Packet(subRply)
	return nil
}

// Handle a Notification Deliver
func (client *Client) HandleNotifyDeliver(buffer []byte) (err error) {
	notifyDeliver := new(NotifyDeliver)
	if err = notifyDeliver.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}

	// Sync the map of subIDs. We can do this once as:
	// * If one disappears it's ok (we don't deliver)
	// * If one appears it's ok (they're sparse)
	client.mu.Lock()
	delivers := client.subscriptions
	client.mu.Unlock()

	// foreach matching subscription deliver it
	for _, subID := range notifyDeliver.Secure {
		log.Printf("NotifyDeliver secure for %d", subID)
		sub, ok := delivers[subID]
		if ok && sub.subID == subID {
			sub.Notifications <- notifyDeliver.NameValue
		}
	}
	for _, subID := range notifyDeliver.Insecure {
		sub, ok := client.subscriptions[subID]
		// log.Printf("NotifyDeliver insecure for %d", subID)
		// log.Printf("client.subDelivers = %v", client.subDelivers)
		if ok && sub.subID == subID {
			sub.Notifications <- notifyDeliver.NameValue
		}
	}
	return nil
}
