// Copyright 2018 Cobaro Pty Ltd. All Rights Reserved.

// This file is part of elvind
//
// elvind is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// elvind is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with elvind. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
)

type Connection struct {
	conn           net.Conn
	writeChannel   chan []byte
	readTerminate  chan int
	writeTerminate chan int
}

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
	fmt.Println(*config)

	for _, uri := range config.Protocols {
		uri, err := url.Parse(uri)
		if err != nil {
			log.Fatal("URI parsing failed:", uri)
		}
		log.Println("uri:", uri)

		if uri.Scheme != "elvin" {
			log.Fatal("URI parsing failed:", "Scheme is not elvin")
		}
	}
	go listener("0.0.0.0", 2917)

	// Set up sigint handling and wait for one
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	log.Println("Exiting on", <-ch)
	return
}

func listener(host string, port int) {

	listento := host + ":" + strconv.Itoa(port)

	fmt.Println("Listening on " + listento)

	ln, err := net.Listen("tcp", listento)
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
		conn := Connection{c, make(chan []byte), make(chan int), make(chan int)}
		go readHandler(conn)
		go writeHandler(conn)
	}
}

func readHandler(conn Connection) {
	fmt.Println("Read Handler starting")
	for {
		buf := make([]byte, 4096)
		length, err := conn.conn.Read(buf)
		fmt.Println("Read Handler received", length, "bytes")
		if err != nil {
			// Deal with more errors
			if err == io.EOF {
				conn.writeTerminate <- 1
			} else {
				fmt.Println("Read Handler error:", err)
			}
			// Clean up connection
			conn.conn.Close()
			fmt.Println("Read Handler exiting")
			return // We're done
		}

		// Decoding etc as future exercise
		// if buf.Len() >= HEADER_LENGTH {
		// get length of packet
		// read until we get length bytes
		// }

		// For now just echo
		conn.writeChannel <- buf[0:length]
	}
}

func writeHandler(conn Connection) {
	fmt.Println("Write Handler starting ")
	for {
		select {
		case buf := <-conn.writeChannel:
			// For now just echo it back (and assume full write)
			_, err := conn.conn.Write(buf)
			if err != nil {
				// Deal with more errors
				if err == io.EOF {
					conn.conn.Close()
				} else {
					fmt.Println("Write handler error:", err)
				}
				return // We're done, cleanup done by read
			}
		case <-conn.writeTerminate:
			fmt.Println("Write Handler exiting ")
			return
		}
	}
}
