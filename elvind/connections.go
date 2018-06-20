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
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// Connection States
const (
	StateNew = iota
	StateConnected
	StateDisconnecting
	StateClosed
)

// Return state (synchronized)
func (conn *Connection) State() uint32 {
	return atomic.LoadUint32(&conn.state)
}

// Get state (synchronized)
func (conn *Connection) SetState(val uint32) {
	atomic.StoreUint32(&conn.state, val)
}

// TestConn/ConfConn Timeout States
const (
	TestConnIdle = iota
	TestConnAwaitingResponse
	TestConnHadResponse
)

// Set TestConn/ConfConn Timeout (synchronized)
func (conn *Connection) TestConnState() uint32 {
	return atomic.LoadUint32(&conn.testConnState)
}

// Get TestConn/ConfConn Timeout (synchronized)
func (conn *Connection) SetTestConnState(val uint32) {
	atomic.StoreUint32(&conn.testConnState, val)
}

// A Connection (e.g. a socket)
type Connection struct {
	id             int32
	subs           map[int32]*Subscription
	quenches       map[int32]*Quench
	reader         io.Reader
	writer         io.Writer
	closer         io.Closer
	state          uint32
	testConnState  uint32
	writeChannel   chan *bytes.Buffer
	writeTerminate chan int

	// Configurable options
	testConnInterval time.Duration
	testConnTimeout  time.Duration
}

// A buffer pool as we use lots of these for writing to
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// Global connections

type Connections struct {
	connections map[int32]*Connection // initialized in init()
	lock        sync.Mutex            // initialized automatically
}

var connections Connections

func init() {
	// rand.New(rand.NewSource(time.Now().UnixNano()))
	connections.connections = make(map[int32]*Connection)
}

// Return our unique 32 bit unsigned identifier
func (conn *Connection) ID() int32 {
	return conn.id
}

// Create a unique 32 bit unsigned integer id
func (conn *Connection) MakeID() {
	connections.lock.Lock()
	defer connections.lock.Unlock()
	var c int32 = rand.Int31()
	for {
		_, err := connections.connections[c]
		if !err {
			break
		}
		c++
	}

	if glog.V(4) {
		glog.Infof("conn id is %d", c)
	}
	conn.id = c
	connections.connections[c] = conn
	return
}

func (conn *Connection) Close() {
	if glog.V(4) {
		glog.Infof("Close()ing client %d", conn.ID())
	}
	select {
	case conn.writeTerminate <- 1:
	default:
	}
	conn.closer.Close()
	connections.lock.Lock()
	delete(connections.connections, conn.ID())
	// FIXME: Clean up subscriptions
	connections.lock.Unlock()

}

// Read n bytes from reader into buffer which must be big enough
func readBytes(reader io.Reader, buffer []byte, numToRead int) (int, error) {
	offset := 0
	for offset < numToRead {
		if glog.V(5) {
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
	if glog.V(4) {
		glog.Infof("Read Handler starting")
	}
	if glog.V(4) {
		defer glog.Infof("Read Handler exiting")
	}

	header := make([]byte, 4)
	buffer := make([]byte, 2048)

	for {
		// Read frame header
		length, err := readBytes(conn.reader, header, 4)
		if length != 4 || err != nil {
			break // We're done
		}

		// Read the protocol packet, starting with it's length
		packetSize := int(binary.BigEndian.Uint32(header))
		// Grow our buffer if needed
		if packetSize > len(buffer) {
			if glog.V(4) {
				glog.Infof("Growing buffer to %d bytes", packetSize)
			}
			buffer = make([]byte, packetSize)
		}

		length, err = readBytes(conn.reader, buffer, packetSize)
		if length != packetSize || err != nil {
			break // We're done
		}

		// Deal with the packet
		if err = conn.HandlePacket(buffer); err != nil {
			glog.Errorf("Read Handler error: %v", err)
			// FIXME: protocol error
			break
		}
	}
	if glog.V(4) {
		glog.Infof("Read Handler Close()")
	}
	conn.Close()
}

// Handle writing for now run as a goroutine
func (conn *Connection) writeHandler() {
	if glog.V(4) {
		glog.Infof("Write Handler starting")
		defer glog.Infof("Write Handler exiting")
	}

	header := make([]byte, 4)

	// TestConn and ConfConn use two timers:
	//
	//   The first, intervalTimer, is  how long a connection
	//   can be idle before we send a TestConn and it's configured
	//   via the configuration option TestConnInterval.
	//   A value of zero means disabled so we simply set it up for
	//   a long time as way of keeping the select loop simple.
	//
	//   The second testConnTimeout is how long we should wait for a
	//   response.
	defaultTimeout := conn.testConnInterval
	if defaultTimeout == 0 {
		defaultTimeout = math.MaxInt64
	}
	currentTimeout := defaultTimeout
	conn.SetTestConnState(TestConnIdle)

	// Our write loop waits for data to write and sends it
	// It can be terminated via the writeTerminate channel
	// It runs a Test/ConfConn timer if configured
	for {
		select {
		case buffer := <-conn.writeChannel:

			// Write the frame header (packetsize)
			binary.BigEndian.PutUint32(header, uint32(buffer.Len()))
			_, err := conn.writer.Write(header)
			if err != nil {
				// Deal with more errors
				if err != io.EOF {
					glog.Errorf("Unexpected write error: %v", err)
				}
				bufferPool.Put(buffer)
				return // We're done, cleanup done by read
			}

			// Write the packet
			_, err = buffer.WriteTo(conn.writer)
			if err != nil {
				// Deal with more errors
				if err != io.EOF {
					glog.Errorf("Unexpected write error: %v", err)
				}
				bufferPool.Put(buffer)
				return // We're done, cleanup done by read
			}
		case <-conn.writeTerminate:
			return // We're done, cleanup done by read

		case <-time.After(currentTimeout):
			if glog.V(4) {
				glog.Infof("writeHandler timeout: %d", conn.TestConnState())
			}

			switch conn.TestConnState() {
			case TestConnIdle:
				conn.SetTestConnState(TestConnAwaitingResponse)
				currentTimeout = conn.testConnTimeout
				testConn := new(elvin.TestConn)
				writeBuf := new(bytes.Buffer)
				testConn.Encode(writeBuf)
				conn.writeChannel <- writeBuf
			case TestConnAwaitingResponse:
				if glog.V(3) {
					glog.Infof("Closing client %d for not responding to TestConn", conn.ID())
				}
				// FIXME:Close the socket to trigger read exit
				conn.closer.Close()
				return
			case TestConnHadResponse:
				conn.SetTestConnState(TestConnIdle)
				currentTimeout = defaultTimeout
			}
		}
	}
}

// Handle a protocol packet
func (conn *Connection) HandlePacket(buffer []byte) (err error) {

	if glog.V(4) {
		glog.Infof("received %s", elvin.PacketIDString(elvin.PacketID(buffer)))
	}

	// Receiving any packet acts a ConfConn
	conn.SetTestConnState(TestConnHadResponse)

	switch elvin.PacketID(buffer) {

	// Client side packets a router shouldn't receive
	case elvin.PacketDropWarn:
	case elvin.PacketReserved:
	case elvin.PacketNotifyDeliver:
	case elvin.PacketNack:
	case elvin.PacketConnReply:
	case elvin.PacketDisconnReply:
	case elvin.PacketQuenchReply:
	case elvin.PacketSubAddNotify:
	case elvin.PacketSubModNotify:
	case elvin.PacketSubDelNotify:
	case elvin.PacketSubReply:
		return fmt.Errorf("ProtocolError: %s received", elvin.PacketIDString(elvin.PacketID(buffer)))

	// Protocol Packets not planned for the short term
	case elvin.PacketSvrRequest:
	case elvin.PacketSvrAdvt:
	case elvin.PacketSvrAdvtClose:
	case elvin.PacketClstJoinRequest:
	case elvin.PacketClstJoinReply:
	case elvin.PacketClstTerms:
	case elvin.PacketClstNotify:
	case elvin.PacketClstRedir:
	case elvin.PacketClstLeave:
	case elvin.PacketFedConnRequest:
	case elvin.PacketFedConnReply:
	case elvin.PacketFedSubReplace:
	case elvin.PacketFedNotify:
	case elvin.PacketFedSubDiff:
	case elvin.PacketFailoverConnRequest:
	case elvin.PacketFailoverConnReply:
	case elvin.PacketFailoverMaster:
	case elvin.PacketServerReport:
	case elvin.PacketServerNack:
	case elvin.PacketServerStatsReport:
		return fmt.Errorf("UnimplementedError: %s received", elvin.PacketIDString(elvin.PacketID(buffer)))
	}

	// Packets dependent upon Client's connection state
	switch conn.State() {
	case StateNew:
		// Connect and Unotify are the only valid packets without
		// a properly established connection
		switch elvin.PacketID(buffer) {
		case elvin.PacketConnRequest:
			return conn.HandleConnRequest(buffer)
		case elvin.PacketUNotify:
			return conn.HandleUNotify(buffer)
		default:
			return fmt.Errorf("ProtocolError: %s received", elvin.PacketIDString(elvin.PacketID(buffer)))
		}

	case StateConnected:
		// Deal with packets that can arrive whilst connected

		// FIXME: implement or move this lot in the short term
		switch elvin.PacketID(buffer) {
		case elvin.PacketDisconnRequest:
			return conn.HandleDisconnRequest(buffer)
		case elvin.PacketDisconn:
			return errors.New("FIXME: Packet Disconn")
		case elvin.PacketSecRequest:
			return errors.New("FIXME: Packet SecRequest")
		case elvin.PacketSecReply:
			return errors.New("FIXME: Packet SecReply")
		case elvin.PacketNotifyEmit:
			return conn.HandleNotifyEmit(buffer)
		case elvin.PacketSubAddRequest:
			return conn.HandleSubAddRequest(buffer)
		case elvin.PacketSubModRequest:
			return conn.HandleSubModRequest(buffer)
		case elvin.PacketSubDelRequest:
			return conn.HandleSubDelRequest(buffer)
		case elvin.PacketQuenchAddRequest:
			return conn.HandleQuenchAddRequest(buffer)
		case elvin.PacketQuenchModRequest:
			return conn.HandleQuenchModRequest(buffer)
		case elvin.PacketQuenchDelRequest:
			return conn.HandleQuenchDelRequest(buffer)
		case elvin.PacketTestConn:
			return conn.HandleTestConn(buffer)
		case elvin.PacketConfConn:
			// Receiving any packet acts a ConfConn so
			// already done
			// return conn.HandleConfConn(buffer)
			return nil
		case elvin.PacketAck:
			return errors.New("FIXME: Packet Ack")
		case elvin.PacketStatusUpdate:
			return errors.New("FIXME: Packet StatusUpdate")
		case elvin.PacketAuthRequest:
			return errors.New("FIXME: Packet AuthRequest")
		case elvin.PacketAuthCont:
			return errors.New("FIXME: Packet AuthCont")
		case elvin.PacketAuthAck:
			return errors.New("FIXME: Packet AuthAck")
		case elvin.PacketQosRequest:
			return errors.New("FIXME: Packet QosRequest")
		case elvin.PacketQosReply:
			return errors.New("FIXME: Packet QosReply")
		case elvin.PacketActivate:
			return errors.New("FIXME: Packet Activate")
		case elvin.PacketStandby:
			return errors.New("FIXME: Packet Standby")
		case elvin.PacketRestart:
			return errors.New("FIXME: Packet Restart")
		case elvin.PacketShutdown:
			return errors.New("FIXME: Packet Shutdown")
		default:
			return fmt.Errorf("FIXME: Packet Unknown [%d]", elvin.PacketID(buffer))
		}

	case StateDisconnecting:
	case StateClosed:
		return fmt.Errorf("ProtocolError: %s received", elvin.PacketIDString(elvin.PacketID(buffer)))
	}

	return fmt.Errorf("Error: %s received and not handled", elvin.PacketIDString(elvin.PacketID(buffer)))
}

// Handle a Connection Request
func (conn *Connection) HandleConnRequest(buffer []byte) (err error) {
	connRequest := new(elvin.ConnRequest)
	if err = connRequest.Decode(buffer); err != nil {
		conn.Close()
	}

	// Check some options
	if _, ok := connRequest.Options["TestNack"]; ok {
		if glog.V(3) {
			glog.Infof("Sending Nack for options:TestNack")
		}
		nack := new(elvin.Nack)
		nack.XID = connRequest.XID
		nack.ErrorCode = elvin.ErrorsImplementationLimit
		nack.Message = elvin.ProtocolErrors[nack.ErrorCode].Message
		nack.Args = nil
		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		conn.writeChannel <- buf
		return nil
	}
	if _, ok := connRequest.Options["TestDisconn"]; ok {
		if glog.V(3) {
			glog.Infof("Sending Disconn for options:TestDisconn")
		}
		disconn := new(elvin.Disconn)
		disconn.Reason = 4 // a little bogus
		buf := bufferPool.Get().(*bytes.Buffer)
		disconn.Encode(buf)
		conn.writeChannel <- buf
		return nil
	}

	// We're now connected
	conn.MakeID()
	conn.SetState(StateConnected)
	conn.subs = make(map[int32]*Subscription)
	conn.quenches = make(map[int32]*Quench)

	// Respond with a Connection Reply
	connReply := new(elvin.ConnReply)
	connReply.XID = connRequest.XID
	// FIXME; totally bogus
	connReply.Options = connRequest.Options

	if glog.V(3) {
		glog.Infof("New client %d connected", conn.ID())
	}

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	connReply.Encode(buf)
	conn.writeChannel <- buf

	return nil
}

// Handle a Disconnection Request
func (conn *Connection) HandleDisconnRequest(buffer []byte) (err error) {

	disconnRequest := new(elvin.DisconnRequest)
	if err = disconnRequest.Decode(buffer); err != nil {
		conn.Close()
	}

	// We're now disconnecting
	conn.SetState(StateDisconnecting)

	// Respond with a Disconnection Reply
	DisconnReply := new(elvin.DisconnReply)
	DisconnReply.XID = disconnRequest.XID

	if glog.V(3) {
		glog.Infof("client %d disconnected", conn.ID())
	}

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	DisconnReply.Encode(buf)
	conn.writeChannel <- buf

	for subID, _ := range conn.subs {
		// FIXME: send subscription removal to sub engine
		delete(conn.subs, subID)
	}

	connections.lock.Lock()
	defer connections.lock.Unlock()
	delete(connections.connections, conn.ID())

	return nil
}

// Handle a TestConn
func (conn *Connection) HandleTestConn(buffer []byte) (err error) {
	// Nothing to decode
	if glog.V(4) {
		glog.Infof("Received TestConn", conn.ID())
	}

	// Only respond is there are no queued packets
	if len(conn.writeChannel) > 1 {
		confConn := new(elvin.ConfConn)
		writeBuf := new(bytes.Buffer)
		confConn.Encode(writeBuf)
		conn.writeChannel <- writeBuf
	}

	return nil
}

// Handle a ConfConn
func (conn *Connection) HandleConfConn(buffer []byte) (err error) {
	// Note: This is never called as it's done
	// in HandlePacket as any Packet acts as a ConfConn
	conn.SetTestConnState(TestConnHadResponse)

	return nil
}

// Handle a NotifyEmit
func (conn *Connection) HandleNotifyEmit(buffer []byte) (err error) {
	nfn := new(elvin.NotifyEmit)
	err = nfn.Decode(buffer)

	// FIXME: NotifyDeliver

	// As a dummy for now we're going to send every message we see
	// to every subscription as if all evaluate to true. Don't
	// worry about the giant lock - this all goes away
	connections.lock.Lock()
	defer connections.lock.Unlock()
	nd := new(elvin.NotifyDeliver)
	nd.NameValue = nfn.NameValue

	for connid, connection := range connections.connections {
		if len(connection.subs) > 0 {
			nd.Insecure = make([]int64, len(connection.subs))
			i := 0
			for id, _ := range connection.subs {
				nd.Insecure[i] = int64(connid)<<32 | int64(id)
				i++
			}
			buf := bufferPool.Get().(*bytes.Buffer)
			nd.Encode(buf)
			connection.writeChannel <- buf
		}
	}
	return nil
}

// Handle a UNotify
func (conn *Connection) HandleUNotify(buffer []byte) (err error) {
	nfn := new(elvin.UNotify)
	err = nfn.Decode(buffer)

	// FIXME: NotifyDeliver

	// As a dummy for now we're going to send every message we see
	// to every subscription as if all evaluate to true. Don't
	// worry about the giant lock - this all goes away
	connections.lock.Lock()
	defer connections.lock.Unlock()
	nd := new(elvin.NotifyDeliver)
	nd.NameValue = nfn.NameValue

	for connid, connection := range connections.connections {
		if len(connection.subs) > 0 {
			nd.Insecure = make([]int64, len(connection.subs))
			i := 0
			for id, _ := range connection.subs {
				nd.Insecure[i] = int64(connid)<<32 | int64(id)
				i++
			}
			buf := bufferPool.Get().(*bytes.Buffer)
			nd.Encode(buf)
			connection.writeChannel <- buf
		}
	}
	return nil
}

// Handle a Subscription Add
func (conn *Connection) HandleSubAddRequest(buffer []byte) (err error) {
	subRequest := new(elvin.SubAddRequest)
	err = subRequest.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
	}

	ast, nack := Parse(subRequest.Expression)
	if nack != nil {
		nack.XID = subRequest.XID
		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		conn.writeChannel <- buf
		return nil
	}

	// Create a subscription and add it to the subscription store
	var sub Subscription
	sub.Ast = ast
	sub.AcceptInsecure = subRequest.AcceptInsecure
	sub.Keys = subRequest.Keys

	// Create a unique sub id
	var s int32 = rand.Int31()
	for {
		_, err := conn.subs[s]
		if !err {
			break
		}
		s++
	}
	conn.subs[s] = &sub
	sub.SubID = (int64(conn.ID()) << 32) | int64(s)

	// FIXME: send subscription addition to sub engine

	// Respond with a SubReply
	subReply := new(elvin.SubReply)
	subReply.XID = subRequest.XID
	subReply.SubID = sub.SubID
	if glog.V(4) {
		glog.Infof("Connection:%d New subscription:%d (%d)", conn.ID(), s, sub.SubID)
	}

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	subReply.Encode(buf)
	conn.writeChannel <- buf
	return nil
}

// Handle a Subscription Delete
func (conn *Connection) HandleSubDelRequest(buffer []byte) (err error) {
	subDelRequest := new(elvin.SubDelRequest)
	err = subDelRequest.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
	}

	// If deletion fails then nack and disconn
	idx := int32(subDelRequest.SubID & 0xfffffffff)
	_, exists := conn.subs[idx]
	if !exists {
		nack := new(elvin.Nack)
		nack.XID = subDelRequest.XID
		nack.ErrorCode = elvin.ErrorsUnknownSubID
		nack.Message = elvin.ProtocolErrors[nack.ErrorCode].Message
		nack.Args = make([]interface{}, 1)
		nack.Args[0] = subDelRequest.SubID
		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		conn.writeChannel <- buf

		// FIXME Disconnect as that's a protocol violation
		return nil
	}

	// Remove it from the connection
	delete(conn.subs, idx)

	// Respond with a SubReply
	subReply := new(elvin.SubReply)
	subReply.XID = subDelRequest.XID
	subReply.SubID = subDelRequest.SubID

	// FIXME: send subscription deletion to sub engine

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	subReply.Encode(buf)
	conn.writeChannel <- buf
	return nil
}

func (conn *Connection) HandleSubModRequest(buffer []byte) (err error) {
	subModRequest := new(elvin.SubModRequest)
	err = subModRequest.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
		return err
	}

	// If modify fails then nack and disconn
	idx := int32(subModRequest.SubID & 0xfffffffff)
	sub, exists := conn.subs[idx]
	if !exists {
		nack := new(elvin.Nack)
		nack.XID = subModRequest.XID
		nack.ErrorCode = elvin.ErrorsUnknownSubID
		nack.Message = elvin.ProtocolErrors[nack.ErrorCode].Message
		nack.Args = make([]interface{}, 1)
		nack.Args[0] = subModRequest.SubID

		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		conn.writeChannel <- buf

		// FIXME Disconnect if that's a repeated protocol violation?
		return nil
	}

	// FIXME: At some point this sub will need a lock but for now it's all handled in the read goroutine
	// FIXME: And any update to the sub should be all or nothing

	// Check the subscription expression. Empty is ok. Incorrect means bail.
	if len(subModRequest.Expression) > 0 {
		ast, nack := Parse(subModRequest.Expression)
		if nack != nil {
			nack.XID = subModRequest.XID
			buf := bufferPool.Get().(*bytes.Buffer)
			nack.Encode(buf)
			conn.writeChannel <- buf
			return nil
		}
		sub.Ast = ast
	}

	// AcceptInsecure is the only piece that must have a value - and it is allowed to be the same
	sub.AcceptInsecure = subModRequest.AcceptInsecure

	// We ignore deletion requests that don't exist
	if len(subModRequest.DelKeys) > 0 {
		// FIXME: implement
	}

	// We ignore addition requests that already exist
	if len(subModRequest.AddKeys) > 0 {
		// FIXME: implement
	}

	// Respond with a SubReply
	subReply := new(elvin.SubReply)
	subReply.XID = subModRequest.XID
	subReply.SubID = subModRequest.SubID

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	subReply.Encode(buf)
	conn.writeChannel <- buf
	return nil
}

// Handle a Quench Add
func (conn *Connection) HandleQuenchAddRequest(buffer []byte) (err error) {
	quenchRequest := new(elvin.QuenchAddRequest)
	err = quenchRequest.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
	}

	// FIXME: what checking do we need to do here

	// Create a quench and add it to the quench store
	var quench Quench
	quench.Names = make(map[string]bool)
	for name, _ := range quenchRequest.Names {
		quench.Names[name] = true
	}
	quench.DeliverInsecure = quenchRequest.DeliverInsecure
	quench.Keys = quenchRequest.Keys

	// Create a unique quench id
	var q int32 = rand.Int31()
	for {
		_, err := conn.quenches[q]
		if !err {
			break
		}
		q++
	}
	conn.quenches[q] = &quench
	quench.QuenchID = (int64(conn.ID()) << 32) | int64(q)

	// FIXME: send quench to sub engine

	// Respond with a QuenchReply
	quenchReply := new(elvin.QuenchReply)
	quenchReply.XID = quenchRequest.XID
	quenchReply.QuenchID = quench.QuenchID

	if glog.V(4) {
		glog.Infof("Connection:%d New quench:%d %+v", conn.ID(), quench.QuenchID, quench)
	}
	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	quenchReply.Encode(buf)
	conn.writeChannel <- buf
	return nil
}

func (conn *Connection) HandleQuenchModRequest(buffer []byte) (err error) {
	quenchModRequest := new(elvin.QuenchModRequest)
	err = quenchModRequest.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
	}

	// If modify fails then nack and disconn
	idx := int32(quenchModRequest.QuenchID & 0xfffffffff)
	quench, exists := conn.quenches[idx]
	if !exists {
		nack := new(elvin.Nack)
		nack.XID = quenchModRequest.XID
		nack.ErrorCode = elvin.ErrorsUnknownQuenchID
		nack.Message = elvin.ProtocolErrors[nack.ErrorCode].Message
		nack.Args = make([]interface{}, 1)
		nack.Args[0] = quenchModRequest.QuenchID

		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		conn.writeChannel <- buf

		// FIXME Disconnect if that's a repeated protocol violation?
		return nil
	}

	for name, _ := range quenchModRequest.AddNames {
		quench.Names[name] = true
	}
	for name, _ := range quenchModRequest.DelNames {
		delete(quench.Names, name)
	}
	quench.DeliverInsecure = quenchModRequest.DeliverInsecure
	// FIXME: implement key changes

	// FIXME: Pass on change to engine

	// Respond with a QuenchReply
	quenchReply := new(elvin.QuenchReply)
	quenchReply.XID = quenchModRequest.XID
	quenchReply.QuenchID = quench.QuenchID

	if glog.V(4) {
		glog.Infof("Connection:%d  quench:%d modified %+v", conn.ID(), quench.QuenchID, quench)
	}

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	quenchReply.Encode(buf)
	conn.writeChannel <- buf
	return nil
}

func (conn *Connection) HandleQuenchDelRequest(buffer []byte) (err error) {
	quenchDelRequest := new(elvin.QuenchDelRequest)
	err = quenchDelRequest.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
	}

	// If deletion fails then nack and disconn
	idx := int32(quenchDelRequest.QuenchID & 0xfffffffff)
	_, exists := conn.quenches[idx]
	if !exists {
		nack := new(elvin.Nack)
		nack.XID = quenchDelRequest.XID
		nack.ErrorCode = elvin.ErrorsUnknownQuenchID
		nack.Message = elvin.ProtocolErrors[nack.ErrorCode].Message
		nack.Args = make([]interface{}, 1)
		nack.Args[0] = quenchDelRequest.QuenchID
		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		conn.writeChannel <- buf

		// FIXME Disconnect as that's a protocol violation
		return nil
	}

	// Remove it from the connection
	delete(conn.quenches, idx)

	// Respond with a QuenchReply
	quenchReply := new(elvin.QuenchReply)
	quenchReply.XID = quenchDelRequest.XID
	quenchReply.QuenchID = quenchDelRequest.QuenchID

	// FIXME: send quench deletion to sub engine

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	quenchReply.Encode(buf)
	conn.writeChannel <- buf
	return nil
}
