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

// Packet: Connection Reply
type ConnRply struct {
	Xid     uint32
	Options map[string]interface{}
}

// Integer value of packet type
func (c *ConnRply) Id() int {
	return PacketConnRply
}

// String representation of packet type
func (c *ConnRply) IdString() string {
	return "ConnRply"
}

// Pretty print with indent
func (c *ConnRply) IString(indent string) string {
	return fmt.Sprintf("%sXid: %d\n%sOptions %v\n",
		indent, c.Xid,
		indent, c.Options)

}

// Pretty print without indent so generic ToString() works
func (c *ConnRply) String() string {
	return c.IString("")
}

// Decode a ConnRply packet from a byte array
func (c *ConnRply) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	c.Xid, used = XdrGetUint32(bytes[offset:])
	offset += used

	c.Options, used, err = XdrGetNotification(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

// Encode a ConnRply from a buffer
func (c *ConnRply) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, c.Id())
	XdrPutUint32(buffer, c.Xid)
	XdrPutNotification(buffer, c.Options)
}
