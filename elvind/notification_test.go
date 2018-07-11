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
	"testing"
)

// I wanted to see how long it takes to create a router's Notification
// from a NotifyEmit to see if I should perhaps optimize. Seems less
// than a nanosecond
func BenchmarkFillNotification(b *testing.B) {
	nv := map[string]interface{}{"Benchmark creating a Notification from a NotifyEmit": int32(1)}
    ne := elvin.NotifyEmit{NameValue:nv, DeliverInsecure:true, Keys:nil}
	client := new(Client)
	var n Notification

	for i := 0; i < b.N; i++ {
		n = Notification{client.keysNfn, ne.NameValue, ne.DeliverInsecure, ne.Keys}
	}
	// Required to use n
	if n.Keys == nil {
		n.Keys = nil
	}

}
