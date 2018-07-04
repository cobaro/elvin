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
	"encoding/hex"
	"testing"
)

func TestPrime(t *testing.T) {
	var in = []byte("foo")
	// echo -n foo | sha1sum
	expect := "0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33"
	out := PrimeSha1(in)
	if expect != hex.EncodeToString(out) {
		t.Fatalf("SHA1(%s) != %s (got %s)", in, string(expect), hex.EncodeToString(out))
	}

	// echo -n foo | sha256sum
	expect = "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"
	out = PrimeSha256(in)
	if expect != hex.EncodeToString(out) {
		t.Fatalf("SHA256(%s) != %s (got %s)", in, string(expect), hex.EncodeToString(out))
	}

}

func TestKeySet(t *testing.T) {
	var ks KeySet

	KeySetAddKey(&ks, []byte("foo"))
	if len(ks) != 1 {
		t.Fatalf("AddKey() failed")
	}

	KeySetAddKey(&ks, []byte("foo"))
	if len(ks) != 1 {
		t.Fatalf("AddKey() failed")
	}

	KeySetDeleteKey(&ks, []byte("NOTfoo"))
	if len(ks) != 1 {
		t.Fatalf("DeleteKey() failed")
	}

	KeySetDeleteKey(&ks, []byte("foo"))
	if len(ks) != 0 {
		t.Fatalf("DeleteKey() failed")
	}
}
