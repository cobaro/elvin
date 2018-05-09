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
type NotifyEmit struct {
	NameValue       map[string]interface{}
	DeliverInsecure bool
	Keys            []Keyset
}

// Integer value of packet type
func (n *NotifyEmit) Id() int {
	return PacketNotifyEmit
}

// String representation of packet type
func (n *NotifyEmit) IdString() string {
	return "NotifyEmit"
}

// Pretty print with indent
func (n *NotifyEmit) IString(indent string) string {
	return fmt.Sprintf("%sNameValue %v\n%sDeliverInsecure %v\n%sKeys %v\n",
		indent, n.NameValue,
		indent, n.DeliverInsecure,
		indent, n.Keys)
}

// Pretty print without indent so generic ToString() works
func (n *NotifyEmit) String() string {
	return n.IString("")
}

// Decode a NotifyEmit packet from a byte array
func (n *NotifyEmit) Decode(bytes []byte) (err error) {
	var used int
	offset := 4
	if n.NameValue, used, err = XdrGetNotification(bytes[offset:]); err != nil {
		return err
	}
	offset += used

	n.DeliverInsecure, used = XdrGetBool(bytes[offset:])
	offset += used

	if n.Keys, used, err = XdrGetKeys(bytes[offset:]); err != nil {
		return err
	}
	offset += used

	// FIXME: at some point we will want to return how many bytes we consumed
	return nil
}

// Encode a NotifyEmit from a buffer
func (n *NotifyEmit) Encode(buffer *bytes.Buffer) {
	XdrPutNotification(buffer, n.NameValue)
	XdrPutBool(buffer, n.DeliverInsecure)
	XdrPutKeys(buffer, n.Keys)
}
