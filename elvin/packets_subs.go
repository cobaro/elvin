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

// Packet: SubAddRequest
type SubAddRequest struct {
	XID            uint32
	Expression     string
	AcceptInsecure bool
	Keys           []Keyset
}

// Integer value of packet type
func (pkt *SubAddRequest) ID() int {
	return PacketSubAddRequest
}

// String representation of packet type
func (pkt *SubAddRequest) IDString() string {
	return "SubAddRequest"
}

// Pretty print with indent
func (pkt *SubAddRequest) IString(indent string) string {
	return fmt.Sprintf("%sXID %v\n%sExpression %v\n%sAcceptInsecure %v\n%sKeys %v\n",
		indent, pkt.XID,
		indent, pkt.Expression,
		indent, pkt.AcceptInsecure,
		indent, pkt.Keys,
	)
}

// Pretty print without indent so generic ToString() works
func (pkt *SubAddRequest) String() string {
	return pkt.IString("")
}

// Decode a SubAddRequest packet from a byte array
func (pkt *SubAddRequest) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
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

// Encode a SubAddRequest from a buffer
func (pkt *SubAddRequest) Encode(buffer *bytes.Buffer) (xID uint32) {
	xID = XID()
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutUint32(buffer, xID)
	XdrPutString(buffer, pkt.Expression)
	XdrPutBool(buffer, pkt.AcceptInsecure)
	XdrPutKeys(buffer, pkt.Keys)

	return
}

// Packet: SubReply
type SubReply struct {
	XID   uint32
	SubID int64
}

// Integer value of packet type
func (pkt *SubReply) ID() int {
	return PacketSubReply
}

// String representation of packet type
func (pkt *SubReply) IDString() string {
	return "SubReply"
}

// Pretty print with indent
func (pkt *SubReply) IString(indent string) string {
	return fmt.Sprintf("%sXID %v\n%sSubID %v\n",
		indent, pkt.XID,
		indent, pkt.SubID,
	)
}

// Pretty print without indent so generic ToString() works
func (pkt *SubReply) String() string {
	return pkt.IString("")
}

// Decode a SubReply packet from a byte array
func (pkt *SubReply) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.SubID, used, err = XdrGetInt64(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

// Encode a SubReply from a buffer
func (pkt *SubReply) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutUint32(buffer, pkt.XID)
	XdrPutInt64(buffer, pkt.SubID)
}

// Packet: SubDelRequest
type SubDelRequest struct {
	XID   uint32
	SubID int64
}

// Integer value of packet type
func (pkt *SubDelRequest) ID() int {
	return PacketSubDelRequest
}

// String representation of packet type
func (pkt *SubDelRequest) IDString() string {
	return "SubDelRequest"
}

// Pretty print with indent
func (pkt *SubDelRequest) IString(indent string) string {
	return fmt.Sprintf("%sXID %v\n%sSubID %v\n",
		indent, pkt.XID,
		indent, pkt.SubID,
	)
}

// Pretty print without indent so generic ToString() works
func (pkt *SubDelRequest) String() string {
	return pkt.IString("")
}

// Decode a SubDelRequest packet from a byte array
func (pkt *SubDelRequest) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.SubID, used, err = XdrGetInt64(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

// Encode a SubDelRequest from a buffer
func (pkt *SubDelRequest) Encode(buffer *bytes.Buffer) (xID uint32) {
	xID = XID()
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutUint32(buffer, xID)
	XdrPutInt64(buffer, pkt.SubID)
	return
}

// Packet: SubModRequest
type SubModRequest struct {
	XID            uint32
	SubID          int64
	Expression     string
	AcceptInsecure bool
	AddKeys        []Keyset
	DelKeys        []Keyset
}

// Integer value of packet type
func (pkt *SubModRequest) ID() int {
	return PacketSubModRequest
}

// String representation of packet type
func (pkt *SubModRequest) IDString() string {
	return "SubModRequest"
}

// Pretty print with indent
func (pkt *SubModRequest) IString(indent string) string {
	return fmt.Sprintf("%sXID %v\n%sSubID %v\n%sExpression %v\n%sAcceptInsecure %v\n%sAddKeys %v\n%sDelKeys %v\n",
		indent, pkt.XID,
		indent, pkt.SubID,
		indent, pkt.Expression,
		indent, pkt.AcceptInsecure,
		indent, pkt.AddKeys,
		indent, pkt.DelKeys,
	)
}

// Pretty print without indent so generic ToString() works
func (pkt *SubModRequest) String() string {
	return pkt.IString("")
}

// Decode a SubModRequest packet from a byte array
func (pkt *SubModRequest) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.SubID, used, err = XdrGetInt64(bytes[offset:])
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

// Encode a SubModRequest from a buffer
func (pkt *SubModRequest) Encode(buffer *bytes.Buffer) (xID uint32) {
	xID = XID()
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutUint32(buffer, xID)
	XdrPutInt64(buffer, pkt.SubID)
	XdrPutString(buffer, pkt.Expression)
	XdrPutBool(buffer, pkt.AcceptInsecure)
	XdrPutKeys(buffer, pkt.AddKeys)
	XdrPutKeys(buffer, pkt.DelKeys)

	return
}
