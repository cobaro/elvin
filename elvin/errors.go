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
)

// Provide a map of error code to string Each error string has a
// number of arguments that may be substituted as a hook for message
// localization
var ProtocolErrors map[int]string

func init() {
	ProtocolErrors = make(map[int]string)

	ProtocolErrors[ErrorsProtocolIncompatible] = "Version %1.%2 of the protocol is incompatible" // int int32
	ProtocolErrors[ErrorsAuthorizationFailure] = "Authorization failed"
	ProtocolErrors[ErrorsAuthenticationFailure] = "Authentication failed"

	ProtocolErrors[ErrorsProtocolError] = "Protocol Error"
	ProtocolErrors[ErrorsUnknownSubID] = "Unknown subscription id %1"    // uint64
	ProtocolErrors[ErrorsUnknownQuenchID] = "Unknown quench id %1"       // uint64
	ProtocolErrors[ErrorsBadKeyScheme] = "Bad key scheme %1"             // uint32
	ProtocolErrors[ErrorsBadKeysetIndex] = "Bad keyset index %1:%2"      // uint32, int32
	ProtocolErrors[ErrorsBadUTF8] = "Invalid UTF8 string at position %1" // int32, int32

	ProtocolErrors[ErrorsNoSuchKey] = "No such key"
	ProtocolErrors[ErrorsKeyExists] = "Key already exists"
	ProtocolErrors[ErrorsBadKey] = "Key is invalid"
	ProtocolErrors[ErrorsNothingToDo] = "Request contained no keys"
	ProtocolErrors[ErrorsQOSLimit] = "Request out of bounds: %1" // string
	ProtocolErrors[ErrorsImplementationLimit] = "Request out of range"
	ProtocolErrors[ErrorsNotImplemented] = "Request unimplemented"

	ProtocolErrors[ErrorsParsing] = "Parse error before %2 at position %1" // string, int32

	ProtocolErrors[ErrorsInvalidToken] = "Parse error at token %1 offset %2"               // string, int32
	ProtocolErrors[ErrorsUnterminatedString] = "Unterminated string at offset %1"          // int32
	ProtocolErrors[ErrorsUnknownFunction] = "Unknown function at offset %1"                // int32
	ProtocolErrors[ErrorsOverflow] = "Numeric constant overflow at offset %1"              // int32
	ProtocolErrors[ErrorsTypeMismatch] = "Type mismatch between %1 and %2 at offset %3"    // string, string, int32
	ProtocolErrors[ErrorsTooFewArgs] = "Not enough argument to function %1() at offset %2" // string, int32
	ProtocolErrors[ErrorsInvalidRegexp] = "Bad regular expression %1 at offset %2"         // string, int32
	ProtocolErrors[ErrorsExpIsTrivial] = "Expression compiled to a constant value"
	ProtocolErrors[ErrorsRegexpTooComplex] = "Expression %1 too complex at position %2"        // string, int32
	ProtocolErrors[ErrorsNestingTooDeep] = "Expression %1 has nests too deeply at position %2" // string, int32

	ProtocolErrors[ErrorsQuenchEmpty] = "Quench has no attribute names"
	ProtocolErrors[ErrorsQuenchAttributeExists] = "Attribute %1 already present" // string
	ProtocolErrors[ErrorsQuenchNoSuchAttribute] = "No such attribute: %1"        // string
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
				str = append(str, '[')
				str = append(str, in[i+1])
				str = append(str, ']')
				str = append(str, 'v')
				i += 1
			}
		}
	}
	return string(str)
}
