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
	"os/signal"
)

func main() {
	// Argument parsing
	flag.Parse()

	endpoint := "localhost:2917"
	ec := elvin.NewClient(endpoint, nil, nil, nil)

	if err := ec.Connect(); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
	log.Printf("connected to %s", endpoint)

	sub := new(elvin.Subscription)
	sub.Expression = "require(int32)"
	sub.AcceptInsecure = true
	sub.Keys = nil
	sub.Notifications = make(chan map[string]interface{}) // FIXME: do a NewSubscription()

	if err := ec.Subscribe(sub); err != nil {
		log.Printf("Subscribe failed %v", err)
	} else {
		log.Printf("Subscribe succeeded %v", sub)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

Loop:
	for {
		select {
		case sig := <-ch:
			log.Printf("Exiting on %v", sig)
			break Loop
		case nfn := <-sub.Notifications:
			log.Printf("Received notification:\n%v", nfn)
		}
	}

	// Exit a little gracefully
	if err := ec.Disconnect(); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}

	log.Printf("disconnected")
	os.Exit(0)
}
