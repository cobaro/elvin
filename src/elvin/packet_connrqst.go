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
