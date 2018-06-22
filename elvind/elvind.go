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

package main

import (
	"bytes"
	"fmt"
	"github.com/cobaro/elvin/elvin"
	"github.com/golang/glog"
	"math/rand"
	"net"
	"sync"
	"time"
)

// An Elvin router instance
type Router struct {
	Mu        sync.Mutex
	listeners map[string]net.Listener
	clients   Clients // FIXME: lose the global and use this

	// Configurable
	protocols        map[string]Protocol
	failoverProtocol Protocol
	testConnInterval time.Duration
	testConnTimeout  time.Duration
	maxConnections   int
	doFailover       bool
	running          bool
}

// Set the maximum allowed number of clients
func (router *Router) SetMaxConnections(max int) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	router.maxConnections = max
}

// Get the maximum allowed number of clients
func (router *Router) MaxConnections() int {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	return router.maxConnections
}

// Set the interval for TestConn (0 to disable)
func (router *Router) SetTestConnInterval(interval time.Duration) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	router.testConnInterval = interval
}

// Get the current TestConn interval
func (router *Router) TestConnInterval() time.Duration {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	return router.testConnInterval
}

// Set the duration of a TestConn response timeout
func (router *Router) SetTestConnTimeout(timeout time.Duration) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	router.testConnTimeout = timeout
}

// Get the current TestConn interval
func (router *Router) TestConnTimeout() time.Duration {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	return router.testConnTimeout
}

// Set the maximum allowed number of clients
func (router *Router) SetDoFailover(failover bool) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	router.doFailover = failover
}

// Get the maximum allowed number of clients
func (router *Router) DoFailover() bool {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	return router.doFailover
}

// Add a protocol
func (router *Router) AddProtocol(name string, protocol Protocol) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	if router.protocols == nil {
		router.protocols = make(map[string]Protocol)
	}
	router.protocols[name] = protocol

	if router.running {
		go router.Listener(name, protocol)
	}
}

// Delete a protocol
func (router *Router) DeleteProtocol(name string) (err error) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	if _, ok := router.protocols[name]; ok {
		delete(router.protocols, name)
		if listener, ok := router.listeners[name]; ok {
			listener.Close() // Tell it to exit
			delete(router.listeners, name)
		}
		return nil
	} else {
		return fmt.Errorf("No such protocol '%s'", name)
	}
}

// Add a failover host
func (router *Router) SetFailoverProtocol(protocol Protocol) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	router.failoverProtocol = protocol
}

// Delete a failover host
func (router *Router) FailoverProtocol() (protocol Protocol) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	return router.failoverProtocol

}

func (router *Router) ReportClients() {
	clients.mu.Lock()
	defer clients.mu.Unlock()
	glog.Infof("We have %d clients:", len(clients.clients))
	for i, c := range clients.clients {
		glog.Infof("%d: %+v", i, c)
	}
	return

}

func (router *Router) Failover() {
	clients.mu.Lock()
	defer clients.mu.Unlock()
	disconn := new(elvin.Disconn)
	disconn.Reason = 2 // Redirect
	disconn.Args = router.failoverProtocol.Address
	glog.Infof("Disconn: %+v", disconn)
	for _, c := range clients.clients {
		buf := bufferPool.Get().(*bytes.Buffer)
		disconn.Encode(buf)
		c.writeChannel <- buf
	}
	return
}

func (router *Router) Start() (err error) {
	router.Mu.Lock()
	defer router.Mu.Unlock()

	// Check Protocols
	for name, protocol := range router.protocols {
		switch protocol.Network {
		case "tcp":
		default:
			glog.Errorf("network protocol %s is currently unsupported", protocol.Network)
			delete(router.protocols, name)
		}

		switch protocol.Marshal {
		case "xdr":
		default:
			glog.Errorf("marshal protocol %s is currently unsupport", protocol.Marshal)
			delete(router.protocols, name)
		}
	}

	// We're away
	router.running = true

	// Set up listeners
	router.listeners = make(map[string]net.Listener)
	for name, protocol := range router.protocols {
		glog.Infof("listener: %v", name)
		go router.Listener(name, protocol)
	}

	return nil
}

func (router *Router) Stop() (err error) {
	router.Mu.Lock()
	defer router.Mu.Unlock()

	// We're stopping
	router.running = false

	// Shut down the listeners
	for _, listener := range router.listeners {
		listener.Close()
	}

	return nil
}

func (router *Router) Listener(name string, protocol Protocol) (err error) {

	if glog.V(1) {
		glog.Infof("Start listening on %s %s %s", protocol.Network, protocol.Marshal, protocol.Address)
		defer glog.Infof("Stop listening on %s %s %s", protocol.Network, protocol.Marshal, protocol.Address)
	}

	listener, err := net.Listen(protocol.Network, protocol.Address)
	if err != nil {
		return fmt.Errorf("FIXME: Listen failed: %v", err)
	}
	router.Mu.Lock()
	router.listeners[name] = listener
	router.Mu.Unlock()

	var conn net.Conn
	for {
		if conn, err = listener.Accept(); err != nil {
			return nil // Happens when we're closed so simply bail
		}

		var client Client
		clients.Add(&client) // track it

		client.reader = conn
		client.writer = conn
		client.closer = conn
		client.testConnInterval = router.testConnInterval
		client.testConnTimeout = router.testConnTimeout

		client.SetState(StateNew)
		// Some queuing allowed to smooth things out
		client.writeChannel = make(chan *bytes.Buffer, 4)
		client.writeTerminate = make(chan int)

		go client.readHandler()
		go client.writeHandler()
	}
}

// Global clients
var clients Clients

func init() {
	clients.clients = make(map[int32]*Client)
	clients.channels.remove = make(chan int32)
	clients.channels.notify = make(chan *elvin.NotifyEmit)
	clients.channels.subAdd = make(chan *Subscription)
	clients.channels.subMod = make(chan *Subscription)
	clients.channels.subDel = make(chan *Subscription)
	clients.channels.quenchAdd = make(chan *Quench)
	clients.channels.quenchMod = make(chan *Quench)
	clients.channels.quenchDel = make(chan *Quench)

	// Start remove goroutine for client cleanup
	go Remove(&clients)

	// Start goroutine for notification eval
	go Notify(&clients)

	// Start goroutine for subscription changes
	go Subscriptions(&clients)

	// Start goroutine for quench changes
	go Quenches(&clients)
}

// Clients is a set of client
type Clients struct {
	mu       sync.Mutex
	clients  map[int32]*Client // initialized in init()
	channels ClientChannels    // For sending notifications, subs, quenches, delete etc to engine
}

// Operations from a client handled via channel to clients
type ClientChannels struct {
	remove    chan int32             // Client removal channel
	notify    chan *elvin.NotifyEmit // Notifications
	subAdd    chan *Subscription     // Subscription Add
	subMod    chan *Subscription     // Subscription Mod
	subDel    chan *Subscription     // Subscription Del
	quenchAdd chan *Quench           // Quench Add
	quenchMod chan *Quench           // Quench Mod
	quenchDel chan *Quench           // Quench Del
}

// Create a unique 32 bit unsigned integer id
func (c *Clients) Add(conn *Client) {
	clients.mu.Lock()
	defer clients.mu.Unlock()
	var id int32 = rand.Int31()
	for {
		_, err := c.clients[id]
		if !err {
			break
		}
		id++
	}

	if glog.V(4) {
		glog.Infof("New client %d", id)
	}
	conn.id = id
	c.clients[id] = conn
	conn.channels = c.channels
	return
}

// Remove will purge a client from the set of clients (run as goroutine)
func Remove(c *Clients) {
	for {
		id := <-c.channels.remove
		if glog.V(4) {
			glog.Infof("Remove client %d", id)
		}

		clients.mu.Lock()
		delete(clients.clients, id)
		clients.mu.Unlock()
		// FIXME: Clean up the subscriptions and quenches
	}
}

// Notify is our queue of incoming messages (run as goroutine)
func Notify(c *Clients) {
	for {
		nfn := <-c.channels.notify
		if glog.V(5) {
			glog.Infof("Remove notification %+v", nfn)
		}

		// FIXME: eval
		// As a dummy for now we're going to send every message we see
		// to every subscription as if all evaluate to true.
		deliver := new(elvin.NotifyDeliver)
		deliver.NameValue = nfn.NameValue

		// Grab a copy of the current client list
		// For now we don't care if one updates mid stream
		c.mu.Lock()
		clients := c.clients
		c.mu.Unlock()

		for connid, client := range clients {
			if len(client.subs) > 0 {
				deliver.Insecure = make([]int64, len(client.subs))
				i := 0
				for id, _ := range client.subs {
					deliver.Insecure[i] = int64(connid)<<32 | int64(id)
					i++
				}
				buf := bufferPool.Get().(*bytes.Buffer)
				deliver.Encode(buf)
				client.writeChannel <- buf
			}
		}
	}
}

// FIXME: implement
// Subscriptions deals with changes to all of our client's subscriptions (run as goroutine)
func Subscriptions(c *Clients) {
	for {
		var sub *Subscription
		select {
		case sub = <-c.channels.subAdd:
			if glog.V(4) {
				glog.Infof("SubAdd")
			}
		case sub = <-c.channels.subMod:
			if glog.V(4) {
				glog.Infof("SubMod")
			}
		case sub = <-c.channels.subDel:
			if glog.V(4) {
				glog.Infof("SubDel")
			}
		}

		if sub.SubID == 0 {
			glog.Infof("FIXME: Use sub to keep compiler happy")
		}

	}
}

// FIXME: implement
// Quenches deals with quencehs to all of our client's quenches (run as goroutine)
func Quenches(c *Clients) {
	for {
		var quench *Quench
		select {
		case quench = <-c.channels.quenchAdd:
			if glog.V(4) {
				glog.Infof("QuenchAdd")
			}
		case quench = <-c.channels.quenchMod:
			if glog.V(4) {
				glog.Infof("QuenchMod")
			}
		case quench = <-c.channels.quenchDel:
			if glog.V(4) {
				glog.Infof("QuenchDel")
			}
		}
		if quench.QuenchID == 0 {
			glog.Infof("FIXME: Use quench to keep compiler happy")
		}
	}

}
