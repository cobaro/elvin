package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

type Connection struct {
	conn           net.Conn
	writeChannel   chan []byte
	readTerminate  chan int
	writeTerminate chan int
}

func main() {
	listener("0.0.0.0", 2917)
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
