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
	"net"
	"sync"
	"time"
)

// An Elvin router instance
type Router struct {
	Mu          sync.Mutex
	listeners   map[string]net.Listener
	connections Connections // FIXME: lose the global and use this

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
	connections.lock.Lock()
	defer connections.lock.Unlock()
	glog.Infof("We have %d clients:", len(connections.connections))
	for i, conn := range connections.connections {
		glog.Infof("%d: %+v", i, conn)
	}
	return

}

func (router *Router) Failover() {
	connections.lock.Lock()
	defer connections.lock.Unlock()
	disconn := new(elvin.Disconn)
	disconn.Reason = 2 // Redirect
	disconn.Args = router.failoverProtocol.Address
	glog.Infof("Disconn: %+v", disconn)
	connections.lock.Lock()
	for _, conn := range connections.connections {
		buf := bufferPool.Get().(*bytes.Buffer)
		disconn.Encode(buf)
		conn.writeChannel <- buf
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

		var connection Connection
		connection.reader = conn
		connection.writer = conn
		connection.closer = conn
		connection.testConnInterval = router.testConnInterval
		connection.testConnTimeout = router.testConnTimeout

		connection.SetState(StateNew)
		// Some queuing allowed to smooth things out
		connection.writeChannel = make(chan *bytes.Buffer, 4)
		connection.writeTerminate = make(chan int)

		go connection.readHandler()
		go connection.writeHandler()
	}

	return nil
}
