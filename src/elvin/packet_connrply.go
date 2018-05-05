// Copyright 2018 Cobaro Pty Ltd. All Rights Reserved.

// This file is part of elvin
//
// elvin is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// elvin is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with elvin. If not, see <http://www.gnu.org/licenses/>.

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
