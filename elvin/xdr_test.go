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
	"math"
	"reflect"
	"testing"
)

func TestXdrInt16(t *testing.T) {
	var b bytes.Buffer
	tests := []int16{1, math.MaxInt16, 0x7ead, 0x73b9}

	for _, test := range tests {
		XdrPutInt16(&b, test)
		i, used := XdrGetInt16(b.Bytes())
		if i != test {
			t.Fatalf("Marshal/Unmarshal of %d failed", test)
		}
		if used != 4 { // int32 on the wire
			t.Fatalf("Unmarshal of 16 bit type used %d bytes", used)
		}
		b.Reset()
	}
	return
}

func TestXdrUint16(t *testing.T) {
	var b bytes.Buffer
	tests := []uint16{1, math.MaxUint16, 0xdead, 0xfedc}
	for _, test := range tests {
		XdrPutUint16(&b, test)
		u, used := XdrGetUint16(b.Bytes())
		if u != test {
			t.Fatalf("Marshal/Unmarshal of %d failed", test)
		}
		if used != 4 { // int32 on the wire
			t.Fatalf("Unmarshal of 16 bit type used %d bytes", used)
		}
		b.Reset()
	}
	return
}

func TestXdrInt32(t *testing.T) {
	var b bytes.Buffer
	tests := []int{1, math.MaxInt32, 0x7eadbeef, 0x7373b9b9}

	for _, test := range tests {
		XdrPutInt32(&b, test)
		i, used := XdrGetInt32(b.Bytes())
		if i != test {
			t.Fatalf("Marshal/Unmarshal of %d failed", test)
		}
		if used != 4 {
			t.Fatalf("Unmarshal of 32 bit type used %d bytes", used)
		}
		b.Reset()
	}
	return
}

func TestXdrUint32(t *testing.T) {
	var b bytes.Buffer
	tests := []uint32{1, math.MaxUint32, 0xdeadbeef, 0xfedccdef}
	for _, test := range tests {
		XdrPutUint32(&b, test)
		u, used := XdrGetUint32(b.Bytes())
		if u != test {
			t.Fatalf("Marshal/Unmarshal of %d failed", test)
		}
		if used != 4 {
			t.Fatalf("Unmarshal of 32 bit type used %d bytes", used)
		}
		b.Reset()
	}
	return
}

func TestXdrInt64(t *testing.T) {
	var b bytes.Buffer
	tests := []int64{1, math.MaxInt64, 0x7eadbeefdeadbeef, 0x7373b9b9dada5151}

	for _, test := range tests {
		XdrPutInt64(&b, test)
		v, used := XdrGetInt64(b.Bytes())
		if v != test {
			t.Fatalf("Marshal/Unmarshal of %v->%v failed", test, v)
		}
		if used != 8 {
			t.Fatalf("Unmarshal of 64 bit type used %d bytes", used)
		}
		b.Reset()
	}
	return
}

func TestXdrUint64(t *testing.T) {
	var b bytes.Buffer
	tests := []uint64{1, math.MaxUint64, 0xdeadbeefdeadbeef, 0xfedccdef98766789}
	for _, test := range tests {
		XdrPutUint64(&b, test)
		v, used := XdrGetUint64(b.Bytes())
		if v != test {
			t.Fatalf("Marshal/Unmarshal of %v->%v failed", test, v)
		}
		if used != 8 {
			t.Fatalf("Unmarshal of 64 bit type used %d bytes", used)
		}
		b.Reset()
	}
	return
}

func TestXdrBool(t *testing.T) {
	var b bytes.Buffer

	tests := []bool{true, false}
	for _, test := range tests {
		XdrPutBool(&b, test)
		v, used := XdrGetBool(b.Bytes())
		if v != test {
			t.Fatalf("Marshal/Unmarshal of %v->%v failed", test, v)
		}
		if used != 4 { // int32 on the wire
			t.Fatalf("Unmarshal of 64 bit type used %d bytes", used)
		}
		b.Reset()
	}
	return
}

func TestXdrFloat64(t *testing.T) {
	var b bytes.Buffer

	tests := []float64{0, 1, math.Pi, math.E, float64(math.MaxFloat32), math.MaxFloat64}
	for _, test := range tests {
		XdrPutFloat64(&b, test)
		v, used := XdrGetFloat64(b.Bytes())
		if v != test {
			t.Fatalf("Marshal/Unmarshal of %v->%v failed", test, v)
		}
		if used != 8 {
			t.Fatalf("Unmarshal of 64 bit type used %d bytes", used)
		}
		b.Reset()
	}
	return
}

func TestXdrString(t *testing.T) {
	var b bytes.Buffer

	tests := []string{"", "a", "ab", "abc", "abcd", "abcde", ";kashf;kdhsaflkadsflkhasdlkfhladkshflkadhsflkhasdlfkhalskdhfksdjahfklasdhfklahdsfk9843y5043ryehfdlskhsdlkfy90834yrid;kafknzxcn@%$#%$@%$W&%^*T*(&&()*)"}
	for _, test := range tests {
		XdrPutString(&b, test)
		v, _ := XdrGetString(b.Bytes())
		if v != test {
			t.Fatalf("Marshal/Unmarshal of %v->%v failed", test, v)
		}
		b.Reset()
	}
	return
}

func TestXdrOpaque(t *testing.T) {
	var b bytes.Buffer

	tests := [][]byte{[]byte{}, []byte{0}, []byte{0, 1}, []byte{0, 1, 2}, []byte{0, 1, 2, 3}, []byte{0, 1, 2, 3, 127, 255}}
	for _, test := range tests {
		XdrPutOpaque(&b, test)
		v, _ := XdrGetOpaque(b.Bytes())
		if bytes.Compare(v, test) != 0 {
			t.Fatalf("Marshal/Unmarshal of %v->%v failed", test, v)
		}
		b.Reset()
	}
	return
}

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
