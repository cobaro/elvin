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

// Keys are a set of Keysets each with a scheme
const (
	KeySchemeSha1Dual       = 1 // The SHA-1 dual key scheme
	KeySchemeSha1Producer   = 2 // The SHA-1 producer key scheme
	KeySchemeSha1Consumer   = 3 // The SHA-1 consumer key scheme
	KeySchemeSha256Dual     = 7 // The SHA-256 dual key scheme
	KeySchemeSha256Producer = 8 // The SHA-256 producer key scheme
	KeySchemeSha256Consumer = 9 // The SHA-256 consumer key scheme
)

type Keyset struct {
	KeyScheme int
	Keysets   [][]byte
}
