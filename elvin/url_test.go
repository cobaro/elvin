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
	"testing"
)

func TestURLToProtocol(t *testing.T) {

	passingTests := []string{
		"elvin://",
		"elvin:4.1//",
		"elvin:4.1/tcp,xdr/",
		"elvin:4.1/tcp,none,xdr/",
		"elvin:4.1/tcp,none,xdr/host",
		"elvin:4.1/tcp,none,xdr/host:2917",
		"elvin:4.1/tcp,none,xdr/:2917",
		"elvin:4.1//host:2917",
		"elvin:42.42//host:2917",
		"elvin://host:2917/foo/bar",
	}

	failingTests := []string{
		"elvin",
		" elvin:",
		"elvin:x/",
		"elvin:x//",
		"elvin:4-1//",
		"elvin:/one/",
		"elvin:/one,two,three,four/",
		"elvin://host:",
		"elvin://host:notanumber",
		"elvin://host:2917:extra",
		"elvin://:",
	}

	for _, test := range passingTests {

		if _, err := URLToProtocol(test); err != nil {
			t.Fatalf("Parse failed for: %s (%v)", test, err)
		}

	}

	for _, test := range failingTests {
		if _, err := URLToProtocol(test); err == nil {
			t.Fatalf("Parse succeeded for: %s (%v)", test, err)
		}
	}
	return
}

func TestProtocolToURL(t *testing.T) {
	protocol := Protocol{"tcp", "xdr", "localhost:2917", "args", 4, 1}
	expect := "elvin:4.1/tcp,xdr/localhost:2917/args"
	get := ProtocolToURL(&protocol)
	if expect != get {
		t.Fatalf("%s != %s", expect, get)
	}
}
