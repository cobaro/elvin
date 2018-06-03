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
	"flag"
	"github.com/cobaro/elvin/elvin"
	"github.com/golang/glog"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Argument parsing
	configFile := flag.String("config", "elvind.json", "JSON config file path")
	flag.Parse()

	// Load config
	config, err := LoadConfig(*configFile)
	if err != nil {
		glog.Fatal("config load failed:", err)
	}

	if glog.V(2) {
		glog.Infof("Config: %+v", *config)
	}

	// Check Protocols and set up listeners
	for _, protocol := range config.Protocols {
		switch protocol.Network {
		case "tcp":
			break
		case "udp":
		case "ssl":
			glog.Warningf("network protocol %s is currently unsupported", protocol.Network)
			continue
		default:
			glog.Warningf("network protocol %s is unknown", protocol.Network)
			continue
		}

		switch protocol.Marshal {
		case "xdr":
			break
		case "protobuf":
			glog.Warningf("marshal protocol %s is currently unsupported", protocol.Marshal)
			continue
		default:
			glog.Warningf("marshal protocol %s is unknown", protocol.Marshal)
			continue
		}
		// TODO: track listeners for shutdown
		go Listener(protocol)
	}

	// Set up sigint handling and wait for one
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	// FIXME: These will go away as things develop but for now
	// they're convenient

	// State reporting on SIGUSR1 (testing/debugging)
	signal.Notify(ch, syscall.SIGUSR1)

	// Failover on SIGUSR2 (testing)
	if config.DoFailover && len(config.FailoverHosts) > 0 {
		// FIXME: elvin://
		signal.Notify(ch, syscall.SIGUSR2)
	}

	for {
		sig := <-ch
		switch sig {
		case os.Interrupt:
			glog.Info("Exiting on ", sig)
			glog.Flush()
			os.Exit(0)
		case syscall.SIGUSR1:
			connections.lock.Lock()
			glog.Infof("Client list:")
			for i, conn := range connections.connections {
				glog.Infof("%d: %+v", i, conn)
			}
			connections.lock.Unlock()
			break
		case syscall.SIGUSR2:
			disconn := new(elvin.Disconn)
			disconn.Reason = 2 // Redirect
			disconn.Args = config.FailoverHosts[0].Address
			glog.Infof("Disconn: %+v", disconn)
			connections.lock.Lock()
			for _, conn := range connections.connections {
				buf := bufferPool.Get().(*bytes.Buffer)
				disconn.Encode(buf)
				conn.writeChannel <- buf
				// delete(connections.connections, i)
				// conn.writeTerminate <- 1
				// conn.closer.Close()
			}
			connections.lock.Unlock()
			break
		}
	}

	return

}

func Listener(protocol Protocol) {

	glog.Infof("Listening on %s %s %s", protocol.Network, protocol.Marshal, protocol.Address)

	ln, err := net.Listen(protocol.Network, protocol.Address)
	if err != nil {
		glog.Fatal("Listen failed:", err)
	}

	for {
		c, err := ln.Accept()
		if err != nil {
			glog.Fatal("Accept failed:", err)
		}

		var conn Connection
		conn.reader = c
		conn.writer = c
		conn.closer = c
		conn.state = StateNew
		conn.writeChannel = make(chan *bytes.Buffer, 4) // Some queuing allowed to smooth things out
		conn.readTerminate = make(chan int)
		conn.writeTerminate = make(chan int)

		go conn.readHandler()
		go conn.writeHandler()
	}
}
