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
	"fmt"
)

// Elvin defines the following error codes
const (
	// 1-999 Connection establishment
	ErrorsProtocolIncompatible  = 1
	ErrorsAuthorizationFailure  = 2
	ErrorsAuthenticationFailure = 3
	// 4-499 reserved for defined protocol errors
	// 500-999 reserved for implementation specific errors

	// 1000-1999 Protocol errors caused by connection corruption
	// or implementation failure
	ErrorsProtocolError   = 1001
	ErrorsUnknownSubID    = 1002
	ErrorsUnknownQuenchID = 1003
	ErrorsBadKeyScheme    = 1004
	ErrorsBadKeysetIndex  = 1005
	ErrorsBadUTF8         = 1006

	// 2000-2999 An error detected in a request
	ErrorsNoSuchKey           = 2001
	ErrorsKeyExists           = 2002
	ErrorsBadKey              = 2003
	ErrorsNothingToDo         = 2004
	ErrorsQOSLimit            = 2005
	ErrorsImplementationLimit = 2006
	ErrorsNotImplemented      = 2007
	// 2008-2100 reserved

	ErrorsParsing            = 2101
	ErrorsInvalidToken       = 2102
	ErrorsUnterminatedString = 2103
	ErrorsUnknownFunction    = 2104
	ErrorsOverflow           = 2105
	ErrorsTypeMismatch       = 2106
	ErrorsTooFewArgs         = 2107
	ErrorsInvalidRegexp      = 2109
	ErrorsExpIsTrivial       = 2110
	ErrorsRegexpTooComplex   = 2111
	ErrorsNestingTooDeep     = 2112
	// 2113-2200 reserved

	ErrorsQuenchEmpty           = 2201
	ErrorsQuenchAttributeExists = 2202
	ErrorsQuenchNoSuchAttribute = 2203
	// 2204-2499 reserved for defined errors

	// 2500-2999 reserved for implementation specific errors
	// golang local client library errors
	ErrorsTimeout                         = 2500
	ErrorsBadPacket                       = 2501
	ErrorsBadPacketType                   = 2502
	ErrorsMismatchedXIDs                  = 2503
	ErrorsClientNotConnected              = 2504
	ErrorsClientIsConnected               = 2505
	ErrorsClientConnecting                = 2506
	ErrorsClientDisconnecting             = 2507
	ErrorsProtocolPacketStateNotConnected = 2508
	ErrorsProtocolPacketStateIsConnected  = 2509
)

// Provide a map of error code to string Each error string has a
// number of arguments that may be substituted as a hook for message
// localization
const MaxNackArgs = 3

type NackArgs struct {
	Message string
	NumArgs int
	Args    [MaxNackArgs]interface{}
}

var ProtocolErrors map[uint16]NackArgs
var LocalErrors map[uint16]string

// Note that the spec has some unsigned types specified here but also
// wants to marshall them as a Value that only supports signed types.
// This implementation resolves this by using signed types for both
// subscription and quench IDs
func init() {
	ProtocolErrors = make(map[uint16]NackArgs)

	ProtocolErrors[ErrorsProtocolIncompatible] = NackArgs{"Incompatible protocol version", 0, [MaxNackArgs]interface{}{nil, nil, nil}}
	ProtocolErrors[ErrorsAuthorizationFailure] = NackArgs{"Authorization failed", 0, [MaxNackArgs]interface{}{nil, nil, nil}}
	ProtocolErrors[ErrorsAuthenticationFailure] = NackArgs{"Authentication failed", 0, [MaxNackArgs]interface{}{nil, nil, nil}}

	ProtocolErrors[ErrorsProtocolError] = NackArgs{"Protocol Error", 0, [MaxNackArgs]interface{}{nil, nil, nil}}
	ProtocolErrors[ErrorsUnknownSubID] = NackArgs{"Unknown subscription id %1", 1, [MaxNackArgs]interface{}{int64(0), nil, nil}}
	ProtocolErrors[ErrorsUnknownQuenchID] = NackArgs{"Unknown quench id %1", 1, [MaxNackArgs]interface{}{int64(0), nil, nil}}
	ProtocolErrors[ErrorsBadKeyScheme] = NackArgs{"Bad key scheme %1", 1, [MaxNackArgs]interface{}{int32(0), nil, nil}}
	ProtocolErrors[ErrorsBadKeysetIndex] = NackArgs{"Bad keyset index %1:%2", 2, [MaxNackArgs]interface{}{int32(0), int32(0), nil}}
	ProtocolErrors[ErrorsBadUTF8] = NackArgs{"Invalid UTF8 string at position %1", 2, [MaxNackArgs]interface{}{int32(0), nil, nil}}

	ProtocolErrors[ErrorsNoSuchKey] = NackArgs{"No such key", 0, [MaxNackArgs]interface{}{nil, nil, nil}}
	ProtocolErrors[ErrorsKeyExists] = NackArgs{"Key already exists", 0, [MaxNackArgs]interface{}{nil, nil, nil}}
	ProtocolErrors[ErrorsBadKey] = NackArgs{"Key is invalid", 0, [MaxNackArgs]interface{}{nil, nil, nil}}
	ProtocolErrors[ErrorsNothingToDo] = NackArgs{"Request contained no keys", 0, [MaxNackArgs]interface{}{nil, nil, nil}}
	ProtocolErrors[ErrorsQOSLimit] = NackArgs{"Request out of bounds: %1", 1, [MaxNackArgs]interface{}{"", nil, nil}}
	ProtocolErrors[ErrorsImplementationLimit] = NackArgs{"Request out of range", 0, [MaxNackArgs]interface{}{nil, nil, nil}}
	ProtocolErrors[ErrorsNotImplemented] = NackArgs{"Request unimplemented", 0, [MaxNackArgs]interface{}{nil, nil, nil}}

	ProtocolErrors[ErrorsParsing] = NackArgs{"Parse error before %1 at position %2", 2, [MaxNackArgs]interface{}{int32(0), "", nil}}
	ProtocolErrors[ErrorsInvalidToken] = NackArgs{"Parse error at token %1 offset %2", 2, [MaxNackArgs]interface{}{"", int32(0), nil}}
	ProtocolErrors[ErrorsUnterminatedString] = NackArgs{"Unterminated string at offset %1", 1, [MaxNackArgs]interface{}{int32(0), nil, nil}}
	ProtocolErrors[ErrorsUnknownFunction] = NackArgs{"Unknown function at offset %1", 1, [MaxNackArgs]interface{}{int32(0), nil, nil}}
	ProtocolErrors[ErrorsOverflow] = NackArgs{"Numeric constant overflow at offset %1", 1, [MaxNackArgs]interface{}{int32(0), nil, nil}}
	ProtocolErrors[ErrorsTypeMismatch] = NackArgs{"Type mismatch between %1 and %2 at offset %3", 3, [MaxNackArgs]interface{}{"", "", int32(0)}}
	ProtocolErrors[ErrorsTooFewArgs] = NackArgs{"Not enough argument to function %1() at offset %2", 2, [MaxNackArgs]interface{}{"", int32(0), nil}}
	ProtocolErrors[ErrorsInvalidRegexp] = NackArgs{"Bad regular expression %1 at offset %2", 2, [MaxNackArgs]interface{}{"", int32(0), nil}}
	ProtocolErrors[ErrorsExpIsTrivial] = NackArgs{"Expression compiled to a constant value", 0, [MaxNackArgs]interface{}{nil, nil, nil}}
	ProtocolErrors[ErrorsRegexpTooComplex] = NackArgs{"Expression %1 too complex at position %2", 2, [MaxNackArgs]interface{}{"", int32(0), nil}}
	ProtocolErrors[ErrorsNestingTooDeep] = NackArgs{"Expression %1 has nests too deeply at position %2", 2, [MaxNackArgs]interface{}{"", int32(0), nil}}

	ProtocolErrors[ErrorsQuenchEmpty] = NackArgs{"Quench has no attribute names", 0, [MaxNackArgs]interface{}{nil, nil, nil}}
	ProtocolErrors[ErrorsQuenchAttributeExists] = NackArgs{"Attribute %1 already present", 1, [MaxNackArgs]interface{}{"", nil, nil}}
	ProtocolErrors[ErrorsQuenchNoSuchAttribute] = NackArgs{"No such attribute: %1", 1, [MaxNackArgs]interface{}{"", nil, nil}}

	// Local errors
	LocalErrors = make(map[uint16]string)

	LocalErrors[ErrorsTimeout] = "Timeout waiting for response"
	LocalErrors[ErrorsBadPacket] = "Unexpected packet"
	LocalErrors[ErrorsBadPacketType] = "Unexpected packet: %1"
	LocalErrors[ErrorsMismatchedXIDs] = "Unable to match transaction IDs, expected:%1, received:%2"
	LocalErrors[ErrorsClientNotConnected] = "Client is not connected"
	LocalErrors[ErrorsClientIsConnected] = "Client is connected"
	LocalErrors[ErrorsProtocolPacketStateNotConnected] = "Protocol Error. Received %1 when not connected"
	LocalErrors[ErrorsProtocolPacketStateIsConnected] = "Protocol Error. Received %1 when connected"
}

// Convert elvin positional formatting to golang style
// %n -> %[n]v
// %% -> %%
func ElvinStringToFormatString(in string) (out string) {
	var str []byte

	for i := 0; i < len(in); i++ {
		str = append(str, in[i])
		if in[i] == '%' && i < len(in)-1 {
			if in[i+1] == '%' {
				str = append(str, '%')
				i++
			} else if in[i+1] >= '0' && in[i+1] <= '9' {
				str = append(str, '[', in[i+1], ']', 'v')
				i += 1
			}
		}
	}
	return string(str)
}

func LocalError(code uint16, args ...interface{}) (err error) {
	return fmt.Errorf("[%d] %s", code, fmt.Sprintf(ElvinStringToFormatString(LocalErrors[code]), args...))
}

func NackError(nack Nack) (err error) {
	return fmt.Errorf(nack.String())
}
