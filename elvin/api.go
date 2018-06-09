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
	"io"
	"log"
	"net"
	"sync"
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
	Endpoint      string                 // Router descriptor
	Options       map[string]interface{} // Router options
	KeysNfn       []Keyset               // Connections keys for outgoing notifications
	KeysSub       []Keyset               // Connections keys for incoming notifications
	Notifications chan Packet            // Clients may listen here for events

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

// A subscription type used by clients.
type Subscription struct {
	Expression     string                      // Subscription Expression
	AcceptInsecure bool                        // Do we accept notifications with no security keys
	Keys           []Keyset                    // Keys for this subscriptions
	Notifications  chan map[string]interface{} // Notifications delivered on this channel

	subID  int64       // private id
	events chan Packet // synchronous replies
}

func (sub *Subscription) addKeys(keys []Keyset) {
	// FIXME: implement
	return
}

func (sub *Subscription) delKeys(keys []Keyset) {
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
	Keys            []Keyset                // Keys for this quench
	Notifications   chan QuenchNotification // Sub{Add|Del|Mod}Notify delivers
	quenchID        int64                   // private id
	events          chan Packet             // synchronous replies
}

func (quench *Quench) addKeys(keys []Keyset) {
	// FIXME: implement
	return
}

func (quench *Quench) delKeys(keys []Keyset) {
	// FIXME: implement
	return
}

// FIXME: define and maybe make configurable?
const ConnectTimeout = (10 * time.Second)
const DisconnectTimeout = (10 * time.Second)
const SubscriptionTimeout = (10 * time.Second)
const QuenchTimeout = (10 * time.Second)
const TestConnTimeout = (10 * time.Second)

// Create a new client.
// Using new(Client) will not result in proper initialization
func NewClient(endpoint string, options map[string]interface{}, keysNfn []Keyset, keysSub []Keyset) (conn *Client) {
	client := new(Client)
	client.Endpoint = endpoint
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
	client.Notifications = make(chan Packet)
	client.confConn = make(chan bool)

	return client
}

// Connect this client to it's endpoint
func (client *Client) Connect() (err error) {

	client.mu.Lock()
	if client.State() != StateClosed {
		return LocalError(ErrorsClientIsConnected)
	}
	client.SetState(StateConnecting)

	// Establish a socket to the server
	conn, err := net.Dial("tcp", client.Endpoint)
	if err != nil {
		return err
	}

	client.reader = conn
	client.writer = conn
	client.closer = conn

	client.wg.Add(2)
	go client.readHandler()
	go client.writeHandler()

	pkt := new(ConnRequest)
	pkt.XID = XID()
	client.connXID = pkt.XID
	pkt.VersionMajor = 4
	pkt.VersionMinor = 1
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
			client.Close()
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
func (client *Client) Notify(nv map[string]interface{}, deliverInsecure bool, keys []Keyset) (err error) {

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
func (client *Client) SubscriptionModify(sub *Subscription, expr string, acceptInsecure bool, AddKeys []Keyset, DelKeys []Keyset) (err error) {

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
				log.Printf("FIXME: Protocol violation (%v)", reply)
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
				log.Printf("FIXME: Protocol violation (%v)", reply)
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
func (client *Client) QuenchModify(quench *Quench, addNames map[string]bool, delNames map[string]bool, deliverInsecure bool, addKeys []Keyset, delKeys []Keyset) (err error) {

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
				log.Printf("FIXME: Protocol violation (%v)", reply)
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
				log.Printf("FIXME: Protocol violation (%v)", reply)
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
