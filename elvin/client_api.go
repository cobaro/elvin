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
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// A client of an Elvin service, typically used via:
//      client.Connect()
//      client = newClient()
//        client.Subscribe()
//        client.Quench()
//        client.Notify()
//      client.Disonnect()
// See individual methods for details
type Client struct {
	URL      string                 // Router descriptor
	Protocol *Protocol              // Router specification
	Options  map[string]interface{} // Router options
	KeysNfn  KeyBlock               // Connections keys for outgoing notifications
	KeysSub  KeyBlock               // Connections keys for incoming notifications
	Events   chan Packet            // Clients may listen here for connectionq events
	elog     elog.Elog              // Logging

	// Private
	reader         io.Reader
	writer         io.Writer
	closer         io.Closer
	state          uint32
	writeChannel   chan *bytes.Buffer
	readTerminate  chan int
	writeTerminate chan int
	mu             sync.Mutex
	wg             sync.WaitGroup

	// Maps of all current subscriptions used for mapping
	// NotifyDelivers and for maintaining subscriptions across
	// reconnection
	subReplies    map[uint32]*Subscription // map SubAdd/Mod/Del/Nack
	subscriptions map[int64]*Subscription  // All our subscriptions

	// Maps of all current quenches used for mapping quench
	// Notifications and for maintaining quenches across
	// reconnection
	quenchReplies map[uint32]*Quench // map QuenchAdd/Mod/Del/Nack
	quenches      map[int64]*Quench  // All our quenches

	// Connection level packets
	connReplies chan Packet // receive ConnReply, DisconnReply, DropWarn
	connXID     uint32      // XID of any outstanding connrqst
	disconnXID  uint32      // XID of any outstanding disconnrqst
	confConn    chan bool   // signal testConn complete
}

// FIXME: define and maybe make configurable?
const ConnectTimeout = (10 * time.Second)
const DisconnectTimeout = (10 * time.Second)
const SubscriptionTimeout = (10 * time.Second)
const QuenchTimeout = (10 * time.Second)
const TestConnTimeout = (10 * time.Second)

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

// A subscription type used by clients.
type Subscription struct {
	Expression     string                      // Subscription Expression
	AcceptInsecure bool                        // Do we accept notifications with no security keys
	Keys           KeyBlock                    // Keys for this subscriptions
	Notifications  chan map[string]interface{} // Notifications delivered on this channel

	subID  int64       // private id
	events chan Packet // synchronous replies
}

func (sub *Subscription) addKeys(keys KeyBlock) {
	// FIXME: implement
	return
}

func (sub *Subscription) delKeys(keys KeyBlock) {
	// FIXME: implement
	return
}

type QuenchNotification struct {
	TermID  uint64
	SubExpr SubAST
}

// The Quench type used by clients.
type Quench struct {
	Names           map[string]bool         // Quench terms
	DeliverInsecure bool                    // Deliver with no security keys?
	Keys            KeyBlock                // Keys for this quench
	Notifications   chan QuenchNotification // Sub{Add|Del|Mod}Notify delivers
	quenchID        int64                   // private id
	events          chan Packet             // synchronous replies
}

func (quench *Quench) addKeys(keys KeyBlock) {
	// FIXME: implement
	return
}

func (quench *Quench) delKeys(keys KeyBlock) {
	// FIXME: implement
	return
}

// Create a new client.
// Using new(Client) will not result in proper initialization
func NewClient(url string, options map[string]interface{}, keysNfn KeyBlock, keysSub KeyBlock) (conn *Client) {
	client := new(Client)
	client.URL = url
	client.Options = options
	client.KeysNfn = keysNfn
	client.KeysSub = keysSub
	client.writeChannel = make(chan *bytes.Buffer)
	client.readTerminate = make(chan int)
	client.writeTerminate = make(chan int)
	client.subscriptions = make(map[int64]*Subscription)
	client.quenches = make(map[int64]*Quench)
	// Sync Packets
	client.connReplies = make(chan Packet)
	client.subReplies = make(map[uint32]*Subscription)
	client.quenchReplies = make(map[uint32]*Quench)
	// Async Events (Disconn, ECONN, DropWarn, Protocol, ConfConn etc)
	client.Events = make(chan Packet)
	client.confConn = make(chan bool)
	return client
}

// Set the client's log func (rather than use the default)
func (client *Client) SetLogFunc(logger func(io.Writer, string, ...interface{}) (int, error)) {
	client.elog.SetLogFunc(logger)
}

func (client *Client) LogFunc() func(io.Writer, string, ...interface{}) (int, error) {
	return client.elog.LogFunc()
}

// Call client's log func
func (client *Client) Logf(level int, format string, a ...interface{}) (int, error) {
	return client.elog.Logf(level, format, a...)
}

// Set the logfile to an open file
func (client *Client) SetLogFile(file *os.File) {
	client.elog.SetLogFile(file)
}

// Set the log level
func (client *Client) SetLogLevel(level int) {
	client.elog.SetLogLevel(level)
}

// Set the log format
func (client *Client) SetLogDateFormat(format int) {
	client.elog.SetLogDateFormat(format)
}

// Connect this client
// Note this is not thread safe and hence not public
// Client's should call Unotify() or Connect()
func (client *Client) open() (err error) {
	// Establish a socket to the server
	protocol, err := URLToProtocol(client.URL)
	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", protocol.Address)
	if err != nil {
		return err
	}
	client.SetState(StateOpen)

	client.reader = conn
	client.writer = conn
	client.closer = conn

	client.wg.Add(2)
	go client.readHandler()
	go client.writeHandler()

	return nil
}

// This closes a client's sockets/endpoints and cleans state
// returning things to where they were following a NewClient()
// with the exception that the subscription list is maintained
// so it can be re-established on re-connection
func (client *Client) close() {
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

// Connect this client
func (client *Client) Connect() (err error) {

	client.mu.Lock()
	// log.Printf("connect:%s, %d", client.Endpoint, client.State())

	switch client.State() {
	case StateClosed:
		if err = client.open(); err != nil {
			client.mu.Unlock()
			return err
		}
	case StateOpen:
		// It's legal to call Unotify() and then Connect()
	default:
		client.mu.Unlock()
		return LocalError(ErrorsClientIsConnected)
	}

	client.SetState(StateConnecting)
	pkt := new(ConnRequest)
	pkt.XID = XID()
	client.connXID = pkt.XID
	pkt.VersionMajor = ProtocolVersionMajor()
	pkt.VersionMinor = ProtocolVersionMinor()
	pkt.Options = client.Options
	pkt.KeysNfn = client.KeysNfn
	pkt.KeysSub = client.KeysSub

	client.mu.Unlock()

	writeBuf := new(bytes.Buffer)
	pkt.Encode(writeBuf)
	client.writeChannel <- writeBuf

	// Wait for the reply
	select {
	case reply := <-client.connReplies:
		switch reply.(type) {
		case *ConnReply:
			connReply := reply.(*ConnReply)
			// Check XID matches
			if connReply.XID != pkt.XID {
				err = LocalError(ErrorsMismatchedXIDs, pkt.XID, connReply.XID)
			} else {
				// FIXME: Options check/save?
				client.SetState(StateConnected)
			}
		case *Nack:
			client.SetState(StateClosed)
			err = NackError(*reply.(*Nack))
		default:
			client.SetState(StateClosed)
			err = LocalError(ErrorsBadPacket)
		}
	case <-time.After(ConnectTimeout):
		err = LocalError(ErrorsTimeout)
	}

	return err
}

// Disonnect this client from it's endpoint
func (client *Client) Disconnect() (err error) {

	if client.State() != StateConnected {
		return LocalError(ErrorsClientNotConnected)
	}
	client.SetState(StateDisconnecting)

	// FIXME: in a generous world we might unsubscribe, unquench etc
	pkt := new(DisconnRequest)
	pkt.XID = XID()
	client.disconnXID = pkt.XID

	writeBuf := new(bytes.Buffer)
	pkt.Encode(writeBuf)
	client.writeChannel <- writeBuf

	// Wait for the reply
	select {
	case reply := <-client.connReplies:
		switch reply.(type) {
		case *DisconnReply:
			disconnReply := reply.(*DisconnReply)
			// Check XID matches
			if disconnReply.XID != pkt.XID {
				err = LocalError(ErrorsMismatchedXIDs, pkt.XID, disconnReply.XID)
			}
			client.close()
			return err
		default:
			// Didn't hear back, let the client deal with that
			err = LocalError(ErrorsBadPacket)
			return err
		}

	case <-time.After(DisconnectTimeout):
		err = LocalError(ErrorsTimeout)
	}

	return err
}

// Test the connection
func (client *Client) TestConn() (err error) {
	if client.State() != StateConnected {
		return LocalError(ErrorsClientNotConnected)
	}

	pkt := new(TestConn)
	writeBuf := new(bytes.Buffer)
	pkt.Encode(writeBuf)
	client.writeChannel <- writeBuf
	select {
	case <-client.confConn:
		return nil
	case <-time.After(TestConnTimeout):
		return LocalError(ErrorsTimeout)
	}

}

// Send a notification
func (client *Client) Notify(nv map[string]interface{}, deliverInsecure bool, keys KeyBlock) (err error) {

	if client.State() != StateConnected {
		return LocalError(ErrorsClientNotConnected)
	}

	pkt := new(NotifyEmit)
	pkt.NameValue = nv
	pkt.Keys = keys
	pkt.DeliverInsecure = deliverInsecure

	writeBuf := new(bytes.Buffer)
	pkt.Encode(writeBuf)
	client.writeChannel <- writeBuf

	return nil
}

// Send a notification
func (client *Client) UNotify(nv map[string]interface{}, deliverInsecure bool, keys KeyBlock) (err error) {

	switch client.State() {
	case StateClosed:
		if err = client.open(); err != nil {
			return err
		}
	case StateOpen:
		// legal
	case StateConnected:
		// legal
	case StateConnecting:
		return LocalError(ErrorsClientConnecting)
	case StateDisconnecting:
		return LocalError(ErrorsClientDisconnecting)
	}

	switch client.State() {
	case StateConnected:
		break
	case StateClosed:
		return LocalError(ErrorsClientNotConnected)
	}

	pkt := new(UNotify)
	pkt.VersionMajor = ProtocolVersionMajor()
	pkt.VersionMinor = ProtocolVersionMinor()
	pkt.NameValue = nv
	pkt.Keys = keys
	pkt.DeliverInsecure = deliverInsecure

	writeBuf := new(bytes.Buffer)
	pkt.Encode(writeBuf)
	client.writeChannel <- writeBuf

	return nil
}

// Subscribe this client to the subscription
func (client *Client) Subscribe(sub *Subscription) (err error) {

	if client.State() != StateConnected {
		return LocalError(ErrorsClientNotConnected)
	}

	pkt := new(SubAddRequest)
	pkt.Expression = sub.Expression
	pkt.AcceptInsecure = sub.AcceptInsecure
	pkt.Keys = sub.Keys

	sub.events = make(chan Packet)

	writeBuf := new(bytes.Buffer)
	xID := pkt.Encode(writeBuf)

	// Map the XID back to this request along with the notifications
	client.mu.Lock()
	client.subReplies[xID] = sub
	client.mu.Unlock()

	client.writeChannel <- writeBuf

	// Wait for the reply
	select {
	case reply := <-sub.events:
		switch reply.(type) {
		case *SubReply:
			subReply := reply.(*SubReply)
			// Track the subscription id
			sub.subID = subReply.SubID
			client.mu.Lock()
			client.subscriptions[sub.subID] = sub
			client.mu.Unlock()
		case *Nack:
			err = NackError(*reply.(*Nack))
		default:
			err = LocalError(ErrorsBadPacket)
		}

	case <-time.After(SubscriptionTimeout):
		err = LocalError(ErrorsTimeout)
	}

	client.mu.Lock()
	delete(client.subReplies, xID)
	client.mu.Unlock()

	return err
}

// Modify a subscription
// If the expression is empty ("") it will remain unchanged
// Similarly the keysets to add and delete may be empty. It is not an
// error if the added keys already exist or to delete keys that do not
// already exist
func (client *Client) SubscriptionModify(sub *Subscription, expr string, acceptInsecure bool, AddKeys KeyBlock, DelKeys KeyBlock) (err error) {

	if client.State() != StateConnected {
		return LocalError(ErrorsClientNotConnected)
	}

	pkt := new(SubModRequest)
	pkt.SubID = sub.subID
	pkt.Expression = expr
	pkt.AcceptInsecure = acceptInsecure
	pkt.AddKeys = AddKeys
	pkt.DelKeys = DelKeys

	writeBuf := new(bytes.Buffer)
	xID := pkt.Encode(writeBuf)

	// Map the XID back to this request along with the notifications
	client.mu.Lock()
	client.subReplies[xID] = sub
	client.mu.Unlock()

	client.writeChannel <- writeBuf

	// Wait for the reply
	select {
	case reply := <-sub.events:
		switch reply.(type) {
		case *SubReply:
			subReply := reply.(*SubReply)
			// Check the subscription id
			if sub.subID != subReply.SubID {
				client.elog.Logf(elog.LogLevelError, "FIXME: Protocol violation (%v)", reply)
			}

			// Update the local subscription details
			if len(expr) > 0 {
				sub.Expression = expr
			}
			sub.AcceptInsecure = acceptInsecure
			sub.addKeys(AddKeys)
			sub.delKeys(DelKeys)
		case *Nack:
			err = NackError(*reply.(*Nack))
		default:
			err = LocalError(ErrorsBadPacket)
		}

	case <-time.After(SubscriptionTimeout):
		err = LocalError(ErrorsTimeout)
	}

	client.mu.Lock()
	delete(client.subReplies, xID)
	client.mu.Unlock()

	return err
}

// Delete a subscription
func (client *Client) SubscriptionDelete(sub *Subscription) (err error) {

	if client.State() != StateConnected {
		return LocalError(ErrorsClientNotConnected)
	}

	pkt := new(SubDelRequest)
	pkt.SubID = sub.subID

	writeBuf := new(bytes.Buffer)
	xID := pkt.Encode(writeBuf)

	// Map the XID back to this request along with the notifications
	client.mu.Lock()
	client.subReplies[xID] = sub
	client.mu.Unlock()

	client.writeChannel <- writeBuf

	// Wait for the reply
	select {
	case reply := <-sub.events:
		switch reply.(type) {
		case *SubReply:
			subReply := reply.(*SubReply)
			// Check the subscription id
			if sub.subID != subReply.SubID {
				client.elog.Logf(elog.LogLevelError, "FIXME: Protocol violation (%v)", reply)
			}
			// Delete the local subscription details
			client.mu.Lock()
			delete(client.subscriptions, sub.subID)
			client.mu.Unlock()
		case *Nack:
			err = NackError(*reply.(*Nack))
		default:
			err = LocalError(ErrorsBadPacket)
		}

	case <-time.After(SubscriptionTimeout):
		err = LocalError(ErrorsTimeout)
	}

	client.mu.Lock()
	delete(client.subReplies, xID)
	client.mu.Unlock()

	return err
}

// Subscribe this client to the subscription
func (client *Client) Quench(quench *Quench) (err error) {

	if client.State() != StateConnected {
		return LocalError(ErrorsClientNotConnected)
	}

	pkt := new(QuenchAddRequest)
	pkt.Names = quench.Names
	pkt.DeliverInsecure = quench.DeliverInsecure
	pkt.Keys = quench.Keys

	quench.events = make(chan Packet)

	writeBuf := new(bytes.Buffer)
	xID := pkt.Encode(writeBuf)

	// Map the XID back to this request along with the notifications
	client.mu.Lock()
	client.quenchReplies[xID] = quench
	client.mu.Unlock()

	client.writeChannel <- writeBuf

	// Wait for the reply
	select {
	case reply := <-quench.events:
		switch reply.(type) {
		case *QuenchReply:
			quenchReply := reply.(*QuenchReply)
			// Track the quench id
			quench.quenchID = quenchReply.QuenchID
			client.mu.Lock()
			client.quenches[quench.quenchID] = quench
			client.mu.Unlock()
		case *Nack:
			err = NackError(*reply.(*Nack))
		default:
			err = LocalError(ErrorsBadPacket)
		}

	case <-time.After(QuenchTimeout):
		err = LocalError(ErrorsTimeout)
	}

	client.mu.Lock()
	delete(client.quenchReplies, xID)
	client.mu.Unlock()

	return err
}

// Modify a Quench
func (client *Client) QuenchModify(quench *Quench, addNames map[string]bool, delNames map[string]bool, deliverInsecure bool, addKeys KeyBlock, delKeys KeyBlock) (err error) {

	if client.State() != StateConnected {
		return LocalError(ErrorsClientNotConnected)
	}

	pkt := new(QuenchModRequest)
	pkt.QuenchID = quench.quenchID
	pkt.AddNames = addNames
	pkt.DelNames = delNames
	pkt.DeliverInsecure = deliverInsecure
	pkt.AddKeys = addKeys
	pkt.DelKeys = delKeys

	writeBuf := new(bytes.Buffer)
	xID := pkt.Encode(writeBuf)

	// Map the XID back to this request along with the notifications
	client.mu.Lock()
	client.quenchReplies[xID] = quench
	client.mu.Unlock()

	client.writeChannel <- writeBuf

	// Wait for the reply
	select {
	case reply := <-quench.events:
		switch reply.(type) {
		case *QuenchReply:
			quenchReply := reply.(*QuenchReply)
			// Check the quench id
			if quench.quenchID != quenchReply.QuenchID {
				client.elog.Logf(elog.LogLevelError, "FIXME: Protocol violation (%v)", reply)
			}

			quench.DeliverInsecure = deliverInsecure
			quench.addKeys(addKeys)
			quench.delKeys(delKeys)
			for name, _ := range addNames {
				quench.Names[name] = true
			}
			for name, _ := range delNames {
				delete(quench.Names, name)
			}

		case *Nack:
			err = NackError(*reply.(*Nack))
		default:
			err = LocalError(ErrorsBadPacket)
		}

	case <-time.After(QuenchTimeout):
		err = LocalError(ErrorsTimeout)
	}

	client.mu.Lock()
	delete(client.quenchReplies, xID)
	client.mu.Unlock()

	return err
}

func (client *Client) QuenchDelete(quench *Quench) (err error) {

	if client.State() != StateConnected {
		return LocalError(ErrorsClientNotConnected)
	}

	pkt := new(QuenchDelRequest)
	pkt.QuenchID = quench.quenchID

	writeBuf := new(bytes.Buffer)
	xID := pkt.Encode(writeBuf)

	// Map the XID back to this request along with the notifications
	client.mu.Lock()
	client.quenchReplies[xID] = quench
	client.mu.Unlock()

	client.writeChannel <- writeBuf

	// Wait for the reply
	select {
	case reply := <-quench.events:
		switch reply.(type) {
		case *QuenchReply:
			quenchReply := reply.(*QuenchReply)
			// Check the quench id
			if quench.quenchID != quenchReply.QuenchID {
				client.elog.Logf(elog.LogLevelError, "FIXME: Protocol violation (%v)", reply)
			}
			// Delete the local quench details
			client.mu.Lock()
			delete(client.quenches, quench.quenchID)
			client.mu.Unlock()

		case *Nack:
			err = NackError(*reply.(*Nack))
		default:
			err = LocalError(ErrorsBadPacket)
		}

	case <-time.After(QuenchTimeout):
		err = LocalError(ErrorsTimeout)
	}

	client.mu.Lock()
	delete(client.quenchReplies, xID)
	client.mu.Unlock()

	return err
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

	for {
		// We reallocate each time as decoding
		// takes slices out of it
		buffer := make([]byte, 2048)

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
		if err = client.handlePacket(buffer); err != nil {
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

	defer client.close()
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
func (client *Client) handlePacket(buffer []byte) (err error) {

	client.elog.Logf(elog.LogLevelDebug3, "handlePacket received %v (%d)", PacketIDString(PacketID(buffer)), client.State())

	// Packets accepted independent of Client's connection state
	switch PacketID(buffer) {
	case PacketReserved:
		return nil
	case PacketNack:
		return client.handleNack(buffer)
	case PacketTestConn:
		return client.handleTestConn(buffer)
	case PacketConfConn:
		return client.handleConfConn(buffer)
	case PacketDisconn:
		return client.handleDisconn(buffer)
	}

	// Packets dependent upon Client's connection state
	switch client.State() {
	case StateConnecting:
		switch PacketID(buffer) {
		case PacketConnReply:
			return client.handleConnReply(buffer)
		default:
			return LocalError(ErrorsProtocolPacketStateNotConnected, PacketIDString(PacketID(buffer)))
		}

	case StateDisconnecting:
		switch PacketID(buffer) {
		case PacketDisconnReply:
			return client.handleDisconnReply(buffer)
		}

	case StateConnected:
		switch PacketID(buffer) {
		case PacketSubReply:
			return client.handleSubReply(buffer)
		case PacketQuenchReply:
			return client.handleQuenchReply(buffer)
		case PacketNotifyDeliver:
			return client.handleNotifyDeliver(buffer)
		case PacketSubAddNotify:
			return client.handleSubAddNotify(buffer)
		case PacketSubModNotify:
			return client.handleSubModNotify(buffer)
		case PacketSubDelNotify:
			return client.handleSubDelNotify(buffer)
		case PacketDropWarn:
			return client.handleDropWarn(buffer)
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
				client.URL = disconn.Args
				client.close()
				if err := client.Connect(); err != nil {
					client.elog.Logf(elog.LogLevelError, "%v", err)
					os.Exit(1)
				}
				client.elog.Logf(elog.LogLevelInfo1, "connected to %s", client.URL)
			} else {
				client.elog.Logf(elog.LogLevelError, "Disconn to nowhere")
				os.Exit(1)
			}
			break

		case DisconnReasonClientConnectionLost:
			client.elog.Logf(elog.LogLevelWarning, "Lost connection to %s, reconnecting", client.URL)
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
func (client *Client) handleConnReply(buffer []byte) (err error) {
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
func (client *Client) handleDisconnReply(buffer []byte) (err error) {
	disconnReply := new(DisconnReply)
	if err = disconnReply.Decode(buffer); err != nil {
		client.ProtocolError(err)
	}
	// Signal the disconnection requestor
	client.connReplies <- disconnReply
	return nil
}

// Handle a Disconn
func (client *Client) handleDisconn(buffer []byte) (err error) {
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
func (client *Client) handleDropWarn(buffer []byte) (err error) {
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
func (client *Client) handleTestConn(buffer []byte) (err error) {
	// Nothing to decode

	// Respond
	confConn := new(ConfConn)
	writeBuf := new(bytes.Buffer)
	confConn.Encode(writeBuf)
	client.writeChannel <- writeBuf

	return nil
}

// Handle a TestConn
func (client *Client) handleConfConn(buffer []byte) (err error) {
	// Nothing to decode

	// Respond if listening
	select {
	case client.confConn <- true:
	default:
	}

	return nil
}

// Handle a Nack
func (client *Client) handleNack(buffer []byte) (err error) {
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
func (client *Client) handleSubReply(buffer []byte) (err error) {
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
func (client *Client) handleQuenchReply(buffer []byte) (err error) {
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
func (client *Client) handleNotifyDeliver(buffer []byte) (err error) {
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
func (client *Client) handleSubAddNotify(buffer []byte) (err error) {
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
func (client *Client) handleSubModNotify(buffer []byte) (err error) {
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
func (client *Client) handleSubDelNotify(buffer []byte) (err error) {
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
