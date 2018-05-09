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
	"reflect"
	"testing"
)

func TestXdrNotification(t *testing.T) {
	nfn := make(map[string]interface{})

	nfn["int32"] = 3232
	nfn["int64"] = int64(646464646464)
	nfn["string"] = "string"
	nfn["opaque"] = []byte{0, 1, 2, 3, 127, 255}
	nfn["float64"] = 424242.42

	// encode
	var buffer = new(bytes.Buffer)
	XdrPutNotification(buffer, nfn)
	t.Logf("%d:%v\n", buffer.Len(), buffer.Bytes())
	bytes := buffer.Bytes()
	t.Logf("%d: %v", len(bytes), bytes)

	// decode
	nfn2, _, err := XdrGetNotification(bytes)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log("nfn1:", nfn)
	t.Log("nfn2:", nfn2)
	// The floats won't be perfectly equal but should be close
	if nfn["float64"].(float64)-nfn2["float64"].(float64) > 0.000001 {
		t.Logf("Floats differ by too much, %v!=%v",
			nfn["float64"].(float64), nfn2["float64"].(float64))
		t.Fail()
	}

	if nfn["int64"].(int64) != nfn2["int64"].(int64) {
		t.Log("int64s differ")
		t.Fail()
	}

	if nfn["int32"].(int) != nfn2["int32"].(int) {
		t.Log("int32s differ")
		t.Fail()
	}
	if nfn["string"].(string) != nfn2["string"].(string) {
		t.Log("strings differ")
		t.Fail()
	}
	o1 := nfn["opaque"].([]byte)
	o2 := nfn2["opaque"].([]byte)
	if len(o1) != len(o2) {
		t.Log("opaques differ")
		t.Fail()
	}
	for i := 0; i < len(o1); i++ {
		if o1[i] != o2[i] {
			t.Log("opaques differ")
			t.Fail()
		}
	}
}

func TestXdrKeys(t *testing.T) {
	var buffer = new(bytes.Buffer)
	var kl1 []Keyset = []Keyset{
		Keyset{1, [][]byte{{254, 220, 0, 17}, {1, 2, 3, 4}}},
		Keyset{3, [][]byte{{1, 2, 3, 4}}}}
	XdrPutKeys(buffer, kl1)
	expected := buffer.Len()
	kl2, used, _ := XdrGetKeys(buffer.Bytes())
	if used != expected {
		t.Log("Encode/Decode of Keylists had different lengths")
		t.Fail()
	}

	if !reflect.DeepEqual(kl1, kl2) {
		t.Log("Keys differ\n", kl1, "\n", kl2)
		t.Fail()
	}
	t.Log("\n", kl1, "\n", kl2)
}
