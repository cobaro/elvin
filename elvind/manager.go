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
	"flag"
	"github.com/golang/glog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Manager struct {
	config    *Configuration
	router    Router
	protocols map[string]Protocol
	failover  Protocol
}

func main() {
	var manager Manager
	var err error

	// Argument parsing
	configFile := flag.String("config", "elvind.json", "JSON config file path")
	flag.Parse()
	if manager.config, err = LoadConfig(*configFile); err != nil {
		glog.Fatal("config load failed:", err)
	}
	if glog.V(3) {
		glog.Infof("manager.config: %+v", *manager.config)
	}

	glog.Info("Start configuration")
	manager.router.SetMaxConnections(manager.config.MaxConnections)
	manager.router.SetDoFailover(manager.config.DoFailover)
	manager.router.SetTestConnInterval(time.Duration(manager.config.TestConnInterval) * time.Second)
	manager.router.SetTestConnTimeout(time.Duration(manager.config.TestConnTimeout) * time.Second)
	manager.protocols = make(map[string]Protocol)
	for _, protocol := range manager.config.Protocols {
		manager.protocols[protocol.Address] = protocol
		manager.router.AddProtocol(protocol.Address, protocol)
	}
	manager.failover = manager.config.Failover
	manager.router.SetFailoverProtocol(manager.failover)

	glog.Info("Start router")
	go manager.router.Start()

	// Set up sigint handling and wait for one
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	// State reporting on SIGUSR1 (testing/debugging)
	signal.Notify(ch, syscall.SIGUSR1)

	// Failover on SIGUSR2 (testing)
	if manager.router.doFailover {
		// FIXME: elvin://
		signal.Notify(ch, syscall.SIGUSR2)
	}

	for {
		sig := <-ch
		switch sig {
		case os.Interrupt:
			glog.Info("Exiting on ", sig)
			glog.Flush()
			manager.router.Stop()
			// FIXME: wait group
			os.Exit(0)
		case syscall.SIGUSR1:
			manager.router.ReportClients()
		case syscall.SIGUSR2:
			manager.router.Failover()
		}
	}

}
