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

// Packet: Connection Request
type ConnRqst struct {
	Xid          uint32
	VersionMajor uint32
	VersionMinor uint32
	Options      map[string]interface{}
	KeysNfn      []Keyset
	KeysSub      []Keyset
}

// Integer value of packet type
func (pkt *ConnRqst) Id() int {
	return PacketConnRqst
}

// String representation of packet type
func (pkt *ConnRqst) IdString() string {
	return "ConnRqst"
}

// Pretty print with indent
func (pkt *ConnRqst) IString(indent string) string {
	return fmt.Sprintf(
		"%sXid: %d\n"+
			"%sVersionMajor %d\n"+
			"%sVersionMinor %d\n"+
			"%sOptions %v\n"+
			"%sKeysNfn: %v\n"+
			"%sKeysSub: %v\n",
		indent, pkt.Xid,
		indent, pkt.VersionMajor,
		indent, pkt.VersionMinor,
		indent, pkt.Options,
		indent, pkt.KeysNfn,
		indent, pkt.KeysSub)
}

// Pretty print without indent so generic ToString() works
func (pkt *ConnRqst) String() string {
	return pkt.IString("")
}

// Decode a ConnRqst packet from a byte array
func (pkt *ConnRqst) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.Xid, used = XdrGetUint32(bytes[offset:])
	offset += used

	pkt.VersionMajor, used = XdrGetUint32(bytes[offset:])
	offset += used

	pkt.VersionMinor, used = XdrGetUint32(bytes[offset:])
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

func (pkt *ConnRqst) Encode(buffer *bytes.Buffer) {
	// FIXME: error handling
	XdrPutInt32(buffer, pkt.Id())
	XdrPutUint32(buffer, pkt.Xid)
	XdrPutUint32(buffer, pkt.VersionMajor)
	XdrPutUint32(buffer, pkt.VersionMinor)
	XdrPutNotification(buffer, pkt.Options)
	XdrPutKeys(buffer, pkt.KeysNfn)
	XdrPutKeys(buffer, pkt.KeysSub)
}

// Packet: Connection Reply
type ConnRply struct {
	Xid     uint32
	Options map[string]interface{}
}

// Integer value of packet type
func (pkt *ConnRply) Id() int {
	return PacketConnRply
}

// String representation of packet type
func (pkt *ConnRply) IdString() string {
	return "ConnRply"
}

// Pretty print with indent
func (pkt *ConnRply) IString(indent string) string {
	return fmt.Sprintf("%sXid: %d\n%sOptions %v\n",
		indent, pkt.Xid,
		indent, pkt.Options)

}

// Pretty print without indent so generic ToString() works
func (pkt *ConnRply) String() string {
	return pkt.IString("")
}

// Decode a ConnRply packet from a byte array
func (pkt *ConnRply) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.Xid, used = XdrGetUint32(bytes[offset:])
	offset += used

	pkt.Options, used, err = XdrGetNotification(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

// Encode a ConnRply from a buffer
func (pkt *ConnRply) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, pkt.Id())
	XdrPutUint32(buffer, pkt.Xid)
	XdrPutNotification(buffer, pkt.Options)
}
