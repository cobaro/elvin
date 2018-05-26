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

// A Connection (e.g. a socket)
type Client struct {
	// Public
	Endpoint string
	Options  map[string]interface{}
	KeysNfn  []Keyset
	KeysSub  []Keyset

	// Private
	reader         io.Reader
	writer         io.Writer
	closer         io.Closer
	state          uint32
	writeChannel   chan *bytes.Buffer
	readTerminate  chan int
	writeTerminate chan int
	mu             sync.Mutex
	subDelivers    map[uint64]*Subscription // map for NotifyDeliver lookups

	// response channels
	connRply    chan *ConnRply
	disconnRply chan *DisconnRply
	subRplys    map[uint32]*Subscription // map SubAdd/Mod/Del/Nack
}

// Type of event a subscription can receive
const (
	subEventNotifyDeliver = iota
	subEventNack
	subEventSubRply
	subEventSubModRply
	subEventSubDelRply
)

// For lack of a union type we pass one of three packet types discriminated by eventType
type SubscriptionEvent struct {
	eventType     int // subEvent*
	notifyDeliver *NotifyDeliver
	nack          *Nack
	subRply       *SubRply
	// subModRply    *SubModRply
	// subDelRply    *SubDelRply
}

// Client Subscription
type Subscription struct {
	Expression     string                      // Subscription Expression
	AcceptInsecure bool                        // Do we accept notifications with no security keys
	Keys           []Keyset                    // Keys for this subscriptions
	Notifications  chan map[string]interface{} // Notifications delivered on this channel
	subId          uint64
	events         chan SubscriptionEvent
}

// FIXME: define and maybe make configurable?
const ConnectTimeout = (10 * time.Second)
const DisconnectTimeout = (10 * time.Second)
const SubscriptionTimeout = (10 * time.Second)

// Connect this client to it's endpoint
func (client *Client) Connect() (err error) {

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

	go client.readHandler()
	go client.writeHandler()

	pkt := new(ConnRqst)
	pkt.Xid = Xid()
	pkt.VersionMajor = 4
	pkt.VersionMinor = 1
	pkt.Options = client.Options
	pkt.KeysNfn = client.KeysNfn
	pkt.KeysSub = client.KeysSub

	writeBuf := new(bytes.Buffer)
	pkt.Encode(writeBuf)
	client.writeChannel <- writeBuf

	// Wait for the reply
	select {
	case connRply := <-client.connRply:

		// FIXME: check it
		log.Printf("We connected (xid=%d)", connRply.Xid)
		client.SetState(StateConnected)
		return nil

	case <-time.After(ConnectTimeout):
		client.SetState(StateClosed)
		return fmt.Errorf("FIXME: timeout")
	}
}

// Disonnect this client to matchfrom it's endpoint
func (client *Client) Disconnect() (err error) {

	if client.State() != StateConnected {
		return fmt.Errorf("client is not connected")
	}
	client.SetState(StateDisconnecting)

	// FIXME: in a generous world we might unsubscribe, unquench etc
	pkt := new(DisconnRqst)
	pkt.Xid = Xid()

	writeBuf := new(bytes.Buffer)
	pkt.Encode(writeBuf)
	client.writeChannel <- writeBuf

	// Wait for the reply
	select {
	case disconnRply := <-client.disconnRply:
		log.Printf("We disconnected (xid=%d)", disconnRply.Xid)
	case <-time.After(ConnectTimeout):
		err = fmt.Errorf("FIXME: timeout")
	}

	// FIXME: clean up subs
	client.SetState(StateClosed)
	client.closer.Close() // We need to disonnect somehow
	return nil
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

// Subscribe
func (client *Client) Subscribe(sub *Subscription) (err error) {

	if client.State() != StateConnected {
		return fmt.Errorf("FIXME: client not connected")
	}

	pkt := new(SubAddRqst)
	pkt.Expression = sub.Expression
	pkt.AcceptInsecure = sub.AcceptInsecure
	pkt.Keys = sub.Keys

	sub.events = make(chan SubscriptionEvent)

	writeBuf := new(bytes.Buffer)
	xid := pkt.Encode(writeBuf)

	// Map the xid back to this request along with the notifications
	client.mu.Lock()
	client.subRplys[xid] = sub
	client.mu.Unlock()

	client.writeChannel <- writeBuf

	// Wait for the reply
	select {
	case subevent := <-sub.events:
		// Track the subscription id
		sub.subId = subevent.subRply.Subid
		client.mu.Lock()
		client.subDelivers[sub.subId] = sub
		client.mu.Unlock()
		// log.Printf("Subscribe got (%v)", subevent)
	case <-time.After(SubscriptionTimeout):
		err = fmt.Errorf("FIXME: timeout")
	}

	return err
}
