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
func (s *SubAddRqst) Id() int {
	return PacketSubAddRqst
}

// String representation of packet type
func (s *SubAddRqst) IdString() string {
	return "SubAddRqst"
}

// Pretty print with indent
func (s *SubAddRqst) IString(indent string) string {
	return fmt.Sprintf("%sXid %v\n%sExpression %v\n%sAcceptInsecure %v\n%sKeys %v\n",
		indent, s.Xid,
		indent, s.Expression,
		indent, s.AcceptInsecure,
		indent, s.Keys,
	)
}

// Pretty print without indent so generic ToString() works
func (s *SubAddRqst) String() string {
	return s.IString("")
}

// Decode a SubAddRqst packet from a byte array
func (s *SubAddRqst) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	s.Xid, used = XdrGetUint32(bytes[offset:])
	offset += used
	s.Expression, used = XdrGetString(bytes[offset:])
	offset += used
	s.AcceptInsecure, used = XdrGetBool(bytes[offset:])
	offset += used
	s.Keys, used, err = XdrGetKeys(bytes[offset:])
	offset += used
	return err
}

// Encode a SubAddRqst from a buffer
func (s *SubAddRqst) Encode(buffer *bytes.Buffer) {
	// FIXME: error handling
	XdrPutInt32(buffer, s.Id())
	XdrPutUint32(buffer, s.Xid)
	XdrPutBool(buffer, s.AcceptInsecure)
	XdrPutKeys(buffer, s.Keys)
}

// Packet: SubRply

type SubRply struct {
	Xid   uint32
	Subid uint64
}

// Integer value of packet type
func (s *SubRply) Id() int {
	return PacketSubRply
}

// String representation of packet type
func (s *SubRply) IdString() string {
	return "SubRply"
}

// Pretty print with indent
func (s *SubRply) IString(indent string) string {
	return fmt.Sprintf("%sXid %v\n%sSubid %v\n",
		indent, s.Xid,
		indent, s.Subid,
	)
}

// Pretty print without indent so generic ToString() works
func (s *SubRply) String() string {
	return s.IString("")
}

// Decode a SubRply packet from a byte array
func (s *SubRply) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	s.Xid, used = XdrGetUint32(bytes[offset:])
	offset += used
	s.Subid, used = XdrGetUint64(bytes[offset:])
	offset += used
	return err
}

// Encode a SubRply from a buffer
func (s *SubRply) Encode(buffer *bytes.Buffer) {
	// FIXME: error handling
	XdrPutInt32(buffer, s.Id())
	XdrPutUint32(buffer, s.Xid)
	XdrPutUint64(buffer, s.Subid)
}
