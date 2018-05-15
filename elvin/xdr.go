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

// Get an xdr marshalled bool (on the wire it's a 32 bit signed int)
func XdrGetBool(bytes []byte) (b bool, used int) {
	i, used := XdrGetInt32(bytes)
	b = (i == 0)
	return b, used
}

// Put an xdr marshalled bool (ont the wire it's a 32 bit signed int)
func XdrPutBool(buffer *bytes.Buffer, b bool) {
	var i int = 0
	if b {
		i = 1
	}
	XdrPutInt32(buffer, i)
	return
}

// Get an xdr marshalled int16 (on the wire it's a 32 bit signed int)
func XdrGetInt16(bytes []byte) (i16 int16, used int) {
	i, used := XdrGetInt32(bytes)
	return int16(i), used
}

// Put an xdr marshalled int16 (ont the wire it's a 32 bit signed int)
func XdrPutInt16(buffer *bytes.Buffer, i16 int16) {
	XdrPutInt32(buffer, int(i16))
	return
}

// Get an xdr marshalled uint16 (on the wire it's a 32 bit unsigned int)
func XdrGetUint16(bytes []byte) (u16 uint16, used int) {
	u, used := XdrGetUint32(bytes)
	return uint16(u), used
}

// Put an xdr marshalled uint16 (ont the wire it's a 32 bit unsigned int)
func XdrPutUint16(buffer *bytes.Buffer, u16 uint16) {
	XdrPutUint32(buffer, uint32(u16))
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
func XdrGetValue(bytes []byte) (val interface{}, used int, err error) {

	// Type of value
	offset := 0
	elementType, used := XdrGetInt32(bytes[offset:])
	offset += used
	var value interface{}

	// The values itself
	switch elementType {
	case NotificationInt32:
		value, used = XdrGetInt32(bytes[offset:])
		offset += used
		break

	case NotificationInt64:
		value, used = XdrGetInt64(bytes[offset:])
		offset += used
		break

	case NotificationFloat64:
		value, used = XdrGetFloat64(bytes[offset:])
		offset += used
		break

	case NotificationString:
		value, used = XdrGetString(bytes[offset:])
		offset += used
		break

	case NotificationOpaque:
		value, used = XdrGetOpaque(bytes[offset:])
		offset += used
		break

	default:
		return nil, offset, errors.New("Marshalling failed: unknown element type")
	}
	return value, offset, nil
}

// Put a value
func XdrPutValue(buffer *bytes.Buffer, value interface{}) {

	// Value
	switch typ := value.(type) {
	case int:
		XdrPutInt32(buffer, 1)
		XdrPutInt32(buffer, value.(int))
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
		XdrPutInt32(buffer, 0)
		// FIXME: This seems harsh in it's be strict what you send
		panic(fmt.Sprintf("What *type* is: %v", typ))
	}
	return
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

		// The value
		nfn[name], used, err = XdrGetValue(bytes[offset:])
		offset += used
		if err != nil {
			return nil, offset, err
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
		XdrPutValue(buffer, v)
	}
	return
}

// Get an xdr marshalled list of Values
func XdrGetValues(bytes []byte) (values []interface{}, used int, err error) {
	offset := 0

	// Number of elements
	elementCount, used := XdrGetUint32(bytes[offset:])
	offset += used

	v := make([]interface{}, elementCount)
	for i := 0; i < int(elementCount); i++ {
		v[i], used, err = XdrGetValue(bytes[offset:])
		offset += used
		if err != nil {
			return nil, offset, err
		}
	}

	return v, offset, nil
}

// Put an xdr marshalled list of Values
func XdrPutValues(buffer *bytes.Buffer, values []interface{}) {

	// Number of elements
	XdrPutInt32(buffer, len(values))

	for i := 0; i < len(values); i++ {
		XdrPutValue(buffer, values[i])
	}
	return
}

// Get an xdr marshalled keyset list
func XdrGetKeys(bytes []byte) (kl []Keyset, used int, err error) {
	offset := 0

	// Number of keylists
	listCount, used := XdrGetInt32(bytes[offset:])
	offset += used
	keylists := make([]Keyset, listCount)

	for i := 0; i < listCount; i++ {
		// Key scheme
		keylists[i].KeyScheme, used = XdrGetInt32(bytes[offset:])
		offset += used

		// Number of sets
		setCount, used := XdrGetInt32(bytes[offset:])
		offset += used

		keylists[i].Keysets = make([][]byte, setCount)

		// fmt.Println("want", setCount, "sets from", bytes[offset:], keylists)
		for j := 0; j < setCount; j++ {
			// Number of keys
			keyCount, used := XdrGetInt32(bytes[offset:])
			offset += used

			for k := 0; k < keyCount; k++ {
				keylists[i].Keysets[j], used = XdrGetOpaque(bytes[offset:])
				offset += used
			}
		}
	}
	return keylists, offset, nil
}

// Put an xdr marshalled keyset list
func XdrPutKeys(buffer *bytes.Buffer, kl []Keyset) (err error) {
	//fmt.Println("keylist:", kl)

	// Number of keylists
	XdrPutInt32(buffer, len(kl))
	for i := 0; i < len(kl); i++ {
		// The scheme
		XdrPutInt32(buffer, kl[i].KeyScheme)

		// Number of keysets
		// fmt.Println("keysets:", len(kl[i].Keysets))
		XdrPutInt32(buffer, len(kl[i].Keysets))

		// Each keyset
		for j := 0; j < len(kl[i].Keysets); j++ {
			// Number of keys
			XdrPutInt32(buffer, len(kl[i].Keysets[j]))
			// The keys
			for k := 0; k < len(kl[i].Keysets[j]); k++ {
				XdrPutOpaque(buffer, kl[i].Keysets[j])
			}
		}
	}
	return nil // FIXME: things can go wrong here
}
