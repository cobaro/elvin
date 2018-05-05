// Copyright 2018 Cobaro Pty Ltd. All Rights Reserved.

// This file is part of elvin
//
// elvin is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// elvin is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with elvin. If not, see <http://www.gnu.org/licenses/>.

package elvin

import (
	"bytes"
	"testing"
)

func TestXdr(t *testing.T) {
	nfn := make(map[string]interface{})

	nfn["int32"] = 3232
	nfn["int64"] = int64(64646464)
	nfn["string"] = "string"
	nfn["opaque"] = []byte{3}
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
