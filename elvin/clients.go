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
	"fmt"
	"github.com/cobaro/elvin/elog"
	"io"
	"math/rand"
	"os"
	"sync/atomic"
	"time"
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
	StateOpen
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
	case client.writeTerminate <- 1: // Will close the socket
	default:
		client.closer.Close()
	}
	// client.readTerminate <- 1
	client.subReplies = make(map[uint32]*Subscription)
	client.quenchReplies = make(map[uint32]*Quench)
	client.connXID = 0
	client.disconnXID = 0
	client.mu.Unlock()
	client.wg.Wait() // Wait for reader and writer to finish
}

// Read n bytes from reader into buffer which must be big enough
func readBytes(reader io.Reader, buffer []byte, numToRead int) (int, error) {
	offset := 0
	for offset < numToRead {
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
			client.elog.Logf(elog.LogLevelError, "Read Handler error: %v", err)
			// FIXME: protocol error
			// Or if say a disconnect timed out
			// then should we exit?
		}

	}

	// Tell the write handler to exit too
	select {
	case client.writeTerminate <- 1:
	default:
	}

	// Tell the client we lost the connection if we're supposed to be open
	// otherwise this can be socket closure on shutdown or redirect etc
	if client.State() == StateConnected {
		disconn := new(Disconn)
		disconn.Reason = DisconnReasonClientConnectionLost
		select {
		case client.Events <- disconn:
		default:
			go client.ConnectionEventsDefault(disconn)
		}
	}

	client.wg.Done()
	client.elog.Logf(elog.LogLevelDebug2, "read handler exiting")
}

// Handle writing for now run as a goroutine
func (client *Client) writeHandler() {
	header := make([]byte, 4)

	defer client.Close()
	for {
		select {
		case buffer := <-client.writeChannel:

			// Write the frame header (packetsize)
			binary.BigEndian.PutUint32(header, uint32(buffer.Len()))
			_, err := client.writer.Write(header)
			if err != nil {
				// Deal with more errors
				if err != io.EOF {
					client.elog.Logf(elog.LogLevelWarning, "Unexpected write error: %v", err)
				}
				client.wg.Done()
				return
			}

			// Write the packet
			_, err = buffer.WriteTo(client.writer)
			if err != nil {
				// Deal with more errors
				if err != io.EOF {
					client.elog.Logf(elog.LogLevelWarning, "Unexpected write error: %v", err)
				}
				client.wg.Done()
				return
			}
		case <-client.writeTerminate:
			client.elog.Logf(elog.LogLevelDebug2, "Write handler exiting")
			client.wg.Done()
			return
		}
	}
}

// Handle a protocol packet
func (client *Client) HandlePacket(buffer []byte) (err error) {

	client.elog.Logf(elog.LogLevelDebug3, "HandlePacket received %v (%d)", PacketIDString(PacketID(buffer)), client.State())

	// Packets accepted independent of Client's connection state
	switch PacketID(buffer) {
	case PacketReserved:
		return nil
	case PacketNack:
		return client.HandleNack(buffer)
	case PacketTestConn:
		return client.HandleTestConn(buffer)
	case PacketConfConn:
		return client.HandleConfConn(buffer)
	case PacketDisconn:
		return client.HandleDisconn(buffer)
	}

	// Packets dependent upon Client's connection state
	switch client.State() {
	case StateConnecting:
		switch PacketID(buffer) {
		case PacketConnReply:
			return client.HandleConnReply(buffer)
		default:
			return LocalError(ErrorsProtocolPacketStateNotConnected, PacketIDString(PacketID(buffer)))
		}

	case StateDisconnecting:
		switch PacketID(buffer) {
		case PacketDisconnReply:
			return client.HandleDisconnReply(buffer)
		}

	case StateConnected:
		switch PacketID(buffer) {
		case PacketSubReply:
			return client.HandleSubReply(buffer)
		case PacketQuenchReply:
			return client.HandleQuenchReply(buffer)
		case PacketNotifyDeliver:
			return client.HandleNotifyDeliver(buffer)
		case PacketSubAddNotify:
			return client.HandleSubAddNotify(buffer)
		case PacketSubModNotify:
			return client.HandleSubModNotify(buffer)
		case PacketSubDelNotify:
			return client.HandleSubDelNotify(buffer)
		case PacketDropWarn:
			return client.HandleDropWarn(buffer)
		default:
			return LocalError(ErrorsProtocolPacketStateIsConnected, PacketIDString(PacketID(buffer)))
		}

	case StateClosed:
		return LocalError(ErrorsProtocolPacketStateNotConnected, PacketIDString(PacketID(buffer)))
	}

	return LocalError(ErrorsBadPacketType, PacketIDString(PacketID(buffer)))
}

// This function is called by the library if the client has not
// registered for the notification channel. It provides an example
// of what event types can occur and some default behaviour
func (client *Client) ConnectionEventsDefault(event Packet) {
	switch event.(type) {
	case *Disconn:
		disconn := event.(*Disconn)
		client.elog.Logf(elog.LogLevelDebug3, "Received Disconn:\n%+v", disconn)
		switch disconn.Reason {

		case DisconnReasonRouterShuttingDown:
			client.elog.Logf(elog.LogLevelError, "router shutting down, exiting")
			os.Exit(1)

		case DisconnReasonRouterProtocolErrors:
			client.elog.Logf(elog.LogLevelError, "router detected protocol violation")
			os.Exit(1)

		case DisconnReasonRouterRedirect:
			if len(disconn.Args) > 0 {
				client.elog.Logf(elog.LogLevelInfo1, "redirected to %s", disconn.Args)
				client.Endpoint = disconn.Args
				client.Close()
				if err := client.Connect(); err != nil {
					client.elog.Logf(elog.LogLevelError, "%v", err)
					os.Exit(1)
				}
				client.elog.Logf(elog.LogLevelInfo1, "connected to %s", client.Endpoint)
			} else {
				client.elog.Logf(elog.LogLevelError, "Disconn to nowhere")
				os.Exit(1)
			}
			break

		case DisconnReasonClientConnectionLost:
			client.elog.Logf(elog.LogLevelWarning, "Lost connection to %s, reconnecting", client.Endpoint)
			if err := client.DefaultReconnect(10, time.Duration(0), time.Minute*2); err != nil {
				client.elog.Logf(elog.LogLevelError, "Giving up reconnecting")
				os.Exit(1)
			}
			client.elog.Logf(elog.LogLevelWarning, "Reconnected")

		case DisconnReasonClientProtocolErrors:
			client.elog.Logf(elog.LogLevelError, "client library detected protocol errors")
			os.Exit(1)
		}
	case *DropWarn:
		client.elog.Logf(elog.LogLevelWarning, "DropWarn (lost one or more packets)")

	default:
		client.elog.Logf(elog.LogLevelError, "FIXME: bad connection notification")
		os.Exit(1)
	}
}

// The default behaviour for reconnection handling.
// Will retry forever if retries is 0.
func (client *Client) DefaultReconnect(retries int, minWait time.Duration, maxWait time.Duration) (err error) {

	// Add up to 50ms randomness to initial backoff
	wait := minWait + time.Duration(rand.Intn(50))*time.Millisecond

	for {
		time.Sleep(wait)
		err = client.Connect()
		if err == nil {
			// We connected, so resubscribe, requench
			// If anything fails here we cleanup
			subs := client.subscriptions
			client.subscriptions = make(map[int64]*Subscription)
			for _, sub := range subs {
				if err = client.Subscribe(sub); err != nil {
					client.subscriptions = subs
					client.Disconnect()
					return
				}
			}
			// We connected, so resubscribe, requench
			// If anything fails here we cleanup
			quenches := client.quenches
			client.quenches = make(map[int64]*Quench)
			for _, quench := range quenches {
				if err = client.Quench(quench); err != nil {
					client.quenches = quenches
					client.Disconnect()
					return
				}
			}
			return
		}

		// If retries is zero we loop effectively forever
		retries--
		if retries == 0 {
			return
		}

		wait *= 4
		if wait > maxWait {
			wait = maxWait
		}
	}
}

// On a protocol error we want to alert the client and reset the connection
func (client *Client) ProtocolError(err error) {

	// Log
	client.elog.Logf(elog.LogLevelError, "%s", err.Error())

	// Tell the client (if they are listening)
	disconn := new(Disconn)
	disconn.Reason = DisconnReasonClientConnectionLost
	select {
	case client.Events <- disconn:
	default:
		go client.ConnectionEventsDefault(disconn)
	}
}

// Handle a Connection Reply
func (client *Client) HandleConnReply(buffer []byte) (err error) {
	connReply := new(ConnReply)
	if err = connReply.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}

	// We're now connected
	client.SetState(StateConnected)

	// FIXME; check options
	// connReply.Options

	// Signal the connection requestor
	client.connReplies <- connReply
	return nil
}

// Handle a Disconnection reply
func (client *Client) HandleDisconnReply(buffer []byte) (err error) {
	disconnReply := new(DisconnReply)
	if err = disconnReply.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}
	// Signal the disconnection requestor
	client.connReplies <- disconnReply
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
	case client.Events <- disconn:
	default:
		go client.ConnectionEventsDefault(disconn)
	}

	return nil
}

// Handle a DropWarn
func (client *Client) HandleDropWarn(buffer []byte) (err error) {
	dropWarn := new(DropWarn)
	// Nothing to decode

	// Signal the DropWarn
	// If a client library isn't listening we ignore it
	select {
	case client.Events <- dropWarn:
	default:
		go client.ConnectionEventsDefault(dropWarn)
	}

	return nil
}

// Handle a TestConn
func (client *Client) HandleTestConn(buffer []byte) (err error) {
	// Nothing to decode

	// Respond
	confConn := new(ConfConn)
	writeBuf := new(bytes.Buffer)
	confConn.Encode(writeBuf)
	client.writeChannel <- writeBuf

	return nil
}

// Handle a TestConn
func (client *Client) HandleConfConn(buffer []byte) (err error) {
	// Nothing to decode

	// Respond if listening
	select {
	case client.confConn <- true:
	default:
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

	quench, ok := client.quenchReplies[nack.XID]
	if ok {
		delete(client.quenchReplies, nack.XID)
		quench.events <- Packet(nack)
		return nil
	}

	if client.connXID == nack.XID {
		client.connXID = 0
		client.connReplies <- Packet(nack)
		return nil
	}

	return fmt.Errorf("Unhandled nack xid=%d, (conn:%d)\n", nack.XID, client.connXID)
}

// Handle a Subscription reply
func (client *Client) HandleSubReply(buffer []byte) (err error) {
	subReply := new(SubReply)
	if err = subReply.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}

	client.mu.Lock()
	sub, ok := client.subReplies[subReply.XID]
	client.mu.Unlock()
	if ok {
		// Signal the subscription
		delete(client.subReplies, subReply.XID)
		sub.events <- Packet(subReply)
	} // else it will time out
	return nil
}

// Handle a Qeunch reply
func (client *Client) HandleQuenchReply(buffer []byte) (err error) {
	quenchReply := new(QuenchReply)
	if err = quenchReply.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}

	client.mu.Lock()
	quench, ok := client.quenchReplies[quenchReply.XID]
	client.mu.Unlock()
	if ok {
		delete(client.quenchReplies, quenchReply.XID)
		quench.events <- Packet(quenchReply)
	} // else it will time out
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
	subscriptions := client.subscriptions
	client.mu.Unlock()

	// foreach matching subscription deliver it
	for _, subID := range notifyDeliver.Secure {
		client.elog.Logf(elog.LogLevelDebug3, "NotifyDeliver secure for %d", subID)
		sub, ok := subscriptions[subID]
		if ok && sub.subID == subID {
			sub.Notifications <- notifyDeliver.NameValue
		}
	}
	for _, subID := range notifyDeliver.Insecure {
		client.elog.Logf(elog.LogLevelDebug3, "NotifyDeliver insecure for %d", subID)
		sub, ok := client.subscriptions[subID]
		if ok && sub.subID == subID {
			sub.Notifications <- notifyDeliver.NameValue
		}
	}
	return nil
}

// Handle a quench's SubAddNotify
func (client *Client) HandleSubAddNotify(buffer []byte) (err error) {
	subAddNotify := new(SubAddNotify)
	if err = subAddNotify.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}

	// Sync the map of quench IDs. We can do this once as:
	// * If one disappears it's ok (we don't deliver)
	// * If one appears it's ok (they're sparse)
	client.mu.Lock()
	quenches := client.quenches
	client.mu.Unlock()

	notification := QuenchNotification{subAddNotify.TermID, subAddNotify.SubExpr}
	// foreach matching quench deliver it
	for _, quenchID := range subAddNotify.SecureQuenchIDs {
		client.elog.Logf(elog.LogLevelDebug3, "QuenchAddNotify secure for %d", quenchID)
		quench, ok := quenches[quenchID]
		if ok && quench.quenchID == quenchID {
			quench.Notifications <- notification
		}
	}
	for _, quenchID := range subAddNotify.InsecureQuenchIDs {
		client.elog.Logf(elog.LogLevelDebug3, "QuenchAddNotify insecure for %d", quenchID)
		quench, ok := quenches[quenchID]
		if ok && quench.quenchID == quenchID {
			quench.Notifications <- notification
		}
	}
	return nil
}

// Handle a quench's SubModNotify
func (client *Client) HandleSubModNotify(buffer []byte) (err error) {
	subModNotify := new(SubModNotify)
	if err = subModNotify.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}

	// Sync the map of quench IDs. We can do this once as:
	// * If one disappears it's ok (we don't deliver)
	// * If one appears it's ok (they're sparse)
	client.mu.Lock()
	quenches := client.quenches
	client.mu.Unlock()

	notification := QuenchNotification{subModNotify.TermID, subModNotify.SubExpr}
	// foreach matching quench deliver it
	for _, quenchID := range subModNotify.SecureQuenchIDs {
		client.elog.Logf(elog.LogLevelDebug3, "QuenchModNotify secure for %d", quenchID)
		quench, ok := quenches[quenchID]
		if ok && quench.quenchID == quenchID {
			quench.Notifications <- notification
		}
	}
	for _, quenchID := range subModNotify.InsecureQuenchIDs {
		client.elog.Logf(elog.LogLevelDebug3, "QuenchModNotify insecure for %d", quenchID)
		quench, ok := quenches[quenchID]
		if ok && quench.quenchID == quenchID {
			quench.Notifications <- notification
		}
	}
	return nil
}

// Handle a quench's SubDelNotify
func (client *Client) HandleSubDelNotify(buffer []byte) (err error) {
	subDelNotify := new(SubDelNotify)
	if err = subDelNotify.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}

	// Sync the map of quench IDs. We can do this once as:
	// * If one disappears it's ok (we don't deliver)
	// * If one appears it's ok (they're sparse)
	client.mu.Lock()
	quenches := client.quenches
	client.mu.Unlock()

	// FIXME: AST is mock (should be empty for delete)
	notification := QuenchNotification{subDelNotify.TermID, SubAST{1}}
	for _, quenchID := range subDelNotify.QuenchIDs {
		client.elog.Logf(elog.LogLevelDebug3, "QuenchDelNotify for %d", quenchID)
		quench, ok := quenches[quenchID]
		if ok && quench.quenchID == quenchID {
			quench.Notifications <- notification
		}
	}
	return nil
}

// Seed the random number generator
func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
