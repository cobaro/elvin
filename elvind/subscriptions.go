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
	"sync"
)

// FIXME:dummy
type Ast int

// A Subscription
type Subscription struct {
	Subid          uint64
	AcceptInsecure bool
	Keys           []elvin.Keyset
	Ast            Ast
}

// Map of subscriptions to clients
type Subscriptions struct {
	subscriptions map[uint64]*Connection // initialized in init()
	current       uint64                 // Subscription Ids are simply globally incremented as 64 bits is sufficient
	lock          sync.Mutex             // initialized aautomatically
}

// Global subscriptions
var subscriptions Subscriptions

func init() {
	subscriptions.subscriptions = make(map[uint64]*Connection)
}

// Parse a subscription expression into an AST
func Parse(subexpr string) (ast Ast, n *elvin.Nack) {
	// For now 'bogus' fails and everything else succeeds
	if subexpr == "bogus" {
		nack := new(elvin.Nack)
		nack.ErrorCode = elvin.ErrorsParsing
		nack.Message = elvin.ProtocolErrors[elvin.ErrorsParsing]
		nack.Args = make([]interface{}, 2)
		nack.Args[0] = 0
		nack.Args[1] = "bogus"
		return 0, nack
	}
	return 0, nil
}

// Handle add a subscription into our global subscription engine
func (sub *Subscription) Add(conn *Connection) {
	subscriptions.lock.Lock()

	sub.Subid = subscriptions.current
	subscriptions.current++
	subscriptions.subscriptions[sub.Subid] = conn

	subscriptions.lock.Unlock()
}
