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
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAx1MAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package elvin

import (
	"bytes"
	"fmt"
)

// Packet: Nack

type Nack struct {
	XID       uint32
	ErrorCode uint16
	Message   string
	Args      []interface{}
}

// Integer value of packet type
func (pkt *Nack) ID() int {
	return PacketNack
}

// String representation of packet type
func (pkt *Nack) IDString() string {
	return "Nack"
}

// Pretty print with indent
func (pkt *Nack) IString(indent string) string {
	return fmt.Sprintf("%sXID:%v [%d] %s",
		indent, pkt.XID, int(pkt.ErrorCode),
		fmt.Sprintf(ElvinStringToFormatString(pkt.Message), pkt.Args...))
}

// Pretty print without indent so generic ToString() works
func (pkt *Nack) String() string {
	return pkt.IString("")
}

// Decode a Nack packet from a byte array
func (pkt *Nack) Decode(bytes []byte) (err error) {
	var used int
	offset := 4
	// FIXME: at some point we will want to return how many bytes we consumed
	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.ErrorCode, used, err = XdrGetUint16(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Message, used, err = XdrGetString(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	// The type of the arguments is defined by the ErrorCode so we
	// look this up in NackArgs
	nackArgs, ok := ProtocolErrors[pkt.ErrorCode]
	if !ok {
		return // Nothing we can do
	}

	// Arg values
	pkt.Args = make([]interface{}, nackArgs.NumArgs)
	for i := 0; i < nackArgs.NumArgs; i++ {
		switch nackArgs.ArgTypes[i].(type) {
		case int32:
			pkt.Args[i], used, err = XdrGetInt32(bytes[offset:])
			if err != nil {
				return err
			}
		case uint32:
			pkt.Args[i], used, err = XdrGetUint32(bytes[offset:])
			if err != nil {
				return err
			}
		case int64:
			pkt.Args[i], used, err = XdrGetInt64(bytes[offset:])
			if err != nil {
				return err
			}
		case uint64:
			pkt.Args[i], used, err = XdrGetUint64(bytes[offset:])
			if err != nil {
				return err
			}
		case float64:
			pkt.Args[i], used, err = XdrGetFloat64(bytes[offset:])
			if err != nil {
				return err
			}
		case string:
			pkt.Args[i], used, err = XdrGetString(bytes[offset:])
			if err != nil {
				return err
			}
		default:
			panic(fmt.Sprintf("Bad *type* in Nack arg %d", i))
		}
		offset += used
	}
	return nil
}

// Encode a Nack from a buffer
func (pkt *Nack) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutUint32(buffer, pkt.XID)
	XdrPutUint16(buffer, pkt.ErrorCode)
	XdrPutString(buffer, pkt.Message)

	// The type of the arguments is defined by the ErrorCode so we
	// look this up in NackArgs
	nackArgs, ok := ProtocolErrors[pkt.ErrorCode]
	if !ok {
		return // Nothing we can do
	}

	// Arg values
	for i := 0; i < nackArgs.NumArgs; i++ {
		value := nackArgs.ArgTypes[i]
		switch nackArgs.ArgTypes[i].(type) {
		case int32:
			XdrPutInt32(buffer, value.(int32))
		case uint32:
			XdrPutUint32(buffer, value.(uint32))
		case int64:
			XdrPutInt64(buffer, value.(int64))
		case uint64:
			XdrPutUint64(buffer, value.(uint64))
		case float64:
			XdrPutFloat64(buffer, value.(float64))
		case string:
			XdrPutString(buffer, value.(string))
		case []byte:
			XdrPutOpaque(buffer, value.([]uint8))
		default:
			panic(fmt.Sprintf("Bad *type* in Nack arg: %v", value))
		}
	}
	return
}
