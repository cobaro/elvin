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
	"strings"
	"testing"
)

func TestFmt(t *testing.T) {

	//fmt.Printf("Arg 2:%[2]v, Arg 1:%[1]v, Arg 2:%[2]v", "arg1", "arg2")
	var in, out, expected string

	in = "100% quick brown fox"
	expected = "100% quick brown fox"
	out = ElvinStringToFormatString(in)
	if 0 != strings.Compare(out, expected) {
		t.Fatalf("%v -> %v, expected %v", in, out, expected)
	}

	in = "quick brown fox 100%"
	expected = "quick brown fox 100%"
	out = ElvinStringToFormatString(in)
	if 0 != strings.Compare(out, expected) {
		t.Fatalf("%v -> %v, expected %v", in, out, expected)
	}

	in = "% "
	expected = "% "
	out = ElvinStringToFormatString(in)
	if 0 != strings.Compare(out, expected) {
		t.Fatalf("%v -> %v, expected %v", in, out, expected)
	}

	in = "%3"
	expected = "%[3]v"
	out = ElvinStringToFormatString(in)
	if 0 != strings.Compare(out, expected) {
		t.Fatalf("%v -> %v, expected %v", in, out, expected)
	}

	in = "%%"
	expected = "%%"
	out = ElvinStringToFormatString(in)
	if 0 != strings.Compare(out, expected) {
		t.Fatalf("%v -> %v, expected %v", in, out, expected)
	}

	in = "%% %2 %1 %%"
	expected = "%% %[2]v %[1]v %%"
	out = ElvinStringToFormatString(in)
	if 0 != strings.Compare(out, expected) {
		t.Fatalf("%v -> %v, expected %v", in, out, expected)
	}

	in = "%1 %3 %5 5% %7"
	expected = "%[1]v %[3]v %[5]v 5% %[7]v"
	out = ElvinStringToFormatString(in)
	if 0 != strings.Compare(out, expected) {
		t.Fatalf("%v -> %v, expected %v", in, out, expected)
	}
}
