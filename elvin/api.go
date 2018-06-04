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
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// A client of an Elvin service, typically used via:
//      client.Connect()
//      client = newClient()
//      client.Subscribe()
//      client.Notify()
//      client.Disonnect()
// See individual methods for details

type Client struct {
	// Public
	Endpoint       string
	Options        map[string]interface{}
	KeysNfn        []Keyset
	KeysSub        []Keyset
	DisconnChannel chan *Disconn // Clients may listen here for disconnects

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

	// Map of all current subscriptions used for mapping NotifyDelivers
	//  and for maintaining subscriptions across reconnection
	subscriptions map[uint64]*Subscription

	// response channels
	connXID        uint32                   // XID of outstanding connrqst
	connReplies    chan Packet              // Channel for Connect() packets
	disconnXID     uint32                   // XID of outstanding disconnrqst
	disconnReplies chan Packet              // Channel for Connect() packets
	subReplies     map[uint32]*Subscription // map SubAdd/Mod/Del/Nack

}

// Types of event a subscription can receive
const (
	subEventNotifyDeliver = iota
	subEventNack
	subEventSubRply
	subEventSubModRply
	subEventSubDelRply
)

// The Subscription type used by clients.
type Subscription struct {
	Expression     string                      // Subscription Expression
	AcceptInsecure bool                        // Do we accept notifications with no security keys
	Keys           []Keyset                    // Keys for this subscriptions
	Notifications  chan map[string]interface{} // Notifications delivered on this channel
	subID          uint64
	events         chan Packet
}

func (sub *Subscription) addKeys(keys []Keyset) {
	// FIXME: implement
	return
}

func (sub *Subscription) delKeys(keys []Keyset) {
	// FIXME: implement
	return
}

// FIXME: define and maybe make configurable?
const ConnectTimeout = (10 * time.Second)
const DisconnectTimeout = (10 * time.Second)
const SubscriptionTimeout = (10 * time.Second)

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
	client.subscriptions = make(map[uint64]*Subscription)
	// Sync Packets
	client.connReplies = make(chan Packet)
	client.disconnReplies = make(chan Packet)
	client.subReplies = make(map[uint32]*Subscription)
	client.DisconnChannel = make(chan *Disconn)

	return client
}

// Connect this client to it's endpoint
func (client *Client) Connect() (err error) {

	client.mu.Lock()
	if client.State() != StateClosed {
		return fmt.Errorf("FIXME: client already connected")
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

	pkt := new(ConnRqst)
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
	case rply := <-client.connReplies:
		switch rply.(type) {
		case *ConnRply:
			connRply := rply.(*ConnRply)
			// Check XID matches
			if connRply.XID != pkt.XID {
				err = fmt.Errorf("Mismatched transaction IDs, expected %d, received %d", pkt.XID, connRply.XID)
			} else {
				// FIXME: Options check/save?
				client.SetState(StateConnected)
			}
			break

		case *Nack:
			nack := rply.(*Nack)
			err = fmt.Errorf(nack.String())
			client.SetState(StateConnected)
			break
		default:
			// FIXME: die
			err = fmt.Errorf("Unexpected packet")
			break
		}
	case <-time.After(ConnectTimeout):
		client.SetState(StateClosed)
		err = fmt.Errorf("FIXME: timeout")
	}

	return err
}

// Disonnect this client from it's endpoint
func (client *Client) Disconnect() (err error) {

	if client.State() != StateConnected {
		return fmt.Errorf("client is not connected")
	}
	client.SetState(StateDisconnecting)

	// FIXME: in a generous world we might unsubscribe, unquench etc
	pkt := new(DisconnRqst)
	pkt.XID = XID()
	client.disconnXID = pkt.XID

	writeBuf := new(bytes.Buffer)
	pkt.Encode(writeBuf)
	client.writeChannel <- writeBuf

	// Wait for the reply
loop:
	select {
	case rply := <-client.disconnReplies:
		switch rply.(type) {
		case *DisconnRply:
			disconnRply := rply.(*DisconnRply)
			// Check XID matches
			if disconnRply.XID != pkt.XID {
				err = fmt.Errorf("Mismatched transaction IDs, expected %d, received %d", pkt.XID, disconnRply.XID)
			}
			client.Close()
			break loop
		default:
			// Didn't hear back, let the client deal with that
			err = fmt.Errorf("Unexpected packet")
			break loop

		}

	case <-time.After(DisconnectTimeout):
		err = fmt.Errorf("FIXME: timeout")
		break
	}

	return err
}

// Send a notification
func (client *Client) Notify(nv map[string]interface{}, deliverInsecure bool, keys []Keyset) (err error) {

	if client.State() != StateConnected {
		return fmt.Errorf("FIXME: client not connected")
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
		return fmt.Errorf("FIXME: client not connected")
	}

	pkt := new(SubAddRqst)
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
	case rply := <-sub.events:
		switch rply.(type) {
		case *SubRply:
			subRply := rply.(*SubRply)
			// Track the subscription id
			sub.subID = subRply.SubID
			client.mu.Lock()
			client.subscriptions[sub.subID] = sub
			client.mu.Unlock()
			break
		case *Nack:
			nack := rply.(*Nack)
			err = fmt.Errorf(nack.String())
			break
		default:
			log.Printf("OOPS (%v)", rply)
		}

	case <-time.After(SubscriptionTimeout):
		err = fmt.Errorf("FIXME: timeout")
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
		return fmt.Errorf("FIXME: client not connected")
	}

	pkt := new(SubModRqst)
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
	case rply := <-sub.events:
		switch rply.(type) {
		case *SubRply:
			subRply := rply.(*SubRply)
			// Check the subscription id
			if sub.subID != subRply.SubID {
				log.Printf("FIXME: Protocol violation (%v)", rply)
			}

			// Update the local subscription details
			if len(expr) > 0 {
				sub.Expression = expr
			}
			sub.AcceptInsecure = acceptInsecure
			sub.addKeys(AddKeys)
			sub.delKeys(DelKeys)

			break
		case *Nack:
			nack := rply.(*Nack)
			err = fmt.Errorf(nack.String())
			break
		default:
			log.Printf("OOPS (%v)", rply)
		}

	case <-time.After(SubscriptionTimeout):
		err = fmt.Errorf("FIXME: timeout")
	}

	client.mu.Lock()
	delete(client.subReplies, xID)
	client.mu.Unlock()

	return err
}

// Delete a subscription
func (client *Client) SubscriptionDelete(sub *Subscription) (err error) {

	if client.State() != StateConnected {
		return fmt.Errorf("FIXME: client not connected")
	}

	pkt := new(SubDelRqst)
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
	case rply := <-sub.events:
		switch rply.(type) {
		case *SubRply:
			subRply := rply.(*SubRply)
			// Check the subscription id
			if sub.subID != subRply.SubID {
				log.Printf("FIXME: Protocol violation (%v)", rply)
			}
			// Delete the local subscription details
			client.mu.Lock()
			delete(client.subscriptions, sub.subID)
			client.mu.Unlock()

			break
		case *Nack:
			nack := rply.(*Nack)
			err = fmt.Errorf(nack.String())
			break
		default:
			log.Printf("OOPS (%v)", rply)
		}

	case <-time.After(SubscriptionTimeout):
		err = fmt.Errorf("FIXME: timeout")
	}

	client.mu.Lock()
	delete(client.subReplies, xID)
	client.mu.Unlock()

	return err
}
