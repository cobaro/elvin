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
	"bytes"
	"encoding/binary"
	_ "errors"
	"fmt"
	"io"
	"log"
	"sync/atomic"
)

// Transaction Ids on packets
func Xid() uint32 {
	return atomic.AddUint32(&xid, 1)
}

// private
var xid uint32 = 0

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

// Create a new client. This must be called as using new(Client) will
// not result in proper initialization
func NewClient(endpoint string, options map[string]interface{}, keysNfn []Keyset, keysSub []Keyset) (conn *Client) {
	client := new(Client)
	client.Endpoint = endpoint
	client.Options = options
	client.KeysNfn = keysNfn
	client.KeysSub = keysSub
	client.writeChannel = make(chan *bytes.Buffer)
	client.readTerminate = make(chan int)
	client.writeTerminate = make(chan int)
	// Async packets
	client.subDelivers = make(map[uint64]*Subscription)
	// Sync Packets
	client.connRply = make(chan *ConnRply)
	client.disconnRply = make(chan *DisconnRply)
	client.subRplys = make(map[uint32]*Subscription)
	return client
}

// Handle reading for now run as a goroutine
func (client *Client) readHandler() {
	defer client.closer.Close()

	header := make([]byte, 4)
	buffer := make([]byte, 2048)

	for {
		// Read frame header
		length, err := readBytes(client.reader, header, 4)
		if length != 4 || err != nil {
			// Deal with more errors
			if err == io.EOF {
				client.writeTerminate <- 1
			} else {
				if client.State() != StateClosed {
					log.Printf("Read Handler error: %v", err)
				}
			}
			return // We're done
		}

		// Read the protocol packet, starting with it's length
		packetSize := int32(binary.BigEndian.Uint32(header))
		// Grow our buffer if needed
		if int(packetSize) > len(buffer) {
			buffer = make([]byte, packetSize)
		}

		length, err = readBytes(client.reader, buffer, int(packetSize))
		if err != nil {
			// Deal with more errors
			if err == io.EOF {
				client.writeTerminate <- 1
			} else {
				if client.State() != StateClosed {
					log.Printf("Read Handler error: %v", err)
				}
			}
			return // We're done
		}

		// Deal with the packet
		if err = client.HandlePacket(buffer); err != nil {
			log.Printf("Read Handler error: %v", err)
			// FIXME: protocol error
		}

	}
}

// Handle writing for now run as a goroutine
func (client *Client) writeHandler() {
	defer client.closer.Close()
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
				if err == io.EOF {
					client.closer.Close()
				} else {
					log.Printf("Unexpected write error: %v", err)
				}
				return // We're done, cleanup done by read
			}

			// Write the packet
			_, err = buffer.WriteTo(client.writer)
			if err != nil {
				// Deal with more errors
				if err == io.EOF {
					client.closer.Close()
				} else {
					log.Printf("Unexpected write error: %v", err)
				}
				return // We're done, cleanup done by read
			}
		case <-client.writeTerminate:
			return
		}
	}
}

// Handle a protocol packet
func (client *Client) HandlePacket(buffer []byte) (err error) {

	// Packets dependent upon Client's connection state
	switch client.State() {
	case StateConnecting:
		switch PacketId(buffer) {
		case PacketConnRply:
			return client.HandleConnRply(buffer)
		default:
			return fmt.Errorf("ProtocolError: %s received", PacketIdString(PacketId(buffer)))
		}

	case StateDisconnecting:
		switch PacketId(buffer) {
		case PacketDisconnRply:
			return client.HandleDisconnRply(buffer)
		}

	case StateConnected:
		switch PacketId(buffer) {
		case PacketSubRply:
			return client.HandleSubRply(buffer)
		case PacketNotifyDeliver:
			return client.HandleNotifyDeliver(buffer)
		case PacketDropWarn:
		case PacketReserved:
		case PacketNack:
		case PacketDisconn:
		case PacketQnchRply:
		case PacketSubAddNotify:
		case PacketSubModNotify:
		case PacketSubDelNotify:
			return fmt.Errorf("FIXME implement: %s received", PacketIdString(PacketId(buffer)))
		default:
			return fmt.Errorf("ProtocolError: %s received", PacketIdString(PacketId(buffer)))
		}

	case StateClosed:
		return fmt.Errorf("ProtocolError: %s received", PacketIdString(PacketId(buffer)))
	}

	return fmt.Errorf("Error: %s received and not handled", PacketIdString(PacketId(buffer)))
}

// Handle a Connection Reply
func (client *Client) HandleConnRply(buffer []byte) (err error) {
	connRply := new(ConnRply)
	if err = connRply.Decode(buffer); err != nil {
		client.SetState(StateClosed)
		client.closer.Close()
		// FIXME: return error
	}

	// We're now connected
	client.SetState(StateConnected)

	// FIXME; check options
	// connRply.Options

	// Signal the connection request
	client.connRply <- connRply
	return nil
}

// Handle a Disconenction reply
func (client *Client) HandleDisconnRply(buffer []byte) (err error) {
	disconnRply := new(DisconnRply)
	if err = disconnRply.Decode(buffer); err != nil {
		client.closer.Close()
		// FIXME: return error
	}

	// We're now disconnected
	client.SetState(StateClosed)
	client.subRplys = nil    // harsh but fair
	client.subDelivers = nil // harsh but fair

	// Signal the connection request
	client.disconnRply <- disconnRply
	return nil
}

// Handle a Subscription reply
func (client *Client) HandleSubRply(buffer []byte) (err error) {
	subRply := new(SubRply)
	if err = subRply.Decode(buffer); err != nil {
		client.closer.Close()
		// FIXME: return error
	}

	client.mu.Lock()
	sub, ok := client.subRplys[subRply.Xid]
	delete(client.subRplys, subRply.Xid)
	client.mu.Unlock()
	if !ok {
		client.closer.Close()
		// FIXME: return error
	}

	var ev SubscriptionEvent
	ev.eventType = subEventSubRply
	ev.subRply = subRply

	// Signal the connection request
	sub.events <- ev
	return nil
}

// Handle a Notification Deliver
func (client *Client) HandleNotifyDeliver(buffer []byte) (err error) {
	notifyDeliver := new(NotifyDeliver)
	if err = notifyDeliver.Decode(buffer); err != nil {
		client.closer.Close()
		// FIXME: return error
	}

	// Sync the map of subids. We can do this once as:
	// * If one disappears it's ok (we don't deliver)
	// * If one appears it's ok (subids are sparse)
	client.mu.Lock()
	delivers := client.subDelivers
	client.mu.Unlock()

	// foreach matching subscription deliver it
	for _, subid := range notifyDeliver.Secure {
		log.Printf("NotifyDeliver secure for %d", subid)
		sub, ok := delivers[subid]
		if ok && sub.subId == subid {
			sub.Notifications <- notifyDeliver.NameValue
		}
	}
	for _, subid := range notifyDeliver.Insecure {
		sub, ok := client.subDelivers[subid]
		// log.Printf("NotifyDeliver insecure for %d", subid)
		// log.Printf("client.subDelivers = %v", client.subDelivers)
		if ok && sub.subId == subid {
			sub.Notifications <- notifyDeliver.NameValue
		}
	}
	return nil
}
