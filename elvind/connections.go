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
	"github.com/golang/glog"
	"io"
	"math/rand"
	"sync"
	// "time"
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
	id             uint32
	subs           map[uint32]*Subscription
	reader         io.Reader
	writer         io.Writer
	closer         io.Closer
	state          int
	writeChannel   chan *bytes.Buffer
	readTerminate  chan int
	writeTerminate chan int
}

// A buffer pool as we use lots of these for writing to
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// Global connections

type Connections struct {
	connections map[uint32]*Connection // initialized in init()
	lock        sync.Mutex             // initialized automatically
}

var connections Connections

func init() {
	// rand.New(rand.NewSource(time.Now().UnixNano()))
	connections.connections = make(map[uint32]*Connection)
}

// Return our unique 32 bit unsigned identifier
func (conn *Connection) Id() uint32 {
	return conn.id
}

// Create a unique 32 bit unsigned integer id
func (conn *Connection) MakeId() {
	connections.lock.Lock()
	defer connections.lock.Unlock()
	var c uint32 = rand.Uint32()
	for {
		_, err := connections.connections[c]
		if !err {
			break
		}
		c++
	}

	if glog.V(3) {
		glog.Infof("conn id is %d", c)
	}
	conn.id = c
	connections.connections[c] = conn
	return
}

// Read n bytes from reader into buffer which must be big enough
func readBytes(reader io.Reader, buffer []byte, numToRead int) (int, error) {
	offset := 0
	for offset < numToRead {
		if glog.V(4) {
			glog.Infof("offset = %d, numToRead = %d", offset, numToRead)
		}
		length, err := reader.Read(buffer[offset:numToRead])
		if err != nil {
			return offset + length, err
		}
		offset += length
	}
	return offset, nil
}

// Handle reading for now run as a goroutine
func (conn *Connection) readHandler() {
	if glog.V(2) {
		glog.Infof("Read Handler starting")
	}
	defer conn.closer.Close()
	if glog.V(2) {
		defer glog.Infof("Read Handler exiting")
	}

	header := make([]byte, 4)
	buffer := make([]byte, 2048)

	for {
		// Read frame header
		length, err := readBytes(conn.reader, header, 4)
		if length != 4 || err != nil {
			// Deal with more errors
			if err == io.EOF {
				conn.writeTerminate <- 1
			} else {
				glog.Errorf("Read Handler error: %v", err)
			}
			return // We're done
		}

		// Read the protocol packet, starting with it's length
		packetSize := int32(binary.BigEndian.Uint32(header))
		// Grow our buffer if needed
		if int(packetSize) > len(buffer) {
			if glog.V(4) {
				glog.Infof("Growing buffer to %d bytes", packetSize)
			}
			buffer = make([]byte, packetSize)
		}

		length, err = readBytes(conn.reader, buffer, int(packetSize))
		if err != nil {
			// Deal with more errors
			if err == io.EOF {
				conn.writeTerminate <- 1
			} else {
				glog.Errorf("Read Handler error: %v", err)
			}
			return // We're done
		}

		// Deal with the packet
		if err = conn.HandlePacket(buffer); err != nil {
			glog.Errorf("Read Handler error: %v", err)
			// FIXME: protocol error
		}

	}
}

// Handle writing for now run as a goroutine
func (conn *Connection) writeHandler() {
	if glog.V(2) {
		glog.Infof("Write Handler starting")
	}
	if glog.V(2) {
		defer glog.Infof("Write Handler exiting")
	}
	defer conn.closer.Close()

	header := make([]byte, 4)

	for {
		select {
		case buffer := <-conn.writeChannel:

			// Write the frame header (packetsize)
			binary.BigEndian.PutUint32(header, uint32(buffer.Len()))
			_, err := conn.writer.Write(header)
			if err != nil {
				// Deal with more errors
				if err == io.EOF {
					conn.closer.Close()
				} else {
					glog.Errorf("Unexpected write error: %v", err)
				}
				bufferPool.Put(buffer)
				return // We're done, cleanup done by read
			}

			// Write the packet
			_, err = buffer.WriteTo(conn.writer)
			if err != nil {
				// Deal with more errors
				if err == io.EOF {
					conn.closer.Close()
				} else {
					glog.Errorf("Unexpected write error: %v", err)
				}
				bufferPool.Put(buffer)
				return // We're done, cleanup done by read
			}
		case <-conn.writeTerminate:
			return
		}
	}
}

// Handle a protocol packet
func (conn *Connection) HandlePacket(buffer []byte) (err error) {

	if glog.V(4) {
		glog.Infof("received %s", elvin.PacketIdString(elvin.PacketId(buffer)))
	}
	switch elvin.PacketId(buffer) {

	// Packets a router should never receive
	case elvin.PacketDropWarn:
	case elvin.PacketReserved:
	case elvin.PacketNotifyDeliver:
	case elvin.PacketNack:
	case elvin.PacketConnRply:
	case elvin.PacketDisconnRply:
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
	case elvin.PacketSubAddNotify:
	case elvin.PacketSubModNotify:
	case elvin.PacketSubDelNotify:
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
		case elvin.PacketDisconnRqst:
			return conn.HandleDisconnRqst(buffer)
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
			return conn.HandleSubModRqst(buffer)
		case elvin.PacketSubDelRqst:
			return conn.HandleSubDelRqst(buffer)
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
		case elvin.PacketActivate:
			return errors.New("FIXME: Packet Activate")
		case elvin.PacketStandby:
			return errors.New("FIXME: Packet Standby")
		case elvin.PacketRestart:
			return errors.New("FIXME: Packet Restart")
		case elvin.PacketShutdown:
			return errors.New("FIXME: Packet Shutdown")
		default:
			return fmt.Errorf("FIXME: Packet Unknown [%d]", elvin.PacketId(buffer))
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
		conn.state = StateClosed
		conn.closer.Close()
	}

	// We're now connected
	conn.MakeId()
	conn.state = StateConnected
	conn.subs = make(map[uint32]*Subscription)

	// Respond with a Connection Reply
	connRply := new(elvin.ConnRply)
	connRply.Xid = connRqst.Xid
	// FIXME; totally bogus
	connRply.Options = connRqst.Options

	glog.Infof("New client %d connected", conn.id)

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	connRply.Encode(buf)
	conn.writeChannel <- buf

	return nil
}

// Handle a Disconnection Request
func (conn *Connection) HandleDisconnRqst(buffer []byte) (err error) {
	// FIXME: no range checking
	disconnRqst := new(elvin.DisconnRqst)
	if err = disconnRqst.Decode(buffer); err != nil {
		conn.state = StateClosed
		conn.closer.Close()
	}

	// We're now disconnecting
	conn.state = StateDisconnecting

	// Respond with a Disconnection Reply
	DisconnRply := new(elvin.DisconnRply)
	DisconnRply.Xid = disconnRqst.Xid

	glog.Infof("client %d disconnected", conn.Id)

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	DisconnRply.Encode(buf)
	conn.writeChannel <- buf

	for subid, _ := range conn.subs {
		// FIXME: send subscription removal to sub engine
		delete(conn.subs, subid)
	}

	connections.lock.Lock()
	defer connections.lock.Unlock()
	delete(connections.connections, conn.id)

	return nil
}

// Handle a NotifyEmit
func (conn *Connection) HandleNotifyEmit(buffer []byte) (err error) {
	// FIXME: no range checking
	ne := new(elvin.NotifyEmit)
	err = ne.Decode(buffer)

	// FIXME: NotifyDeliver

	// As a dummy for now we're going to send every message we see
	// to every subscription as if all evaluate to true. Don't
	// worry about the giant lock - this all goes away
	connections.lock.Lock()
	defer connections.lock.Unlock()
	nd := new(elvin.NotifyDeliver)
	nd.NameValue = ne.NameValue

	for connid, connection := range connections.connections {
		nd.Insecure = make([]uint64, len(connection.subs))
		i := 0
		for id, _ := range connection.subs {
			nd.Insecure[i] = uint64(connid)<<32 | uint64(id)
			i++
		}
		buf := bufferPool.Get().(*bytes.Buffer)
		nd.Encode(buf)
		connection.writeChannel <- buf
	}
	return nil
}

// Handle a Subscription Add
func (conn *Connection) HandleSubAddRqst(buffer []byte) (err error) {
	// FIXME: no range checking
	subRqst := new(elvin.SubAddRqst)
	err = subRqst.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
	}

	ast, nack := Parse(subRqst.Expression)
	if nack != nil {
		nack.Xid = subRqst.Xid
		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		conn.writeChannel <- buf
		return nil
	}

	// Create a subscription and add it to the subscription store
	var sub Subscription
	sub.Ast = ast
	sub.AcceptInsecure = subRqst.AcceptInsecure
	sub.Keys = subRqst.Keys

	// Create a unique sub id
	var s uint32 = rand.Uint32()
	for {
		_, err := conn.subs[s]
		if !err {
			break
		}
		s++
	}
	conn.subs[s] = &sub
	sub.Subid = (uint64(conn.id) << 32) | uint64(s)

	// FIXME: send subscription addition to sub engine

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

// Handle a Subscription Delete
func (conn *Connection) HandleSubDelRqst(buffer []byte) (err error) {
	// FIXME: no range checking
	subDelRqst := new(elvin.SubDelRqst)
	err = subDelRqst.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
	}

	// If deletion fails then nack and disconn
	idx := uint32(subDelRqst.Subid & 0xfffffffff)
	sub, exists := conn.subs[idx]
	if !exists {
		nack := new(elvin.Nack)
		nack.ErrorCode = elvin.ErrorsUnknownSubid
		nack.Message = elvin.ProtocolErrors[elvin.ErrorsUnknownSubid]
		nack.Args = make([]interface{}, 1)
		nack.Args[0] = sub.Subid
		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		conn.writeChannel <- buf

		// FIXME Disconnect as that's a protocol violation
		return nil
	}

	// Remove it from the connection
	delete(conn.subs, idx)

	// Respond with a SubRply
	subRply := new(elvin.SubRply)
	subRply.Xid = subDelRqst.Xid
	subRply.Subid = subDelRqst.Subid

	// FIXME: send subscription deletion to sub engine

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	subRply.Encode(buf)
	conn.writeChannel <- buf
	return nil
}

func (conn *Connection) HandleSubModRqst(buffer []byte) (err error) {
	// FIXME: no range checking
	subModRqst := new(elvin.SubModRqst)
	err = subModRqst.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
	}

	// If modify fails then nack and disconn
	idx := uint32(subModRqst.Subid & 0xfffffffff)
	sub, exists := conn.subs[idx]
	if !exists {
		nack := new(elvin.Nack)
		nack.ErrorCode = elvin.ErrorsUnknownSubid
		nack.Message = elvin.ProtocolErrors[elvin.ErrorsUnknownSubid]
		nack.Args = make([]interface{}, 1)
		nack.Args[0] = sub.Subid
		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		conn.writeChannel <- buf

		// FIXME Disconnect as that's a protocol violation
		return nil
	}

	// FIXME: At some point this sub will need a lock but for now it's all handled in the read goroutine
	// FIXME: And any update to the sub should be all or nothing

	// Check the subscription expression. Empty is ok. Incorrect means bail.
	if len(subModRqst.Expression) > 0 {
		ast, nack := Parse(subModRqst.Expression)
		if nack != nil {
			nack.Xid = subModRqst.Xid
			buf := bufferPool.Get().(*bytes.Buffer)
			nack.Encode(buf)
			conn.writeChannel <- buf
			return nil
		}
		sub.Ast = ast
	}

	// AcceptInsecure is the only piece that must have a value - and it is allowed to be the same
	sub.AcceptInsecure = subModRqst.AcceptInsecure

	// We ignore deletion requests that don't exist
	if len(subModRqst.DelKeys) > 0 {
		// FIXME: implement
	}

	// We ignore addition requests that already exist
	if len(subModRqst.AddKeys) > 0 {
		// FIXME: implement
	}

	// Respond with a SubRply
	subRply := new(elvin.SubRply)
	subRply.Xid = subModRqst.Xid
	subRply.Subid = subModRqst.Subid

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	subRply.Encode(buf)
	conn.writeChannel <- buf
	return nil
}
