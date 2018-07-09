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
	"github.com/cobaro/elvin/elog"
	"github.com/cobaro/elvin/elvin"
	"io"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Client States
const (
	StateNew = iota
	StateConnected
	StateDisconnecting
	StateClosed
)

// Return state (synchronized)
func (client *Client) State() int {
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.state
}

// Get state (synchronized)
func (client *Client) SetState(state int) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.state = state
}

// TestConn/ConfConn Timeout States
const (
	TestConnIdle = iota
	TestConnAwaitingResponse
	TestConnHadResponse
)

// Set TestConn/ConfConn Timeout (synchronized)
func (client *Client) TestConnState() int {
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.testConnState
}

// Get TestConn/ConfConn Timeout (synchronized)
func (client *Client) SetTestConnState(state int) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.testConnState = state
}

// A Client (e.g. a socket)
type Client struct {
	id             int32 // Id assigned by router
	mu             sync.Mutex
	elog           elog.Elog
	channels       ClientChannels
	subs           map[int32]*Subscription
	quenches       map[int32]*Quench
	reader         io.Reader
	writer         io.Writer
	closer         io.Closer
	state          int
	testConnState  int
	keysNfn        elvin.KeyBlock
	keysSub        elvin.KeyBlock
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

// Return our unique 32 bit unsigned identifier
func (client *Client) ID() int32 {
	return client.id
}

func (client *Client) Close() {
	client.elog.Logf(elog.LogLevelInfo2, "Closing client %d", client.ID())
	select {
	case client.writeTerminate <- 1:
	default:
	}
	client.closer.Close()
	client.channels.remove <- client.ID()

}

// Read n bytes from reader into buffer which must be big enough
func readBytes(reader io.Reader, buffer []byte, numToRead int) (int, error) {
	offset := 0
	for offset < numToRead {
		length, err := reader.Read(buffer[offset:numToRead])
		if err != nil {
			return offset + length, err
		}
		offset += length
	}
	return offset, nil
}

// Handle reading for now run as a goroutine
func (client *Client) readHandler() {
	client.elog.Logf(elog.LogLevelDebug1, "Read Handler starting")
	defer client.elog.Logf(elog.LogLevelDebug1, "Read Handler exiting")

	header := make([]byte, 4)

	for {
		// We reallocate each time as decoding
		// takes slices out of it
		buffer := make([]byte, 2048)

		// Read frame header
		length, err := readBytes(client.reader, header, 4)
		if length != 4 || err != nil {
			break // We're done
		}

		// Read the protocol packet, starting with it's length
		packetSize := int(binary.BigEndian.Uint32(header))
		// Grow our buffer if needed
		if packetSize > len(buffer) {
			client.elog.Logf(elog.LogLevelDebug2, "Growing buffer to %d bytes", packetSize)
			buffer = make([]byte, packetSize)
		}

		length, err = readBytes(client.reader, buffer, packetSize)
		if length != packetSize || err != nil {
			break // We're done
		}

		// Deal with the packet
		if err = client.HandlePacket(buffer); err != nil {
			client.elog.Logf(elog.LogLevelError, "Read Handler error: %v", err)
			// FIXME: protocol error
			break
		}
	}
	client.Close()
}

// Handle writing for now run as a goroutine
func (client *Client) writeHandler() {
	client.elog.Logf(elog.LogLevelDebug1, "Write Handler starting")
	defer client.elog.Logf(elog.LogLevelDebug1, "Write Handler exiting")

	header := make([]byte, 4)

	// TestConn and ConfConn use two timers:
	//
	//   The first, intervalTimer, is  how long a client
	//   can be idle before we send a TestConn and it's configured
	//   via the configuration option TestConnInterval.
	//   A value of zero means disabled so we simply set it up for
	//   a long time as way of keeping the select loop simple.
	//
	//   The second testConnTimeout is how long we should wait for a
	//   response.
	defaultTimeout := client.testConnInterval
	if defaultTimeout == 0 {
		defaultTimeout = math.MaxInt64
	}
	currentTimeout := defaultTimeout
	client.SetTestConnState(TestConnIdle)

	// Our write loop waits for data to write and sends it
	// It can be terminated via the writeTerminate channel
	// It runs a Test/ConfConn timer if configured
	for {
		select {
		case buffer := <-client.writeChannel:

			// Write the frame header (packetsize)
			binary.BigEndian.PutUint32(header, uint32(buffer.Len()))
			_, err := client.writer.Write(header)
			if err != nil {
				// Deal with more errors
				if err != io.EOF {
					client.elog.Logf(elog.LogLevelError, "Unexpected write error: %v", err)
				}
				bufferPool.Put(buffer)
				return // We're done, cleanup done by read
			}

			// Write the packet
			_, err = buffer.WriteTo(client.writer)
			if err != nil {
				// Deal with more errors
				if err != io.EOF {
					client.elog.Logf(elog.LogLevelError, "Unexpected write error: %v", err)
				}
				bufferPool.Put(buffer)
				return // We're done, cleanup done by read
			}
		case <-client.writeTerminate:
			return // We're done, cleanup done by read

		case <-time.After(currentTimeout):
			client.elog.Logf(elog.LogLevelDebug3, "writeHandler timeout: %d", client.TestConnState())

			switch client.TestConnState() {
			case TestConnIdle:
				client.SetTestConnState(TestConnAwaitingResponse)
				currentTimeout = client.testConnTimeout
				testConn := new(elvin.TestConn)
				writeBuf := new(bytes.Buffer)
				testConn.Encode(writeBuf)
				client.writeChannel <- writeBuf
			case TestConnAwaitingResponse:
				client.elog.Logf(elog.LogLevelInfo1, "Closing client %d for not responding to TestConn", client.ID())
				// FIXME:Close the socket to trigger read exit
				client.closer.Close()
				return
			case TestConnHadResponse:
				client.SetTestConnState(TestConnIdle)
				currentTimeout = defaultTimeout
			}
		}
	}
}

// Handle a protocol packet
func (client *Client) HandlePacket(buffer []byte) (err error) {

	client.elog.Logf(elog.LogLevelDebug3, "received %s", elvin.PacketIDString(elvin.PacketID(buffer)))

	// Receiving any packet acts a ConfConn
	client.SetTestConnState(TestConnHadResponse)

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

	// Packets dependent upon Client's client state
	switch client.State() {
	case StateNew:
		// Connect and Unotify are the only valid packets without
		// a properly established client
		switch elvin.PacketID(buffer) {
		case elvin.PacketConnRequest:
			return client.HandleConnRequest(buffer)
		case elvin.PacketUNotify:
			return client.HandleUNotify(buffer)
		default:
			return fmt.Errorf("ProtocolError: %s received", elvin.PacketIDString(elvin.PacketID(buffer)))
		}

	case StateConnected:
		// Deal with packets that can arrive whilst connected

		// FIXME: implement or move this lot in the short term
		switch elvin.PacketID(buffer) {
		case elvin.PacketDisconnRequest:
			return client.HandleDisconnRequest(buffer)
		case elvin.PacketDisconn:
			return errors.New("FIXME: Packet Disconn")
		case elvin.PacketSecRequest:
			return errors.New("FIXME: Packet SecRequest")
		case elvin.PacketSecReply:
			return errors.New("FIXME: Packet SecReply")
		case elvin.PacketNotifyEmit:
			return client.HandleNotifyEmit(buffer)
		case elvin.PacketSubAddRequest:
			return client.HandleSubAddRequest(buffer)
		case elvin.PacketSubModRequest:
			return client.HandleSubModRequest(buffer)
		case elvin.PacketSubDelRequest:
			return client.HandleSubDelRequest(buffer)
		case elvin.PacketQuenchAddRequest:
			return client.HandleQuenchAddRequest(buffer)
		case elvin.PacketQuenchModRequest:
			return client.HandleQuenchModRequest(buffer)
		case elvin.PacketQuenchDelRequest:
			return client.HandleQuenchDelRequest(buffer)
		case elvin.PacketTestConn:
			return client.HandleTestConn(buffer)
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

// Handle a Client Request
func (client *Client) HandleConnRequest(buffer []byte) (err error) {
	connRequest := new(elvin.ConnRequest)
	if err = connRequest.Decode(buffer); err != nil {
		client.Close()
	}

	// Check some options
	if _, ok := connRequest.Options["TestNack"]; ok {
		client.elog.Logf(elog.LogLevelInfo1, "Sending Nack for options:TestNack")
		nack := new(elvin.Nack)
		nack.XID = connRequest.XID
		nack.ErrorCode = elvin.ErrorsImplementationLimit
		nack.Message = elvin.ProtocolErrors[nack.ErrorCode].Message
		nack.Args = nil
		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		client.writeChannel <- buf
		return nil
	}
	if _, ok := connRequest.Options["TestDisconn"]; ok {
		client.elog.Logf(elog.LogLevelInfo1, "Sending Disconn for options:TestDisconn")
		disconn := new(elvin.Disconn)
		disconn.Reason = 4 // a little bogus
		buf := bufferPool.Get().(*bytes.Buffer)
		disconn.Encode(buf)
		client.writeChannel <- buf
		return nil
	}

	// We're now connected
	client.SetState(StateConnected)
	client.subs = make(map[int32]*Subscription)
	client.quenches = make(map[int32]*Quench)

	// Prime any keys if they gave us some
	client.keysNfn = connRequest.KeysNfn
	PrimeProducer(client.keysNfn)
	client.keysSub = connRequest.KeysSub
	PrimeConsumer(client.keysSub)

	// Respond with a ConnReply
	connReply := new(elvin.ConnReply)
	connReply.XID = connRequest.XID
	// FIXME; totally bogus
	connReply.Options = connRequest.Options

	client.elog.Logf(elog.LogLevelInfo1, "New client %d connected", client.ID())

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	connReply.Encode(buf)
	client.writeChannel <- buf

	return nil
}

// Handle a Disclient Request
func (client *Client) HandleDisconnRequest(buffer []byte) (err error) {

	disconnRequest := new(elvin.DisconnRequest)
	if err = disconnRequest.Decode(buffer); err != nil {
		client.Close()
	}

	// We're now disconnecting
	client.SetState(StateDisconnecting)

	// Respond with a Disclient Reply
	DisconnReply := new(elvin.DisconnReply)
	DisconnReply.XID = disconnRequest.XID

	client.elog.Logf(elog.LogLevelInfo1, "client %d disconnected", client.ID())

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	DisconnReply.Encode(buf)
	client.writeChannel <- buf

	for subID, _ := range client.subs {
		// FIXME: send subscription removal to sub engine
		delete(client.subs, subID)
	}

	client.channels.remove <- client.ID()

	return nil
}

// Handle a TestConn
func (client *Client) HandleTestConn(buffer []byte) (err error) {
	// Nothing to decode
	client.elog.Logf(elog.LogLevelInfo2, "Received TestConn", client.ID())

	// Only respond is there are no queued packets
	if len(client.writeChannel) > 1 {
		confConn := new(elvin.ConfConn)
		writeBuf := new(bytes.Buffer)
		confConn.Encode(writeBuf)
		client.writeChannel <- writeBuf
	}

	return nil
}

// Handle a ConfConn
func (client *Client) HandleConfConn(buffer []byte) (err error) {
	// Note: This is never called as it's done
	// in HandlePacket as any Packet acts as a ConfConn
	client.SetTestConnState(TestConnHadResponse)

	return nil
}

// Handle a NotifyEmit
func (client *Client) HandleNotifyEmit(buffer []byte) (err error) {
	ne := new(elvin.NotifyEmit)
	if err = ne.Decode(buffer); err != nil {
		return err
	}

	client.channels.notify <- Notification{client.keysNfn, ne.NameValue, ne.DeliverInsecure, ne.Keys}
	return nil
}

// Handle a UNotify
func (client *Client) HandleUNotify(buffer []byte) (err error) {
	unotify := new(elvin.UNotify)
	if err = unotify.Decode(buffer); err != nil {
		return err
	}

	// FIXME: Check version and ?

	client.channels.notify <- Notification{client.keysNfn, unotify.NameValue, unotify.DeliverInsecure, unotify.Keys}
	return nil
}

// Handle a Subscription Add
func (client *Client) HandleSubAddRequest(buffer []byte) (err error) {
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
		client.writeChannel <- buf
		return nil
	}

	// Create a subscription and add it to the subscription store
	var sub Subscription
	sub.Ast = ast
	sub.AcceptInsecure = subRequest.AcceptInsecure
	sub.Keys = subRequest.Keys
	PrimeConsumer(sub.Keys)

	// Create a unique sub id
	var s int32 = rand.Int31()
	for {
		_, err := client.subs[s]
		if !err {
			break
		}
		s++
	}
	client.subs[s] = &sub
	sub.SubID = (int64(client.ID()) << 32) | int64(s)

	client.channels.subAdd <- &sub

	// Respond with a SubReply
	subReply := new(elvin.SubReply)
	subReply.XID = subRequest.XID
	subReply.SubID = sub.SubID
	client.elog.Logf(elog.LogLevelInfo2, "Client:%d New subscription:%d (%d)", client.ID(), s, sub.SubID)

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	subReply.Encode(buf)
	client.writeChannel <- buf
	return nil
}

// Handle a Subscription Delete
func (client *Client) HandleSubDelRequest(buffer []byte) (err error) {
	subDelRequest := new(elvin.SubDelRequest)
	err = subDelRequest.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
	}

	// If deletion fails then nack and disconn
	idx := int32(subDelRequest.SubID & 0xfffffffff)
	sub, exists := client.subs[idx]
	if !exists {
		nack := new(elvin.Nack)
		nack.XID = subDelRequest.XID
		nack.ErrorCode = elvin.ErrorsUnknownSubID
		nack.Message = elvin.ProtocolErrors[nack.ErrorCode].Message
		nack.Args = make([]interface{}, 1)
		nack.Args[0] = subDelRequest.SubID
		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		client.writeChannel <- buf

		// FIXME Disconnect as that's a protocol violation
		return nil
	}

	// Remove it from the client
	delete(client.subs, idx)

	// Send it to the subscription engine
	client.channels.subDel <- sub

	// Respond with a SubReply
	subReply := new(elvin.SubReply)
	subReply.XID = subDelRequest.XID
	subReply.SubID = subDelRequest.SubID

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	subReply.Encode(buf)
	client.writeChannel <- buf
	return nil
}

func (client *Client) HandleSubModRequest(buffer []byte) (err error) {
	subModRequest := new(elvin.SubModRequest)
	err = subModRequest.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
		return err
	}

	// If modify fails then nack and disconn
	idx := int32(subModRequest.SubID & 0xfffffffff)
	sub, exists := client.subs[idx]
	if !exists {
		nack := new(elvin.Nack)
		nack.XID = subModRequest.XID
		nack.ErrorCode = elvin.ErrorsUnknownSubID
		nack.Message = elvin.ProtocolErrors[nack.ErrorCode].Message
		nack.Args = make([]interface{}, 1)
		nack.Args[0] = subModRequest.SubID

		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		client.writeChannel <- buf

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
			client.writeChannel <- buf
			return nil
		}
		sub.Ast = ast
	}

	// AcceptInsecure is the only piece that must have a value - and it is allowed to be the same
	sub.AcceptInsecure = subModRequest.AcceptInsecure

	// Merge in any new keys
	if len(subModRequest.AddKeys) > 0 {
		PrimeConsumer(subModRequest.AddKeys)
		elvin.KeyBlockAddKeys(sub.Keys, subModRequest.AddKeys)
	}

	// Remove any old keys
	if len(subModRequest.DelKeys) > 0 {
		PrimeConsumer(subModRequest.DelKeys)
		elvin.KeyBlockDeleteKeys(sub.Keys, subModRequest.DelKeys)
	}

	// Send it to the subscription engine
	client.channels.subMod <- sub

	// Respond with a SubReply
	subReply := new(elvin.SubReply)
	subReply.XID = subModRequest.XID
	subReply.SubID = subModRequest.SubID

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	subReply.Encode(buf)
	client.writeChannel <- buf
	return nil
}

// Handle a Quench Add
func (client *Client) HandleQuenchAddRequest(buffer []byte) (err error) {
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
		_, err := client.quenches[q]
		if !err {
			break
		}
		q++
	}
	client.quenches[q] = &quench
	quench.QuenchID = (int64(client.ID()) << 32) | int64(q)

	// send quench to sub engine
	client.channels.quenchAdd <- &quench

	// Respond with a QuenchReply
	quenchReply := new(elvin.QuenchReply)
	quenchReply.XID = quenchRequest.XID
	quenchReply.QuenchID = quench.QuenchID

	client.elog.Logf(elog.LogLevelInfo2, "Client:%d New quench:%d %+v", client.ID(), quench.QuenchID, quench)
	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	quenchReply.Encode(buf)
	client.writeChannel <- buf
	return nil
}

func (client *Client) HandleQuenchModRequest(buffer []byte) (err error) {
	quenchModRequest := new(elvin.QuenchModRequest)
	err = quenchModRequest.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
	}

	// If modify fails then nack and disconn
	idx := int32(quenchModRequest.QuenchID & 0xfffffffff)
	quench, exists := client.quenches[idx]
	if !exists {
		nack := new(elvin.Nack)
		nack.XID = quenchModRequest.XID
		nack.ErrorCode = elvin.ErrorsUnknownQuenchID
		nack.Message = elvin.ProtocolErrors[nack.ErrorCode].Message
		nack.Args = make([]interface{}, 1)
		nack.Args[0] = quenchModRequest.QuenchID

		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		client.writeChannel <- buf

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

	// send quench to sub engine
	client.channels.quenchMod <- quench

	// Respond with a QuenchReply
	quenchReply := new(elvin.QuenchReply)
	quenchReply.XID = quenchModRequest.XID
	quenchReply.QuenchID = quench.QuenchID

	client.elog.Logf(elog.LogLevelInfo2, "Client:%d  quench:%d modified %+v", client.ID(), quench.QuenchID, quench)

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	quenchReply.Encode(buf)
	client.writeChannel <- buf
	return nil
}

func (client *Client) HandleQuenchDelRequest(buffer []byte) (err error) {
	quenchDelRequest := new(elvin.QuenchDelRequest)
	err = quenchDelRequest.Decode(buffer)
	if err != nil {
		// FIXME: Protocol violation
	}

	// If deletion fails then nack and disconn
	idx := int32(quenchDelRequest.QuenchID & 0xfffffffff)
	quench, exists := client.quenches[idx]
	if !exists {
		nack := new(elvin.Nack)
		nack.XID = quenchDelRequest.XID
		nack.ErrorCode = elvin.ErrorsUnknownQuenchID
		nack.Message = elvin.ProtocolErrors[nack.ErrorCode].Message
		nack.Args = make([]interface{}, 1)
		nack.Args[0] = quenchDelRequest.QuenchID
		buf := bufferPool.Get().(*bytes.Buffer)
		nack.Encode(buf)
		client.writeChannel <- buf

		// FIXME Disconnect as that's a protocol violation
		return nil
	}

	// Remove it from the client
	delete(client.quenches, idx)

	// send quench to sub engine
	client.channels.quenchDel <- quench

	// Respond with a QuenchReply
	quenchReply := new(elvin.QuenchReply)
	quenchReply.XID = quenchDelRequest.XID
	quenchReply.QuenchID = quenchDelRequest.QuenchID

	// Encode that into a buffer for the write handler
	buf := bufferPool.Get().(*bytes.Buffer)
	quenchReply.Encode(buf)
	client.writeChannel <- buf
	return nil
}
