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

// Packet: NotifyEmit
type NotifyEmit struct {
	NameValue       map[string]interface{}
	DeliverInsecure bool
	Keys            []Keyset
}

// Integer value of packet type
func (pkt *NotifyEmit) Id() int {
	return PacketNotifyEmit
}

// String representation of packet type
func (pkt *NotifyEmit) IdString() string {
	return "NotifyEmit"
}

// Pretty print with indent
func (pkt *NotifyEmit) IString(indent string) string {
	return fmt.Sprintf("%sNameValue %v\n%sDeliverInsecure %v\n%sKeys %v\n",
		indent, pkt.NameValue,
		indent, pkt.DeliverInsecure,
		indent, pkt.Keys)
}

// Pretty print without indent so generic ToString() works
func (pkt *NotifyEmit) String() string {
	return pkt.IString("")
}

// Decode a NotifyEmit packet from a byte array
func (pkt *NotifyEmit) Decode(bytes []byte) (err error) {
	var used int
	offset := 4
	if pkt.NameValue, used, err = XdrGetNotification(bytes[offset:]); err != nil {
		return err
	}
	offset += used

	pkt.DeliverInsecure, used, err = XdrGetBool(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	if pkt.Keys, used, err = XdrGetKeys(bytes[offset:]); err != nil {
		return err
	}
	offset += used

	// FIXME: at some point we will want to return how many bytes we consumed
	return nil
}

// Encode a NotifyEmit from a buffer
func (pkt *NotifyEmit) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, int32(pkt.Id()))
	XdrPutNotification(buffer, pkt.NameValue)
	XdrPutBool(buffer, pkt.DeliverInsecure)
	XdrPutKeys(buffer, pkt.Keys)
}

// Packet: NotifyDeliver
type NotifyDeliver struct {
	NameValue map[string]interface{}
	Secure    []uint64
	Insecure  []uint64
}

// Integer value of packet type
func (pkt *NotifyDeliver) Id() int {
	return PacketNotifyDeliver
}

// String representation of packet type
func (pkt *NotifyDeliver) IdString() string {
	return "NotifyDeliver"
}

// Pretty print with indent
func (pkt *NotifyDeliver) IString(indent string) string {
	return fmt.Sprintf("%sNameValue %v\n%sSecure %v\n%sInsecure %v\n",
		indent, pkt.NameValue,
		indent, pkt.Secure,
		indent, pkt.Insecure,
	)
}

// Pretty print without indent so generic ToString() works
func (pkt *NotifyDeliver) String() string {
	return pkt.IString("")
}

// Decode a NotifyDeliver packet from a byte array
func (pkt *NotifyDeliver) Decode(bytes []byte) (err error) {
	var used int
	offset := 4
	if pkt.NameValue, used, err = XdrGetNotification(bytes[offset:]); err != nil {
		return err
	}
	offset += used

	secureCount, used, err := XdrGetInt32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	for i := int32(0); i < secureCount; i++ {
		pkt.Secure[i], used, err = XdrGetUint64(bytes[offset:])
		if err != nil {
			return err
		}
		offset += used
	}

	insecureCount, used, err := XdrGetInt32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	for i := int32(0); i < insecureCount; i++ {
		pkt.Insecure[i], used, err = XdrGetUint64(bytes[offset:])
		if err != nil {
			return err
		}
		offset += used
	}

	return nil
}

// Encode a NotifyDeliver from a buffer
func (pkt *NotifyDeliver) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, int32(pkt.Id()))
	XdrPutNotification(buffer, pkt.NameValue)
	XdrPutInt32(buffer, int32(len(pkt.Secure)))
	for i := 0; i < len(pkt.Secure); i++ {
		XdrPutUint64(buffer, pkt.Secure[i])
	}
	XdrPutInt32(buffer, int32(len(pkt.Insecure)))
	for i := 0; i < len(pkt.Insecure); i++ {
		XdrPutUint64(buffer, pkt.Insecure[i])
	}
}
