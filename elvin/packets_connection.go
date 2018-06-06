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

package elvin

import (
	"bytes"
	"fmt"
)

// reasons for disconnection (Disconn packet)
const (
	// Defined in the Elvin spec
	DisconnReasonReserved             = 0
	DisconnReasonRouterShuttingDown   = 1
	DisconnReasonRouterRedirect       = 2
	DisconnReasonRouterProtocolErrors = 4

	// Local to client library
	DisconnReasonClientConnectionLost = 100 // e.g., econnlost
	DisconnReasonClientProtocolErrors = 101 // e.g., packet decoding failed
)

// Packet: Connection Request
type ConnRequest struct {
	XID          uint32
	VersionMajor uint32
	VersionMinor uint32
	Options      map[string]interface{}
	KeysNfn      []Keyset
	KeysSub      []Keyset
}

// Integer value of packet type
func (pkt *ConnRequest) ID() int {
	return PacketConnRequest
}

// String representation of packet type
func (pkt *ConnRequest) IDString() string {
	return "ConnRequest"
}

// Pretty print with indent
func (pkt *ConnRequest) IString(indent string) string {
	return fmt.Sprintf(
		"%sXID: %d\n"+
			"%sVersionMajor %d\n"+
			"%sVersionMinor %d\n"+
			"%sOptions %v\n"+
			"%sKeysNfn: %v\n"+
			"%sKeysSub: %v\n",
		indent, pkt.XID,
		indent, pkt.VersionMajor,
		indent, pkt.VersionMinor,
		indent, pkt.Options,
		indent, pkt.KeysNfn,
		indent, pkt.KeysSub)
}

// Pretty print without indent so generic ToString() works
func (pkt *ConnRequest) String() string {
	return pkt.IString("")
}

// Decode a ConnRequest packet from a byte array
func (pkt *ConnRequest) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.VersionMajor, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.VersionMinor, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Options, used, err = XdrGetNotification(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.KeysNfn, used, err = XdrGetKeys(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	if pkt.KeysSub, used, err = XdrGetKeys(bytes[offset:]); err != nil {
		return err
	}
	offset += used

	return nil
}

func (pkt *ConnRequest) Encode(buffer *bytes.Buffer) {
	// FIXME: error handling
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutUint32(buffer, pkt.XID)
	XdrPutUint32(buffer, pkt.VersionMajor)
	XdrPutUint32(buffer, pkt.VersionMinor)
	XdrPutNotification(buffer, pkt.Options)
	XdrPutKeys(buffer, pkt.KeysNfn)
	XdrPutKeys(buffer, pkt.KeysSub)
}

// Packet: Connection Reply
type ConnReply struct {
	XID     uint32
	Options map[string]interface{}
}

// Integer value of packet type
func (pkt *ConnReply) ID() int {
	return PacketConnReply
}

// String representation of packet type
func (pkt *ConnReply) IDString() string {
	return "ConnReply"
}

// Pretty print with indent
func (pkt *ConnReply) IString(indent string) string {
	return fmt.Sprintf("%sXID: %d\n%sOptions %v\n",
		indent, pkt.XID,
		indent, pkt.Options)

}

// Pretty print without indent so generic ToString() works
func (pkt *ConnReply) String() string {
	return pkt.IString("")
}

// Decode a ConnReply packet from a byte array
func (pkt *ConnReply) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Options, used, err = XdrGetNotification(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

// Encode a ConnReply from a buffer
func (pkt *ConnReply) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutUint32(buffer, pkt.XID)
	XdrPutNotification(buffer, pkt.Options)
}

// Packet: Disconnection Request
type DisconnRequest struct {
	XID uint32
}

// Integer value of packet type
func (pkt *DisconnRequest) ID() int {
	return PacketDisconnRequest
}

// String representation of packet type
func (pkt *DisconnRequest) IDString() string {
	return "DisconnRequest"
}

// Pretty print with indent
func (pkt *DisconnRequest) IString(indent string) string {
	return fmt.Sprintf(
		"%sXID: %d\n",
		indent, pkt.XID)
}

// Pretty print without indent so generic ToString() works
func (pkt *DisconnRequest) String() string {
	return pkt.IString("")
}

// Decode a DisconnRequest packet from a byte array
func (pkt *DisconnRequest) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

func (pkt *DisconnRequest) Encode(buffer *bytes.Buffer) {
	// FIXME: error handling
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutUint32(buffer, pkt.XID)
}

// Packet: Disconnection Request
type DisconnReply struct {
	XID uint32
}

// Integer value of packet type
func (pkt *DisconnReply) ID() int {
	return PacketDisconnReply
}

// String representation of packet type
func (pkt *DisconnReply) IDString() string {
	return "DisconnReply"
}

// Pretty print with indent
func (pkt *DisconnReply) IString(indent string) string {
	return fmt.Sprintf(
		"%sXID: %d\n",
		indent, pkt.XID)
}

// Pretty print without indent so generic ToString() works
func (pkt *DisconnReply) String() string {
	return pkt.IString("")
}

// Decode a DisconnReply packet from a byte array
func (pkt *DisconnReply) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

func (pkt *DisconnReply) Encode(buffer *bytes.Buffer) {
	// FIXME: error handling
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutUint32(buffer, pkt.XID)
}

// Packet: Disconn
type Disconn struct {
	Reason uint32
	Args   string
}

// Integer value of packet type
func (pkt *Disconn) ID() int {
	return PacketDisconn
}

// String representation of packet type
func (pkt *Disconn) IDString() string {
	return "Disconn"
}

// Pretty print with indent
func (pkt *Disconn) IString(indent string) string {
	return fmt.Sprintf(
		"%sReason: %d\n%sArgs: %s\n",
		indent, pkt.Reason,
		indent, pkt.Args)
}

// Pretty print without indent so generic ToString() works
func (pkt *Disconn) String() string {
	return pkt.IString("")
}

// Decode a Disconn packet from a byte array
func (pkt *Disconn) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.Reason, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Args, used, err = XdrGetString(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

func (pkt *Disconn) Encode(buffer *bytes.Buffer) {
	// FIXME: error handling
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutUint32(buffer, pkt.Reason)
	XdrPutString(buffer, pkt.Args)
}
