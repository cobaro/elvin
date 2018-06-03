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

	go signalHandler()

	endpoint := "localhost:2917"
	ep := elvin.NewClient(endpoint, nil, nil, nil)

	if err := ep.Connect(); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
	log.Printf("connected to %s", endpoint)

	// Send a dumb single message for now
	nfn := make(map[string]interface{})
	nfn["int32"] = int32(3232)
	nfn["int64"] = int64(646464646464)
	nfn["string"] = "string"
	nfn["opaque"] = []byte{0, 1, 2, 3, 127, 255}
	nfn["float64"] = 424242.42

	if err := ep.Notify(nfn, true, nil); err != nil {
		log.Printf("Notify failed")
	}

	if err := ep.Disconnect(); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
	log.Printf("disconnected")
	os.Exit(0)
}

func signalHandler() {
	// Set up sigint handling and wait for one
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	log.Printf("Exiting on %v", <-ch)
	os.Exit(1)
}
