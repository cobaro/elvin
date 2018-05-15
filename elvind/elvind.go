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
	"fmt"
	"io"
	// "github.com/cobaro/elvin/elvin"
	"log"
	"net"
	"os"
	"os/signal"
)

func main() {
	// Argument parsing
	configFile := flag.String("config", "elvind.json", "JSON config file path")
	flag.Parse()

	// Load config
	config, err := LoadConfig(*configFile)
	if err != nil {
		fmt.Println("config load failed:", err)
		return
	}
	// fmt.Println(*config)

	// Check Protocols and set up listeners
	for _, protocol := range config.Protocols {
		switch protocol.Network {
		case "tcp":
			break
		case "udp":
		case "ssl":
			log.Println("Warning: network protocol", protocol.Network, "is currently unsupported")
			continue
		default:
			log.Println("Warning: network protocol", protocol.Network, "is unknown")
			continue
		}

		switch protocol.Marshal {
		case "xdr":
			break
		case "protobuf":
			log.Println("Warning: marshal protocol", protocol.Marshal, "is currently unsupported")
			continue
		default:
			log.Println("Warning: marshal protocol", protocol.Marshal, "is unknown")
			continue
		}
		// TODO: track listeners for shutdown
		go Listener(protocol)
	}

	// Set up sigint handling and wait for one
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	log.Println("Exiting on", <-ch)
	return
}

func Listener(protocol Protocol) {

	fmt.Println("Listening on", protocol.Network, protocol.Marshal, protocol.Address)

	ln, err := net.Listen(protocol.Network, protocol.Address)
	if err != nil {
		fmt.Println("Listen failed:", err)
		os.Exit(1)
	}

	for {
		c, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept failed:", err)
			os.Exit(1)
		}

		var conn Connection
		conn.reader = c
		conn.writer = c
		conn.state = StateNew
		conn.writeChannel = make(chan *bytes.Buffer, 4) // Some queuing allowed to smooth things out
		conn.readTerminate = make(chan int)
		conn.writeTerminate = make(chan int)

		go conn.readHandler()
		go conn.writeHandler()
	}
}

// Read n bytes from conn into buffer
func readBytes(conn io.Reader, buffer []byte, numToRead int) (int, error) {
	offset := 0
	for offset < numToRead {
		length, err := conn.Read(buffer[offset:])
		if err != nil {
			return offset + length, err
		}
		offset += length
	}
	return offset, nil
}
