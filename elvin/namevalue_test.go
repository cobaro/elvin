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
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

var in map[string]interface{}

func init() {
	in = make(map[string]interface{})
	in["int32"] = int32(32)
	in["int64"] = int64(6464646464646464)
	in["float"] = 3.1415
	in["string"] = "I am a string"
	in["opaque"] = []byte{00, 33, 66, 99, 133, 255}
}

func TestNameValue(t *testing.T) {
	str, err := NameValueToString(in, "", "")
	if err != nil {
		t.Fatalf("NameValueToString failed: %v", err)
	}

	t.Logf("\n%s", str)

	nfns := ParseNotifications(strings.NewReader(str), os.Stderr, 0, fmt.Fprintf)
	out := <-nfns

	if len(in) != len(out) {
		t.Fatalf("input and output maps differ in length")
	}

	for k, v := range in {
		if !reflect.DeepEqual(v, out[k]) {
			t.Fatalf("input (%v) and output (%v) values differ", v, out[k])
		}
	}
}

func TestNameValueDates(t *testing.T) {
	str, _ := NameValueToString(nil, "", DefaultNameValueTimeFormat)
	if str[:6] != "$time " {
		t.Fatalf("No timestamp: %v", str)
	}

	// Check our own timeformat works
	if _, err := time.Parse(DefaultNameValueTimeFormat, DefaultNameValueTimeFormat); err != nil {
		t.Fatalf("%v", err)
	}

	// The old DSTC elvin-utils/ec for compatibility
	if _, err := time.Parse(DefaultNameValueTimeFormat, "2001-02-19T14:48:38.836165+1000"); err != nil {
		t.Fatalf("%v", err)
	}

	// Now check to/from are a reasonably close time
	str, err := NameValueToString(nil, "", DefaultNameValueTimeFormat)
	if err != nil {
		t.Fatalf("%v", err)
	}
	first := strings.Split(str[6:], "\n")[0]

	then, err := time.Parse(DefaultNameValueTimeFormat, first)
	if err != nil {
		t.Fatalf("Failed to parse date from: %s", first)
	}

	now := time.Now()
	duration := now.Sub(then)
	if duration > time.Millisecond {
		t.Fatalf("Timing seems out: duration was %v", duration)
	}
}
