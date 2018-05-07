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
	"fmt"
)

// Packet: Connection Request
type ConnRqst struct {
	Xid          uint32
	VersionMajor uint32
	VersionMinor uint32
	Options      map[string]interface{}
	KeysNfn      [][]byte
	KeysSub      [][]byte
}

// Integer value of packet type
func (c *ConnRqst) Id() int {
	return PacketConnRqst
}

// String representation of packet type
func (c *ConnRqst) IdString() string {
	return "ConnRqst"
}

// Pretty print with indent
func (c *ConnRqst) IString(indent string) string {
	return fmt.Sprintf(
		"%sXid: %d\n"+
			"%sVersionMajor %d\n"+
			"%sVersionMinor %d\n"+
			"%sOptions %v\n"+
			"%sKeysNfn: %v\n"+
			"%sKeysSub: %v\n",
		indent, c.Xid,
		indent, c.VersionMajor,
		indent, c.VersionMinor,
		indent, c.Options,
		indent, c.KeysNfn,
		indent, c.KeysSub)
}

// Pretty print without indent so generic ToString() works
func (c *ConnRqst) String() string {
	return c.IString("")
}

// Decode a ConnRqst packet from a byte array
func (c *ConnRqst) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	c.Xid, used = XdrGetUint32(bytes[offset:])
	offset += used

	c.VersionMajor, used = XdrGetUint32(bytes[offset:])
	offset += used

	c.VersionMinor, used = XdrGetUint32(bytes[offset:])
	offset += used

	c.Options, used, err = XdrGetNotification(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	c.KeysNfn, used, err = XdrGetKeys(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	c.KeysSub, used, err = XdrGetKeys(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	// fmt.Println("keysN:", c.KeysNfn)
	// fmt.Println("keysS:", c.KeysSub)
	// fmt.Println(bytes[offset:])
	// fmt.Println(c)
	return nil
}

func (c *ConnRqst) Encode() (bytes []byte, err error) {
	// FIXME: Strictly speaking the router doesn't need this
	return nil, nil
}
