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

// Packet: SubAddRqst

type SubAddRqst struct {
	Xid            uint32
	Expression     string
	AcceptInsecure bool
	Keys           []Keyset
}

// Integer value of packet type
func (pkt *SubAddRqst) Id() int {
	return PacketSubAddRqst
}

// String representation of packet type
func (pkt *SubAddRqst) IdString() string {
	return "SubAddRqst"
}

// Pretty print with indent
func (pkt *SubAddRqst) IString(indent string) string {
	return fmt.Sprintf("%sXid %v\n%sExpression %v\n%sAcceptInsecure %v\n%sKeys %v\n",
		indent, pkt.Xid,
		indent, pkt.Expression,
		indent, pkt.AcceptInsecure,
		indent, pkt.Keys,
	)
}

// Pretty print without indent so generic ToString() works
func (pkt *SubAddRqst) String() string {
	return pkt.IString("")
}

// Decode a SubAddRqst packet from a byte array
func (pkt *SubAddRqst) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.Xid, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Expression, used, err = XdrGetString(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.AcceptInsecure, used, err = XdrGetBool(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Keys, used, err = XdrGetKeys(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

// Encode a SubAddRqst from a buffer
func (pkt *SubAddRqst) Encode(buffer *bytes.Buffer) (xid uint32) {
	xid = Xid()
	XdrPutInt32(buffer, int32(pkt.Id()))
	XdrPutUint32(buffer, xid)
	XdrPutString(buffer, pkt.Expression)
	XdrPutBool(buffer, pkt.AcceptInsecure)
	XdrPutKeys(buffer, pkt.Keys)

	return
}

// Packet: SubRply

type SubRply struct {
	Xid   uint32
	Subid uint64
}

// Integer value of packet type
func (pkt *SubRply) Id() int {
	return PacketSubRply
}

// String representation of packet type
func (pkt *SubRply) IdString() string {
	return "SubRply"
}

// Pretty print with indent
func (pkt *SubRply) IString(indent string) string {
	return fmt.Sprintf("%sXid %v\n%sSubid %v\n",
		indent, pkt.Xid,
		indent, pkt.Subid,
	)
}

// Pretty print without indent so generic ToString() works
func (pkt *SubRply) String() string {
	return pkt.IString("")
}

// Decode a SubRply packet from a byte array
func (pkt *SubRply) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.Xid, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Subid, used, err = XdrGetUint64(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

// Encode a SubRply from a buffer
func (pkt *SubRply) Encode(buffer *bytes.Buffer) {
	// FIXME: error handling
	XdrPutInt32(buffer, int32(pkt.Id()))
	XdrPutUint32(buffer, pkt.Xid)
	XdrPutUint64(buffer, pkt.Subid)
}

// Packet: SubDelRqst

type SubDelRqst struct {
	Xid   uint32
	Subid uint64
}

// Integer value of packet type
func (pkt *SubDelRqst) Id() int {
	return PacketSubDelRqst
}

// String representation of packet type
func (pkt *SubDelRqst) IdString() string {
	return "SubDelRqst"
}

// Pretty print with indent
func (pkt *SubDelRqst) IString(indent string) string {
	return fmt.Sprintf("%sXid %v\n%sSubid %v\n",
		indent, pkt.Xid,
		indent, pkt.Subid,
	)
}

// Pretty print without indent so generic ToString() works
func (pkt *SubDelRqst) String() string {
	return pkt.IString("")
}

// Decode a SubDelRqst packet from a byte array
func (pkt *SubDelRqst) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.Xid, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Subid, used, err = XdrGetUint64(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

// Encode a SubDelRqst from a buffer
func (pkt *SubDelRqst) Encode(buffer *bytes.Buffer) {
	// FIXME: error handling
	XdrPutInt32(buffer, int32(pkt.Id()))
	XdrPutUint32(buffer, pkt.Xid)
	XdrPutUint64(buffer, pkt.Subid)
}

// Packet: SubModRqst

type SubModRqst struct {
	Xid            uint32
	Subid          uint64
	Expression     string
	AcceptInsecure bool
	AddKeys        []Keyset
	DelKeys        []Keyset
}

// Integer value of packet type
func (pkt *SubModRqst) Id() int {
	return PacketSubModRqst
}

// String representation of packet type
func (pkt *SubModRqst) IdString() string {
	return "SubModRqst"
}

// Pretty print with indent
func (pkt *SubModRqst) IString(indent string) string {
	return fmt.Sprintf("%sXid %v\n%sSubId %v\n%sExpression %v\n%sAcceptInsecure %v\n%sAddKeys %v\n%sDelKeys %v\n",
		indent, pkt.Xid,
		indent, pkt.Subid,
		indent, pkt.Expression,
		indent, pkt.AcceptInsecure,
		indent, pkt.AddKeys,
		indent, pkt.DelKeys,
	)
}

// Pretty print without indent so generic ToString() works
func (pkt *SubModRqst) String() string {
	return pkt.IString("")
}

// Decode a SubModRqst packet from a byte array
func (pkt *SubModRqst) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.Xid, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Subid, used, err = XdrGetUint64(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Expression, used, err = XdrGetString(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.AcceptInsecure, used, err = XdrGetBool(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.AddKeys, used, err = XdrGetKeys(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.DelKeys, used, err = XdrGetKeys(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

// Encode a SubModRqst from a buffer
func (pkt *SubModRqst) Encode(buffer *bytes.Buffer) {
	// FIXME: error handling
	XdrPutInt32(buffer, int32(pkt.Id()))
	XdrPutUint32(buffer, pkt.Xid)
	XdrPutUint64(buffer, pkt.Subid)
	XdrPutBool(buffer, pkt.AcceptInsecure)
	XdrPutKeys(buffer, pkt.AddKeys)
	XdrPutKeys(buffer, pkt.DelKeys)
}
