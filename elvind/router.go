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
	"github.com/cobaro/elvin/elog"
	"github.com/cobaro/elvin/elvin"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

// An Elvin router instance
type Router struct {
	Mu        sync.Mutex
	listeners map[string]net.Listener
	clients   map[int32]*Client // Required to be initialized by Init()
	channels  ClientChannels    // For notifications, subs, quenches, delete etc to engine
	elog      elog.Elog

	// Configurable
	protocols        map[string]*elvin.Protocol
	failoverProtocol *elvin.Protocol
	testConnInterval time.Duration
	testConnTimeout  time.Duration
	maxConnections   int
	doFailover       bool
	logLevel         int
	logFormat        int
	logPath          string // FIXME: implement

	// state
	initialized bool
	running     bool
}

// Operations from a client handled via channel to clients
type ClientChannels struct {
	remove    chan int32         // Client removal channel
	notify    chan Notification  // Notifications
	subAdd    chan *Subscription // Subscription Add
	subMod    chan *Subscription // Subscription Mod
	subDel    chan *Subscription // Subscription Del
	quenchAdd chan *Quench       // Quench Add
	quenchMod chan *Quench       // Quench Mod
	quenchDel chan *Quench       // Quench Del
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

// Set the log level
func (router *Router) SetLogLevel(level int) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	router.elog.SetLogLevel(level)
}

// Get the log level
func (router *Router) LogLevel() int {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	return router.elog.LogLevel()
}

// Set the log format
func (router *Router) SetLogDateFormat(format int) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	router.elog.SetLogDateFormat(format)
}

// Get the log format
func (router *Router) LogDateFormat() int {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	return router.LogDateFormat()
}

// Set the log file
func (router *Router) SetLogFile(file *os.File) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	router.elog.SetLogFile(file)
}

// Get the log file
func (router *Router) LogFile() (file os.File) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	return router.LogFile()
}

// Add a protocol
func (router *Router) AddProtocol(name string, protocol *elvin.Protocol) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	if router.protocols == nil {
		router.protocols = make(map[string]*elvin.Protocol)
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
func (router *Router) SetFailoverProtocol(protocol *elvin.Protocol) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	router.failoverProtocol = protocol
}

// Delete a failover host
func (router *Router) FailoverProtocol() (protocol *elvin.Protocol) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	return router.failoverProtocol

}

// Log info about our clients
func (router *Router) LogClients() {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	router.elog.Logf(elog.LogLevelInfo1, "We have %d clients:", len(router.clients))
	for i, c := range router.clients {
		router.elog.Logf(elog.LogLevelInfo1, "%d: %+v", i, c)
	}
	return

}

// Tell our clients to Failover to the configured failover host
// FIXME: Could take url and shold return an err if no url specified
//        or configured
// FIXME: Should we have an option to stop the listeners to avoid new connections?
//        Or perhaps a state that means we bounce new clients immediately?
func (router *Router) Failover() {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	disconn := new(elvin.Disconn)
	disconn.Reason = elvin.DisconnReasonRouterRedirect
	disconn.Args = router.failoverProtocol.Address
	router.elog.Logf(elog.LogLevelDebug2, "Disconn: %+v", disconn)
	for _, c := range router.clients {
		buf := bufferPool.Get().(*bytes.Buffer)
		disconn.Encode(buf)
		c.writeChannel <- buf
	}
	return
}

// Router initialization
func (router *Router) Init() {
	router.clients = make(map[int32]*Client)
	router.channels.remove = make(chan int32)
	router.channels.notify = make(chan Notification)
	router.channels.subAdd = make(chan *Subscription)
	router.channels.subMod = make(chan *Subscription)
	router.channels.subDel = make(chan *Subscription)
	router.channels.quenchAdd = make(chan *Quench)
	router.channels.quenchMod = make(chan *Quench)
	router.channels.quenchDel = make(chan *Quench)
	router.initialized = true

	// Start remove goroutine for client cleanup
	go router.RemoveClient()

	// Start goroutine for notification eval
	go router.Notify()

	// Start goroutine for subscription changes
	go router.Subscriptions()

	// Start goroutine for quench changes
	go router.Quenches()
}

// Start a router with current configurartion
func (router *Router) Start() (err error) {
	router.Mu.Lock()
	defer router.Mu.Unlock()

	if !router.initialized {
		router.Mu.Unlock()
		router.Init()
		router.Mu.Lock()
	}

	// Check Protocols
	for name, protocol := range router.protocols {
		switch protocol.Network {
		case "tcp":
		default:
			router.elog.Logf(elog.LogLevelWarning, "network protocol %s is currently unsupported", protocol.Network)
			delete(router.protocols, name)
		}

		switch protocol.Marshal {
		case "xdr":
		default:
			router.elog.Logf(elog.LogLevelWarning, "marshal protocol %s is currently unsupported", protocol.Marshal)
			delete(router.protocols, name)
		}
	}

	// We're away
	router.running = true

	// Set up listeners
	router.listeners = make(map[string]net.Listener)
	for name, protocol := range router.protocols {
		go router.Listener(name, protocol)
	}

	return nil
}

// Stop a router, taking us back to a running but clean state
func (router *Router) Stop() (err error) {
	router.Mu.Lock()
	defer router.Mu.Unlock()

	// We're stopping
	router.running = false

	// Shut down the listeners
	router.elog.Logf(elog.LogLevelInfo2, "Closing listeners")
	for name, listener := range router.listeners {
		listener.Close()
		delete(router.listeners, name)

	}

	// Shut down the clients
	router.elog.Logf(elog.LogLevelInfo2, "Closing clients")
	disconn := new(elvin.Disconn)
	disconn.Reason = elvin.DisconnReasonRouterShuttingDown
	router.elog.Logf(elog.LogLevelDebug2, "Disconn: %+v", disconn)
	for _, c := range router.clients {
		buf := bufferPool.Get().(*bytes.Buffer)
		disconn.Encode(buf)
		c.writeChannel <- buf
	}

	// FIXME: Shut down our goroutines
	router.elog.Logf(elog.LogLevelInfo2, "Stopped")
	return nil
}

// Shutdown
func (router *Router) Shutdown() (err error) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	if router.running {
		router.Mu.Unlock()
		router.Stop()
		router.Mu.Lock()
	}

	// FIXME: Shut down our goroutines
	os.Exit(0)
	return nil
}

func (router *Router) Listener(name string, protocol *elvin.Protocol) (err error) {

	router.elog.Logf(elog.LogLevelInfo1, "Start listening on %s %s %s", protocol.Network, protocol.Marshal, protocol.Address)
	defer router.elog.Logf(elog.LogLevelInfo1, "Stop listening on %s %s %s", protocol.Network, protocol.Marshal, protocol.Address)

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

		client.elog = router.elog
		client.reader = conn
		client.writer = conn
		client.closer = conn
		client.testConnInterval = router.testConnInterval
		client.testConnTimeout = router.testConnTimeout

		client.SetState(StateNew)
		// Some queuing allowed to smooth things out
		client.writeChannel = make(chan *bytes.Buffer, 4)
		client.writeTerminate = make(chan int)

		router.AddClient(&client) // track it
		go client.readHandler()
		go client.writeHandler()
	}
}

// Create a unique 32 bit unsigned integer id
func (router *Router) AddClient(conn *Client) {
	router.Mu.Lock()
	defer router.Mu.Unlock()
	var id int32 = rand.Int31()
	for {
		_, err := router.clients[id]
		if !err {
			break
		}
		id++
	}

	router.elog.Logf(elog.LogLevelDebug1, "New client %d", id)
	conn.id = id
	router.clients[id] = conn
	conn.channels = router.channels
	return
}

// Remove will purge a client from the set of clients (run as goroutine)
func (router *Router) RemoveClient() {
	for {
		id := <-router.channels.remove
		router.elog.Logf(elog.LogLevelDebug1, "Remove client %d", id)

		router.Mu.Lock()
		delete(router.clients, id)
		router.Mu.Unlock()
		// FIXME: Clean up the subscriptions and quenches
	}
}

// Notify is our queue of incoming messages (run as goroutine)
func (router *Router) Notify() {
	for {
		nfn := <-router.channels.notify
		router.elog.Logf(elog.LogLevelDebug3, "notification %+v", nfn)

		// FIXME: eval
		// As a dummy for now we're going to send every message we see
		// to every subscription as if all evaluate to true.
		deliver := new(elvin.NotifyDeliver)
		deliver.NameValue = nfn.NameValue

		// Grab a copy of the current client list
		// For now we don't care if one updates mid stream
		router.Mu.Lock()
		clients := router.clients
		router.Mu.Unlock()

		for connid, client := range clients {
			if len(client.subs) > 0 {
				deliver.Insecure = make([]int64, len(client.subs))
				i := 0
				for id, sub := range client.subs {
					// Security check first
					PrimeProducer(nfn.Keys)

					if SecurityMatches(nfn, *sub, nfn.ClientKeys, client.keysSub) {
						router.elog.Logf(elog.LogLevelDebug1, "SecurityMatches true")
						deliver.Insecure[i] = int64(connid)<<32 | int64(id)
						i++
					} else {
						router.elog.Logf(elog.LogLevelDebug1, "SecurityMatches false")
					}

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
func (router *Router) Subscriptions() {
	for {
		var sub *Subscription
		select {
		case sub = <-router.channels.subAdd:
			router.elog.Logf(elog.LogLevelInfo2, "SubAdd")
		case sub = <-router.channels.subMod:
			router.elog.Logf(elog.LogLevelInfo2, "SubMod")
		case sub = <-router.channels.subDel:
			router.elog.Logf(elog.LogLevelInfo2, "SubDel")
		}

		if sub.SubID == 0 {
			router.elog.Logf(elog.LogLevelError, "FIXME: Use sub to keep compiler happy")
		}

	}
}

// FIXME: implement
// Quenches deals with quencehs to all of our client's quenches (run as goroutine)
func (router *Router) Quenches() {
	for {
		var quench *Quench
		select {
		case quench = <-router.channels.quenchAdd:
			router.elog.Logf(elog.LogLevelInfo2, "QuenchAdd")
		case quench = <-router.channels.quenchMod:
			router.elog.Logf(elog.LogLevelInfo2, "QuenchMod")
		case quench = <-router.channels.quenchDel:
			router.elog.Logf(elog.LogLevelInfo2, "QuenchDel")
		}
		if quench.QuenchID == 0 {
			router.elog.Logf(elog.LogLevelError, "FIXME: Use quench to keep compiler happy")
		}
	}

}
