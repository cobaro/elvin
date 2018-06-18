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

package main

import (
	"github.com/cobaro/elvin/elvin"
)

// A Subscription
type Subscription struct {
	SubID          int64
	AcceptInsecure bool
	Keys           elvin.KeyBlock
	Ast            *elvin.AST
}

// Parse a subscription expression into an AST
func Parse(subexpr string) (ast *elvin.AST, n *elvin.Nack) {
	// For now 'bogus' fails and everything else succeeds
	if subexpr == "bogus" {
		nack := new(elvin.Nack)
		nack.ErrorCode = elvin.ErrorsParsing
		nack.Message = elvin.ProtocolErrors[elvin.ErrorsParsing].Message
		nack.Args = make([]interface{}, 2)
		nack.Args[0] = "bogus"
		nack.Args[1] = int32(0)
		return nil, nack
	}
	return nil, nil
}
