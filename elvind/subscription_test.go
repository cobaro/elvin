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
	"github.com/cobaro/elvin/elvin"
	"log"
	"os"
	"testing"
	"time"
)

var protocol Protocol
var client *elvin.Client

func TestMain(m *testing.M) {
	flag.Parse() // FIXME: do something about logging
	// Create a router instance using standard test config
	protocol = Protocol{"tcp", "xdr", "0.0.0.0:2917"}
	var router Router
	router.SetMaxConnections(10)
	router.SetDoFailover(false)
	router.SetTestConnInterval(10 * time.Second)
	router.SetTestConnTimeout(10 * time.Second)
	router.AddProtocol(protocol.Address, protocol)
	go router.Start()
	time.Sleep(time.Millisecond * 10) // Yield to get that started

	// Create and connect a client
	client = elvin.NewClient(protocol.Address, nil, nil, nil)
	if err := client.Connect(); err != nil {
		log.Printf("Connect failed: %v", err)
		return
	}

	// Run all our tests
	ret := m.Run()

	if err := client.Disconnect(); err != nil {
		log.Printf("Disconnect failed: %v", err)
		return
	}
	router.Stop()

	os.Exit(ret)
}

func TestSubscriptionFail(t *testing.T) {
	// Add a subscription
	sub := new(elvin.Subscription)
	sub.Expression = "bogus"
	sub.AcceptInsecure = true
	sub.Keys = nil
	sub.Notifications = make(chan map[string]interface{})

	if err := client.Subscribe(sub); err == nil {
		t.Errorf("Subscribe passed %v", err)
		return
	}
}

func TestSubscriptionPass(t *testing.T) {
	// Create a client
	// Add a subscription
	sub := new(elvin.Subscription)
	sub.Expression = "require(TestPass)"
	sub.AcceptInsecure = true
	sub.Keys = nil
	sub.Notifications = make(chan map[string]interface{})

	if err := client.Subscribe(sub); err != nil {
		t.Errorf("Subscribe failed %v", err)
		return
	}

	var nfn = map[string]interface{}{"TestPass": int32(1)}
	if err := client.Notify(nfn, true, nil); err != nil {
		t.Errorf("Notify failed")
		return
	}

	select {
	case nfn := <-sub.Notifications:
		if nfn["TestPass"] != int32(1) {
			t.Errorf("Received unmatched notification")
			return
		}

	case <-time.After(1 * time.Second):
		t.Errorf("Too slow!")
		return
	}

	if err := client.SubscriptionDelete(sub); err != nil {
		t.Errorf("Unsubscribe failed %v", err)
		return
	}

}
