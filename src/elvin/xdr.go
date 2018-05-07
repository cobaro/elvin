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
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// FIXME The current state is that the getters use a []byte and the
// putters use a bytes.Buffer. This is because for now we're being
// quick and dirty and experimental. Much will depend on subsequent
// performance tuning

// FIXME range checking

// Get an xdr marshalled 32 bit signed int
func XdrGetInt32(bytes []byte) (i int, used int) {
	return int(binary.BigEndian.Uint32(bytes)), 4
}

// Put an xdr marshalled 32 bit signed int
func XdrPutInt32(buffer *bytes.Buffer, i int) {
	buffer.WriteByte(byte(i >> 24))
	buffer.WriteByte(byte(i >> 16))
	buffer.WriteByte(byte(i >> 8))
	buffer.WriteByte(byte(i))
	return
}

// Get an xdr marshalled 32 bit unsigned int
func XdrGetUint32(bytes []byte) (u uint32, used int) {
	return binary.BigEndian.Uint32(bytes), 4
}

// Put an xdr marshalled 32 bit unsigned int
func XdrPutUint32(buffer *bytes.Buffer, u uint32) {
	buffer.WriteByte(byte(u >> 24))
	buffer.WriteByte(byte(u >> 16))
	buffer.WriteByte(byte(u >> 8))
	buffer.WriteByte(byte(u))
}

// Get an xdr marshalled 64 bit signed int
func XdrGetInt64(bytes []byte) (i int64, used int) {
	return int64(binary.BigEndian.Uint64(bytes)), 8
}

// Put an xdr marshalled 64 bit signed int
func XdrPutInt64(buffer *bytes.Buffer, i int64) {
	buffer.WriteByte(byte(i >> 56))
	buffer.WriteByte(byte(i >> 48))
	buffer.WriteByte(byte(i >> 40))
	buffer.WriteByte(byte(i >> 32))
	buffer.WriteByte(byte(i >> 24))
	buffer.WriteByte(byte(i >> 16))
	buffer.WriteByte(byte(i >> 8))
	buffer.WriteByte(byte(i))
	return
}

// Get an xdr marshalled 64 bit unsigned int
func XdrGetUint64(bytes []byte) (u uint64, used int) {
	return binary.BigEndian.Uint64(bytes), 8
}

// Put an xdr marshalled 64 bit unsigned int
func XdrPutUint64(buffer *bytes.Buffer, u uint64) {
	buffer.WriteByte(byte(u >> 56))
	buffer.WriteByte(byte(u >> 48))
	buffer.WriteByte(byte(u >> 40))
	buffer.WriteByte(byte(u >> 32))
	buffer.WriteByte(byte(u >> 24))
	buffer.WriteByte(byte(u >> 16))
	buffer.WriteByte(byte(u >> 8))
	buffer.WriteByte(byte(u))
	return
}

// Get an xdr marshalled string
func XdrGetString(bytes []byte) (s string, used int) {
	// string length
	length, used := XdrGetInt32(bytes)
	// name
	return string(bytes[used : used+length]), used + length + (3 - (length+3)%4) // strings use 4 byte boundaries
}

// Put an xdr marshalled string
func XdrPutString(buffer *bytes.Buffer, s string) {
	// string length
	length := len(s)
	XdrPutInt32(buffer, length)
	buffer.WriteString(s)
	for length%4 > 0 { // align to 4 byte boundaries
		buffer.WriteByte(byte(0))
		length++
	}
	return
}

// Get an xdr marshalled list of opaque bytes
func XdrGetOpaque(bytes []byte) (b []byte, used int) {
	// string length
	length, used := XdrGetInt32(bytes)
	// name
	return bytes[used : used+length], used + length + (3 - (length+3)%4) // opaques use 4 byte boundaries
}

// Put an xdr marshalled list of opaque bytes
func XdrPutOpaque(buffer *bytes.Buffer, b []byte) {
	// FIXME: This needs a test case
	length := len(b)
	XdrPutInt32(buffer, length)
	buffer.Write(b)
	for length%4 > 0 { // align to 4 byte boundaries
		buffer.WriteByte(byte(0))
		length++
	}
	return
}

// Get an xdr marshalled 64 bit floating point
func XdrGetFloat64(bytes []byte) (f float64, used int) {
	f64 := math.Float64frombits(uint64(bytes[7]) | uint64(bytes[6])<<8 |
		uint64(bytes[5])<<16 | uint64(bytes[4])<<24 |
		uint64(bytes[3])<<32 | uint64(bytes[2])<<40 |
		uint64(bytes[1])<<48 | uint64(bytes[0])<<56)
	return f64, 8
}

// Put an xdr marshalled 64 bit floating point
func XdrPutFloat64(buffer *bytes.Buffer, f float64) {
	// FIXME: Introduce a marshaller to hold temp vars
	var b8 = make([]byte, 8)
	u64 := math.Float64bits(f)
	b8[7] = byte(u64)
	b8[6] = byte(u64 >> 8)
	b8[5] = byte(u64 >> 16)
	b8[4] = byte(u64 >> 24)
	b8[3] = byte(u64 >> 32)
	b8[2] = byte(u64 >> 40)
	b8[1] = byte(u64 >> 48)
	b8[0] = byte(u64 >> 56)
	buffer.Write(b8)
	return // err FIXME, Write can fail
}

// Get an xdr marshalled Elvin Notification
func XdrGetNotification(bytes []byte) (nfn map[string]interface{}, used int, err error) {
	nfn = make(map[string]interface{})
	offset := 0

	// Number of elements
	elementCount, used := XdrGetUint32(bytes[offset:])
	offset += used

	for elementCount > 0 {
		name, used := XdrGetString(bytes[offset:])
		offset += used

		// Type of value
		elementType, used := XdrGetInt32(bytes[offset:])
		offset += used

		// values
		switch elementType {
		case NotificationInt32:
			nfn[name], used = XdrGetInt32(bytes[offset:])
			offset += used
			break

		case NotificationInt64:
			nfn[name], used = XdrGetInt64(bytes[offset:])
			offset += used
			break

		case NotificationFloat64:
			nfn[name], used = XdrGetFloat64(bytes[offset:])
			offset += used
			break

		case NotificationString:
			var used int
			nfn[name], used = XdrGetString(bytes[offset:])
			offset += used
			break

		case NotificationOpaque:
			var used int
			nfn[name], used = XdrGetOpaque(bytes[offset:])
			offset += used
			break

		default:
			return nil, 0, errors.New("Marshalling failed: unknown element type")
		}

		elementCount--
	}

	return nfn, offset, err
}

// Put an xdr marshalled Elvin Notification
func XdrPutNotification(buffer *bytes.Buffer, nfn map[string]interface{}) {

	// Number of elements
	XdrPutInt32(buffer, len(nfn))

	for k, v := range nfn {
		// Key
		XdrPutString(buffer, k)
		// Value
		switch t := v.(type) {
		case int:
			XdrPutInt32(buffer, 1)
			XdrPutInt32(buffer, v.(int))
		case int64:
			XdrPutInt32(buffer, 2)
			XdrPutInt64(buffer, v.(int64))
		case float64:
			XdrPutInt32(buffer, 3)
			XdrPutFloat64(buffer, v.(float64))
		case string:
			XdrPutInt32(buffer, 4)
			XdrPutString(buffer, v.(string))
		case []byte:
			XdrPutInt32(buffer, 5)
			XdrPutOpaque(buffer, v.([]uint8))
		default:
			XdrPutInt32(buffer, 0)
			panic(fmt.Sprintf("What *type* is: %v", t))
			return
		}
	}
	return
}

// Fixme: implement
// Get an xdr marshalled keyset
func XdrGetKeys(bytes []byte) (keys [][]byte, used int, err error) {
	return nil, 0, nil

	offset := 0

	// Number of elements
	fmt.Println("BYTES", bytes)
	elementCount, used := XdrGetUint32(bytes[offset:])
	offset += used
	fmt.Println("elementCount =", elementCount)

	keys = make([][]byte, elementCount)
	fmt.Println("now", keys)

	for elementCount > 0 {
		fmt.Println("EC", elementCount)
		keys[elementCount-1], used = XdrGetOpaque(bytes[offset:])
		offset += used
		elementCount--
	}
	fmt.Println("then", keys)

	return keys, offset, nil
}

// FIXME: implement
// Put an xdr marshalled keyset
func XdrPutKeys(buffer *bytes.Buffer, keys [][]byte) (err error) {
	return
}
