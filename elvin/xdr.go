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

// Defined errors we return
var NotEnoughSpace error = errors.New("Input buffer too small")

// FIXME The current state is that the getters use a []byte and the
// putters use a bytes.Buffer. This is because for now we're being
// quick and dirty and experimental. Much will depend on subsequent
// performance tuning

// FIXME range checking

// Get an xdr marshalled 32 bit signed int
func XdrGetInt32(bytes []byte) (i int32, used int, err error) {
	if len(bytes) < 4 {
		return 0, 0, NotEnoughSpace
	}
	return int32(binary.BigEndian.Uint32(bytes)), 4, nil
}

// Put an xdr marshalled 32 bit signed int
func XdrPutInt32(buffer *bytes.Buffer, i int32) {
	buffer.WriteByte(byte(i >> 24))
	buffer.WriteByte(byte(i >> 16))
	buffer.WriteByte(byte(i >> 8))
	buffer.WriteByte(byte(i))
	return
}

// Get an xdr marshalled 32 bit unsigned int
func XdrGetUint32(bytes []byte) (u uint32, used int, err error) {
	if len(bytes) < 4 {
		return 0, 0, NotEnoughSpace
	}
	return binary.BigEndian.Uint32(bytes), 4, nil
}

// Put an xdr marshalled 32 bit unsigned int
func XdrPutUint32(buffer *bytes.Buffer, u uint32) {
	buffer.WriteByte(byte(u >> 24))
	buffer.WriteByte(byte(u >> 16))
	buffer.WriteByte(byte(u >> 8))
	buffer.WriteByte(byte(u))
}

// Get an xdr marshalled 64 bit signed int
func XdrGetInt64(bytes []byte) (i int64, used int, err error) {
	return int64(binary.BigEndian.Uint64(bytes)), 8, nil
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
func XdrGetUint64(bytes []byte) (u uint64, used int, err error) {
	return binary.BigEndian.Uint64(bytes), 8, nil
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

// Get an xdr marshalled bool (on the wire it's a 32 bit signed int)
func XdrGetBool(bytes []byte) (b bool, used int, err error) {
	i, used, err := XdrGetInt32(bytes)
	if err != nil {
		return false, 0, err
	}
	b = (i != 0)
	return b, used, nil
}

// Put an xdr marshalled bool (ont the wire it's a 32 bit signed int)
func XdrPutBool(buffer *bytes.Buffer, b bool) {
	var i int32 = 0
	if b {
		i = 1
	}
	XdrPutInt32(buffer, i)
	return
}

// Get an xdr marshalled int16 (on the wire it's a 32 bit signed int)
func XdrGetInt16(bytes []byte) (i16 int16, used int, err error) {
	i, used, err := XdrGetInt32(bytes)
	if err != nil {
		return 0, 0, err
	}
	return int16(i), used, nil
}

// Put an xdr marshalled int16 (ont the wire it's a 32 bit signed int)
func XdrPutInt16(buffer *bytes.Buffer, i16 int16) {
	XdrPutInt32(buffer, int32(i16))
	return
}

// Get an xdr marshalled uint16 (on the wire it's a 32 bit unsigned int)
func XdrGetUint16(bytes []byte) (u16 uint16, used int, err error) {
	u, used, err := XdrGetUint32(bytes)
	if err != nil {
		return 0, 0, err
	}
	return uint16(u), used, nil
}

// Put an xdr marshalled uint16 (ont the wire it's a 32 bit unsigned int)
func XdrPutUint16(buffer *bytes.Buffer, u16 uint16) {
	XdrPutUint32(buffer, uint32(u16))
	return
}

// Get an xdr marshalled string
func XdrGetString(bytes []byte) (s string, used int, err error) {
	// string length
	length, used, err := XdrGetInt32(bytes)
	if err != nil {
		return "", 0, err
	}
	// name
	return string(bytes[used : used+int(length)]), used + int(length) + (3 - (int(length)+3)%4), nil // strings use 4 byte boundaries
}

// Put an xdr marshalled string
func XdrPutString(buffer *bytes.Buffer, s string) {
	// string length
	length := int32(len(s))
	XdrPutInt32(buffer, length)
	buffer.WriteString(s)
	for length%4 > 0 { // align to 4 byte boundaries
		buffer.WriteByte(byte(0))
		length++
	}
	return
}

// Get an xdr marshalled list of opaque bytes
func XdrGetOpaque(bytes []byte) (b []byte, used int, err error) {
	// string length
	length, used, err := XdrGetInt32(bytes)
	if err != nil {
		return nil, 0, err
	}

	// name
	return bytes[used : used+int(length)], used + int(length) + (3 - (int(length)+3)%4), nil // opaques use 4 byte boundaries
}

// Put an xdr marshalled list of opaque bytes
func XdrPutOpaque(buffer *bytes.Buffer, b []byte) {
	// FIXME: This needs a test case
	length := int32(len(b))
	XdrPutInt32(buffer, length)
	buffer.Write(b)
	for length%4 > 0 { // align to 4 byte boundaries
		buffer.WriteByte(byte(0))
		length++
	}
	return
}

// Get an xdr marshalled 64 bit floating point
func XdrGetFloat64(bytes []byte) (f float64, used int, err error) {
	f64 := math.Float64frombits(uint64(bytes[7]) | uint64(bytes[6])<<8 |
		uint64(bytes[5])<<16 | uint64(bytes[4])<<24 |
		uint64(bytes[3])<<32 | uint64(bytes[2])<<40 |
		uint64(bytes[1])<<48 | uint64(bytes[0])<<56)
	return f64, 8, nil
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
func XdrGetValue(bytes []byte) (val interface{}, used int, err error) {

	// Type of value
	offset := 0
	elementType, used, err := XdrGetInt32(bytes[offset:])
	if err != nil {
		return nil, 0, err
	}
	offset += used
	var value interface{}

	// The values itself
	switch elementType {
	case NotificationInt32:
		value, used, err = XdrGetInt32(bytes[offset:])
	case NotificationInt64:
		value, used, err = XdrGetInt64(bytes[offset:])
	case NotificationFloat64:
		value, used, err = XdrGetFloat64(bytes[offset:])
	case NotificationString:
		value, used, err = XdrGetString(bytes[offset:])
	case NotificationOpaque:
		value, used, err = XdrGetOpaque(bytes[offset:])
	default:
		return nil, offset, errors.New("Marshalling failed: unknown element type")
	}

	if err != nil {
		return nil, 0, err
	}
	offset += used

	return value, offset, nil
}

// Put a value. If restrictToNotify is true then we'll fail on invalid types
func XdrPutValue(buffer *bytes.Buffer, value interface{}) {

	// Value
	switch value.(type) {
	case int32:
		XdrPutInt32(buffer, 1)
		XdrPutInt32(buffer, value.(int32))
	case int64:
		XdrPutInt32(buffer, 2)
		XdrPutInt64(buffer, value.(int64))
	case float64:
		XdrPutInt32(buffer, 3)
		XdrPutFloat64(buffer, value.(float64))
	case string:
		XdrPutInt32(buffer, 4)
		XdrPutString(buffer, value.(string))
	case []byte:
		XdrPutInt32(buffer, 5)
		XdrPutOpaque(buffer, value.([]uint8))
	default:
		panic(fmt.Sprintf("Bad *type* in XdrPutValue: %v", value))
	}
	return
}

// Get an xdr marshalled Elvin Notification
func XdrGetNotification(bytes []byte) (nfn map[string]interface{}, used int, err error) {
	nfn = make(map[string]interface{})
	offset := 0

	// Number of elements
	elementCount, used, err := XdrGetUint32(bytes[offset:])
	offset += used

	for elementCount > 0 {
		var name string // Avoid warning from go vet -shadow
		name, used, err = XdrGetString(bytes[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += used

		// The value
		nfn[name], used, err = XdrGetValue(bytes[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += used
		elementCount--
	}

	return nfn, offset, err
}

// Put an xdr marshalled Elvin Notification
func XdrPutNotification(buffer *bytes.Buffer, nfn map[string]interface{}) {

	// Number of elements
	XdrPutInt32(buffer, int32(len(nfn)))

	for k, v := range nfn {
		// Key
		XdrPutString(buffer, k)
		// Value
		XdrPutValue(buffer, v)
	}
	return
}

// Get an xdr marshalled list of Values
func XdrGetValues(bytes []byte) (values []interface{}, used int, err error) {
	offset := 0

	// Number of elements
	elementCount, used, err := XdrGetUint32(bytes[offset:])
	if err != nil {
		return nil, 0, err
	}
	offset += used

	v := make([]interface{}, elementCount)
	for i := 0; i < int(elementCount); i++ {
		v[i], used, err = XdrGetValue(bytes[offset:])
		offset += used
		if err != nil {
			return nil, 0, err
		}
	}

	return v, offset, nil
}

// Put an xdr marshalled list of Values
func XdrPutValues(buffer *bytes.Buffer, values []interface{}) {

	// Number of elements
	XdrPutInt32(buffer, int32(len(values)))

	for i := 0; i < len(values); i++ {
		XdrPutValue(buffer, values[i])
	}
	return
}

// Get an xdr marshalled keyset list
func XdrGetKeys(bytes []byte) (keyBlock KeyBlock, used int, err error) {
	offset := 0

	// Number of keysetlists
	kslCount, used, err := XdrGetInt32(bytes[offset:])
	if err != nil {
		return nil, 0, err
	}
	offset += used
	keyBlock = make(map[int]KeySetList)

	for i := 0; i < int(kslCount); i++ {
		// KeySetList
		scheme, used, err := XdrGetInt32(bytes[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += used

		ksCount, used, err := XdrGetInt32(bytes[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += used

		keyBlock[int(scheme)] = make([]KeySet, ksCount)

		for j := 0; j < int(ksCount); j++ {
			// Number of keys
			keyCount, used, err := XdrGetInt32(bytes[offset:])
			if err != nil {
				return nil, 0, err
			}
			offset += used

			// And finally the keys
			for k := 0; k < int(keyCount); k++ {
				key, used, err := XdrGetOpaque(bytes[offset:])
				if err != nil {
					return nil, 0, err
				}
				offset += used
				keyBlock[int(scheme)][j] = append(keyBlock[int(scheme)][j], key)
			}
		}
	}
	return keyBlock, offset, nil
}

// Put an xdr marshalled keyset list
func XdrPutKeys(buffer *bytes.Buffer, keyBlock KeyBlock) (err error) {
	// Number of KeySetLists
	XdrPutInt32(buffer, int32(len(keyBlock)))
	for scheme, ksl := range keyBlock {
		// The scheme
		XdrPutInt32(buffer, int32(scheme))

		// Number of keysets for this scheme
		XdrPutInt32(buffer, int32(len(ksl)))

		// Each KeySetList
		for i := 0; i < len(ksl); i++ {
			// Number of keys in KeySet
			XdrPutInt32(buffer, int32(len(ksl[i])))
			// The keys
			for j := 0; j < len(ksl[i]); j++ {
				XdrPutOpaque(buffer, ksl[i][j])
			}
		}
	}
	return nil
}
