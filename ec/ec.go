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
	"fmt"
	"github.com/cobaro/elvin/elvin"
	"log"
	"os"
	"os/signal"
	"time"
)

// Handle Disconnects stub
func disconnector(client *elvin.Client) {
}

type arguments struct {
	help     bool
	endpoint string
	number   int
}

func main() {
	// Argument parsing
	args := flags()
	flag.Parse()

	ec := elvin.NewClient(args.endpoint, nil, nil, nil)
	go disconnector(ec)

	ec.Options = make(map[string]interface{})
	// FIXME: At some point let's formalize these as test cases
	// ec.Options["TestNack"] = 1
	// ec.Options["TestDisconn"] = 1
	// log.Printf("Options:%v\n", ec.Options)

	if err := ec.Connect(); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
	log.Printf("connected to %s", args.endpoint)

	// FIXME: do a NewSubscription()
	sub := new(elvin.Subscription)
	sub.Expression = "require(int32)"
	// sub.Expression = "bogus"
	sub.AcceptInsecure = true
	sub.Keys = nil
	sub.Notifications = make(chan map[string]interface{})

	if err := ec.Subscribe(sub); err != nil {
		log.Printf("Subscribe failed %v", err)
	} else {
		log.Printf("Subscribe succeeded %v", sub)
		if err := ec.SubscriptionModify(sub, "bogus", true, nil, nil); err != nil {
			log.Printf("SubMod failed %v", err)
		} else {
			log.Printf("SubMod succeeded %v", sub)
		}
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	received := 0
	timeStart := time.Now()
Loop:
	for {
		select {
		case sig := <-ch:
			log.Printf("Exiting on %v", sig)
			break Loop
		case nfn := <-sub.Notifications:
			if args.number == 1 {
				log.Printf("Received notification:\n%v", nfn)
			} else {
				received++
				if received == args.number {
					timeNow := time.Now()
					received = 0
					fmt.Println(timeNow.Sub(timeStart))
					timeStart = timeNow
				}
			}
		}
	}

	if err := ec.SubscriptionDelete(sub); err != nil {
		log.Printf("SubDel failed %v", err)
	} else {
		log.Printf("SubDel succeeded %v", sub)
	}

	// Exit a little gracefully
	log.Printf("Disconnecting")
	if err := ec.Disconnect(); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}

	log.Printf("Disconnected")
	os.Exit(0)
}

// Argument parsing
func flags() (args arguments) {
	flag.BoolVar(&args.help, "h", false, "Print this help")
	flag.StringVar(&args.endpoint, "e", "localhost:2917", "host:port of router")
	flag.IntVar(&args.number, "n", 1, "number of notifications to send")
	flag.Parse()

	if args.help {
		flag.Usage()
		os.Exit(0)
	}

	return args
}
