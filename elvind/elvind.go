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
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
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
		// TODO: track connections
		conn := Connection{c, make(chan []byte), make(chan int), make(chan int)}
		go readHandler(conn)
		go writeHandler(conn)
	}
}

// Read n bytes from conn into buffer
func readBytes(conn net.Conn, buffer []byte, numToRead int) (int, error) {
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

func readHandler(conn Connection) {
	fmt.Println("Read Handler starting")
	defer conn.conn.Close()
	defer fmt.Println("Read Handler exiting")

	header := make([]byte, 4)

	for {
		// Read frame header
		length, err := readBytes(conn.conn, header, 4)
		if length != 4 || err != nil {
			// Deal with more errors
			if err == io.EOF {
				conn.writeTerminate <- 1
			} else {
				fmt.Println("Read Handler error:", err)
			}
			return // We're done
		}

		// Read the protocol packet
		packetSize := binary.BigEndian.Uint32(header)
		log.Println("Want to read packet of length:", packetSize)
		// TODO: buffer cache
		buffer := make([]byte, packetSize)
		length, err = readBytes(conn.conn, buffer, int(packetSize))
		if err != nil {
			// Deal with more errors
			if err == io.EOF {
				conn.writeTerminate <- 1
			} else {
				fmt.Println("Read Handler error:", err)
			}
			return // We're done
		}

		// Deal with the packet
		HandlePacket(conn, buffer)

		// FIXME: strip echo mode
		// conn.writeChannel <- buffer[0:length]
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

func HandlePacket(conn Connection, buffer []byte) (err error) {

	switch PacketId(buffer) {
	case PacketReserved:
		return errors.New("FIXME: Packet Reserved")
	case PacketSvrRqst:
		return errors.New("FIXME: Packet SvrRqst")
	case PacketSvrAdvt:
		return errors.New("FIXME: Packet SvrAdvt")
	case PacketSvrAdvtClose:
		return errors.New("FIXME: Packet SvrAdvtClose")
	case PacketUnotify:
		return errors.New("FIXME: Packet Unotify")
	case PacketNack:
		return errors.New("FIXME: Packet Nack")
	case PacketConnRqst:
		return HandleConnRqst(conn, buffer)
	case PacketConnRply:
		return errors.New("FIXME: Packet ConnRply")
	case PacketDisconnRqst:
		return errors.New("FIXME: Packet DisconnRqst")
	case PacketDisconnRply:
		return errors.New("FIXME: Packet DisconnRply")
	case PacketDisconn:
		return errors.New("FIXME: Packet Disconn")
	case PacketSecRqst:
		return errors.New("FIXME: Packet SecRqst")
	case PacketSecRply:
		return errors.New("FIXME: Packet SecRply")
	case PacketNotifyEmit:
		return errors.New("FIXME: Packet NotifyEmit")
	case PacketNotifyDeliver:
		return errors.New("FIXME: Packet NotifyDeliver")
	case PacketSubAddRqst:
		return errors.New("FIXME: Packet SubAddRqst")
	case PacketSubModRqst:
		return errors.New("FIXME: Packet SubModRqst")
	case PacketSubDelRqst:
		return errors.New("FIXME: Packet SubDelRqst")
	case PacketSubRply:
		return errors.New("FIXME: Packet SubRply")
	case PacketDropWarn:
		return errors.New("FIXME: Packet DropWarn")
	case PacketTestConn:
		return errors.New("FIXME: Packet TestConn")
	case PacketConfConn:
		return errors.New("FIXME: Packet ConfConn")
	case PacketAck:
		return errors.New("FIXME: Packet Ack")
	case PacketStatusUpdate:
		return errors.New("FIXME: Packet StatusUpdate")
	case PacketAuthRqst:
		return errors.New("FIXME: Packet AuthRqst")
	case PacketAuthCont:
		return errors.New("FIXME: Packet AuthCont")
	case PacketAuthAck:
		return errors.New("FIXME: Packet AuthAck")
	case PacketQosRqst:
		return errors.New("FIXME: Packet QosRqst")
	case PacketQosRply:
		return errors.New("FIXME: Packet QosRply")
	case PacketQnchAddRqst:
		return errors.New("FIXME: Packet QnchAddRqst")
	case PacketQnchModRqst:
		return errors.New("FIXME: Packet QnchModRqst")
	case PacketQnchDelRqst:
		return errors.New("FIXME: Packet QnchDelRqst")
	case PacketQnchRply:
		return errors.New("FIXME: Packet QnchRply")
	case PacketSubAddNotify:
		return errors.New("FIXME: Packet SubAddNotify")
	case PacketSubModNotify:
		return errors.New("FIXME: Packet SubModNotify")
	case PacketSubDelNotify:
		return errors.New("FIXME: Packet SubDelNotify")
	case PacketActivate:
		return errors.New("FIXME: Packet Activate")
	case PacketStandby:
		return errors.New("FIXME: Packet Standby")
	case PacketRestart:
		return errors.New("FIXME: Packet Restart")
	case PacketShutdown:
		return errors.New("FIXME: Packet Shutdown")
	case PacketServerReport:
		return errors.New("FIXME: Packet ServerReport")
	case PacketServerNack:
		return errors.New("FIXME: Packet ServerNack")
	case PacketServerStatsReport:
		return errors.New("FIXME: Packet ServerStatsReport")
	case PacketClstJoinRqst:
		return errors.New("FIXME: Packet ClstJoinRqst")
	case PacketClstJoinRply:
		return errors.New("FIXME: Packet ClstJoinRply")
	case PacketClstTerms:
		return errors.New("FIXME: Packet ClstTerms")
	case PacketClstNotify:
		return errors.New("FIXME: Packet ClstNotify")
	case PacketClstRedir:
		return errors.New("FIXME: Packet ClstRedir")
	case PacketClstLeave:
		return errors.New("FIXME: Packet ClstLeave")
	case PacketFedConnRqst:
		return errors.New("FIXME: Packet FedConnRqst")
	case PacketFedConnRply:
		return errors.New("FIXME: Packet FedConnRply")
	case PacketFedSubReplace:
		return errors.New("FIXME: Packet FedSubReplace")
	case PacketFedNotify:
		return errors.New("FIXME: Packet FedNotify")
	case PacketFedSubDiff:
		return errors.New("FIXME: Packet FedSubDiff")
	case PacketFailoverConnRqst:
		return errors.New("FIXME: Packet FailoverConnRqst")
	case PacketFailoverConnRply:
		return errors.New("FIXME: Packet FailoverConnRply")
	case PacketFailoverMaster:
		return errors.New("FIXME: Packet FailoverMaster")
	default:
		return errors.New("FIXME: Packet Unknown")
	}
}

// Handle a Connection request
func HandleConnRqst(conn Connection, buffer []byte) (err error) {
	// FIXME: no range checking
	connRqst := new(ConnRqst)
	err = connRqst.Decode(buffer)
	// FIXME: Connection handling
	fmt.Println(connRqst)
	return err
}
