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
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cobaro/elvin/elvin"
	"io"
	"net"
	"sync"
)

// Connection States
const (
	StateNew = iota
	StateConnected
	StateDisconnecting
	StateClosed
)

// A Connection (e.g. a socket)
type Connection struct {
	conn           net.Conn
	state          int
	writeChannel   chan *bytes.Buffer
	readTerminate  chan int
	writeTerminate chan int
	// lock           sync.Mutex
}

// Handle reading for now run as a goroutine
func (conn *Connection) readHandler() {
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
		// log.Println("Want to read packet of length:", packetSize)
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
		if err = conn.HandlePacket(buffer); err != nil {
			fmt.Println(err)
		}

	}
}

// Handle writing for now run as a goroutine
func (conn *Connection) writeHandler() {
	fmt.Println("Write Handler starting ")
	header := make([]byte, 4)

	for {
		select {
		case buffer := <-conn.writeChannel:

			// Write the frame header (packetsize)
			binary.BigEndian.PutUint32(header, uint32(buffer.Len()))
			_, err := conn.conn.Write(header)
			if err != nil {
				// Deal with more errors
				if err == io.EOF {
					conn.conn.Close()
				} else {
					fmt.Println("Unexepcted write error:", err)
				}
				return // We're done, cleanup done by read
			}

			// Write the packet
			_, err = buffer.WriteTo(conn.conn)
			if err != nil {
				// Deal with more errors
				if err == io.EOF {
					conn.conn.Close()
				} else {
					fmt.Println("Unexepcted write error:", err)
				}
				return // We're done, cleanup done by read
			}
		case <-conn.writeTerminate:
			fmt.Println("Write Handler exiting ")
			return
		}
	}
}

// Handle a protocol packet
func (conn *Connection) HandlePacket(buffer []byte) (err error) {

	fmt.Println("HandlePacket", elvin.PacketIdString(elvin.PacketId(buffer)))

	switch elvin.PacketId(buffer) {

	// Packets a router should never receive
	case elvin.PacketDropWarn:
	case elvin.PacketReserved:
	case elvin.PacketNotifyDeliver:
		return fmt.Errorf("ProtocolError: %s received", elvin.PacketIdString(elvin.PacketId(buffer)))

		// Protocol Packets not planned for the short term
	case elvin.PacketSvrRqst:
	case elvin.PacketSvrAdvt:
	case elvin.PacketSvrAdvtClose:
	case elvin.PacketQnchAddRqst:
	case elvin.PacketQnchModRqst:
	case elvin.PacketQnchDelRqst:
	case elvin.PacketQnchRply:
	case elvin.PacketClstJoinRqst:
	case elvin.PacketClstJoinRply:
	case elvin.PacketClstTerms:
	case elvin.PacketClstNotify:
	case elvin.PacketClstRedir:
	case elvin.PacketClstLeave:
	case elvin.PacketFedConnRqst:
	case elvin.PacketFedConnRply:
	case elvin.PacketFedSubReplace:
	case elvin.PacketFedNotify:
	case elvin.PacketFedSubDiff:
	case elvin.PacketFailoverConnRqst:
	case elvin.PacketFailoverConnRply:
	case elvin.PacketFailoverMaster:
	case elvin.PacketServerReport:
	case elvin.PacketServerNack:
	case elvin.PacketServerStatsReport:
		return fmt.Errorf("UnimplementedError: %s received", elvin.PacketIdString(elvin.PacketId(buffer)))
	}

	// Packets dependent upon Client's connection state
	switch conn.state {
	case StateNew:
		// Connect and Unotify are the only valid packets without
		// a properly established connection
		switch elvin.PacketId(buffer) {
		case elvin.PacketConnRqst:
			return conn.HandleConnRqst(buffer)
		case elvin.PacketUnotify:
			return errors.New("FIXME: Packet Unotify")
		default:
			return fmt.Errorf("ProtocolError: %s received", elvin.PacketIdString(elvin.PacketId(buffer)))
		}

	case StateConnected:
		// Deal with packets that can arrive whilst connected

		// FIXME: implement or move this lot in the short term
		switch elvin.PacketId(buffer) {
		case elvin.PacketNack:
			return errors.New("FIXME: Packet Nack")
		case elvin.PacketConnRply:
			return errors.New("FIXME: Packet ConnRply")
		case elvin.PacketDisconnRqst:
			return errors.New("FIXME: Packet DisconnRqst")
		case elvin.PacketDisconnRply:
			return errors.New("FIXME: Packet DisconnRply")
		case elvin.PacketDisconn:
			return errors.New("FIXME: Packet Disconn")
		case elvin.PacketSecRqst:
			return errors.New("FIXME: Packet SecRqst")
		case elvin.PacketSecRply:
			return errors.New("FIXME: Packet SecRply")
		case elvin.PacketNotifyEmit:
			return conn.HandleNotifyEmit(buffer)
		case elvin.PacketSubAddRqst:
			return conn.HandleSubAddRqst(buffer)
		case elvin.PacketSubModRqst:
			return errors.New("FIXME: Packet SubModRqst")
		case elvin.PacketSubDelRqst:
			return errors.New("FIXME: Packet SubDelRqst")
		case elvin.PacketSubRply:
			return errors.New("FIXME: Packet SubRply")
		case elvin.PacketTestConn:
			return errors.New("FIXME: Packet TestConn")
		case elvin.PacketConfConn:
			return errors.New("FIXME: Packet ConfConn")
		case elvin.PacketAck:
			return errors.New("FIXME: Packet Ack")
		case elvin.PacketStatusUpdate:
			return errors.New("FIXME: Packet StatusUpdate")
		case elvin.PacketAuthRqst:
			return errors.New("FIXME: Packet AuthRqst")
		case elvin.PacketAuthCont:
			return errors.New("FIXME: Packet AuthCont")
		case elvin.PacketAuthAck:
			return errors.New("FIXME: Packet AuthAck")
		case elvin.PacketQosRqst:
			return errors.New("FIXME: Packet QosRqst")
		case elvin.PacketQosRply:
			return errors.New("FIXME: Packet QosRply")
		case elvin.PacketSubAddNotify:
			return errors.New("FIXME: Packet SubAddNotify")
		case elvin.PacketSubModNotify:
			return errors.New("FIXME: Packet SubModNotify")
		case elvin.PacketSubDelNotify:
			return errors.New("FIXME: Packet SubDelNotify")
		case elvin.PacketActivate:
			return errors.New("FIXME: Packet Activate")
		case elvin.PacketStandby:
			return errors.New("FIXME: Packet Standby")
		case elvin.PacketRestart:
			return errors.New("FIXME: Packet Restart")
		case elvin.PacketShutdown:
			return errors.New("FIXME: Packet Shutdown")
		default:
			return errors.New("FIXME: Packet Unknown")
		}

	case StateDisconnecting:
	case StateClosed:
		return fmt.Errorf("ProtocolError: %s received", elvin.PacketIdString(elvin.PacketId(buffer)))
	}

	return fmt.Errorf("Error: %s received and not handled", elvin.PacketIdString(elvin.PacketId(buffer)))
}

// Handle a Connection Request
func (conn *Connection) HandleConnRqst(buffer []byte) (err error) {
	// FIXME: no range checking
	connRqst := new(elvin.ConnRqst)
	if err = connRqst.Decode(buffer); err != nil {
		// FIXME: Send a nack
		// FIXME: disconnect (and move the close()
		conn.conn.Close()
	}
	// fmt.Println(connRqst)

	// We're now connected
	conn.state = StateConnected

	// Respond with a Connection Reply
	connRply := new(elvin.ConnRply)
	connRply.Xid = connRqst.Xid
	// FIXME; totally bogus
	connRply.Options = connRqst.Options

	fmt.Println("Connected")

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	connRply.Encode(buf)
	conn.writeChannel <- buf

	return nil
}

// Handle a NotifyEmit
func (conn *Connection) HandleNotifyEmit(buffer []byte) (err error) {
	// FIXME: no range checking
	ne := new(elvin.NotifyEmit)
	err = ne.Decode(buffer)
	// fmt.Println(ne)

	// FIXME: NotifyDeliver
	fmt.Println("Received", ne)

	return err
}

// Handle a NotifyEmit
func (conn *Connection) HandleSubAddRqst(buffer []byte) (err error) {
	// FIXME: no range checking
	subRqst := new(elvin.SubAddRqst)
	err = subRqst.Decode(buffer)
	fmt.Println("Received", subRqst)

	ast, err := Parse(subRqst.Expression)
	if err != nil {
		fmt.Println("FIXME: nack")
		return nil
	}

	// Create a subscription and add it to the subscription store
	var sub Subscription
	sub.Ast = ast
	sub.AcceptInsecure = subRqst.AcceptInsecure
	sub.Keys = subRqst.Keys
	sub.Add(conn)

	// Respond with a SubRply
	subRply := new(elvin.SubRply)
	subRply.Xid = subRqst.Xid
	subRply.Subid = sub.Subid

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	subRply.Encode(buf)
	conn.writeChannel <- buf

	return nil
}

// A buffer pool as we use lots of these for writing to
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
