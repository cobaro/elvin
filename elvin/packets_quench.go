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
	"fmt"
)

type SubAST struct {
	dummy int
}

type QnchAddRqst struct {
	XID             uint32
	Names           map[string]bool
	DeliverInsecure bool
	Keys            []Keyset
}

func (pkt *QnchAddRqst) ID() int {
	return PacketQnchAddRqst
}

func (pkt *QnchAddRqst) IDString() string {
	return "QnchAddRqst"
}

func (pkt *QnchAddRqst) IString(indent string) string {
	return fmt.Sprintf(
		"%sXID: %d\n"+
			"%sNames: %v\n"+
			"%sDeliverInsecure: %v\n"+
			"%sKeys: %v\n",
		indent, pkt.XID,
		indent, pkt.Names,
		indent, pkt.DeliverInsecure,
		indent, pkt.Keys)
}

func (pkt *QnchAddRqst) String() string {
	return pkt.IString("")
}

func (pkt *QnchAddRqst) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	nameCount, used, err := XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Names = make(map[string]bool)
	for i := uint32(0); i < nameCount; i++ {
		name, used, err := XdrGetString(bytes[offset:])
		if err != nil {
			return err
		}
		pkt.Names[name] = true
		offset += used
	}

	pkt.DeliverInsecure, used, err = XdrGetBool(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.Keys, used, err = XdrGetKeys(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

func (pkt *QnchAddRqst) Encode(buffer *bytes.Buffer) (xID uint32) {
	xID = XID()
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutInt32(buffer, int32(xID))
	XdrPutUint32(buffer, uint32(len(pkt.Names)))
	for name, _ := range pkt.Names {
		XdrPutString(buffer, name)
	}
	XdrPutBool(buffer, pkt.DeliverInsecure)
	XdrPutKeys(buffer, pkt.Keys)

	return
}

type QnchModRqst struct {
	XID             uint32
	QuenchID        int64
	AddNames        map[string]bool
	DelNames        map[string]bool
	DeliverInsecure bool
	AddKeys         []Keyset
	DelKeys         []Keyset
}

func (pkt *QnchModRqst) ID() int {
	return PacketQnchModRqst
}

func (pkt *QnchModRqst) IDString() string {
	return "QnchModRqst"
}

func (pkt *QnchModRqst) IString(indent string) string {
	return fmt.Sprintf(
		"%sXID: %d\n"+
			"%sQuenchID: %d\n"+
			"%sAddNames: %v\n"+
			"%sDelNames: %v\n"+
			"%sDeliverInsecure: %v\n"+
			"%sAddKeys: %v\n"+
			"%sDelKeys: %v\n",
		indent, pkt.XID,
		indent, pkt.QuenchID,
		indent, pkt.AddNames,
		indent, pkt.DelNames,
		indent, pkt.DeliverInsecure,
		indent, pkt.AddKeys,
		indent, pkt.DelKeys)
}

func (pkt *QnchModRqst) String() string {
	return pkt.IString("")
}

func (pkt *QnchModRqst) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.QuenchID, used, err = XdrGetInt64(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	addNamesCount, used, err := XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.AddNames = make(map[string]bool)
	for i := uint32(0); i < addNamesCount; i++ {
		name, used, err := XdrGetString(bytes[offset:])
		if err != nil {
			return err
		}
		pkt.AddNames[name] = true
		offset += used
	}

	delNamesCount, used, err := XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.DelNames = make(map[string]bool)
	for i := uint32(0); i < delNamesCount; i++ {
		name, used, err := XdrGetString(bytes[offset:])
		if err != nil {
			return err
		}
		pkt.DelNames[name] = true
		offset += used
	}

	pkt.DeliverInsecure, used, err = XdrGetBool(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.AddKeys, used, err = XdrGetKeys(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.DelKeys, used, err = XdrGetKeys(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

func (pkt *QnchModRqst) Encode(buffer *bytes.Buffer) (xID uint32) {
	xID = XID()
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutInt32(buffer, int32(xID))
	XdrPutInt64(buffer, pkt.QuenchID)
	XdrPutUint32(buffer, uint32(len(pkt.AddNames)))
	for name, _ := range pkt.AddNames {
		XdrPutString(buffer, name)
	}
	XdrPutUint32(buffer, uint32(len(pkt.DelNames)))
	for name, _ := range pkt.DelNames {
		XdrPutString(buffer, name)
	}
	XdrPutBool(buffer, pkt.DeliverInsecure)
	XdrPutKeys(buffer, pkt.AddKeys)
	XdrPutKeys(buffer, pkt.DelKeys)

	return
}

type QnchDelRqst struct {
	XID      uint32
	QuenchID int64
}

func (pkt *QnchDelRqst) ID() int {
	return PacketQnchDelRqst
}

func (pkt *QnchDelRqst) IDString() string {
	return "QnchDelRqst"
}

func (pkt *QnchDelRqst) IString(indent string) string {
	return fmt.Sprintf(
		"%sXID: %d\n"+
			"%sQuenchID: %d\n",
		indent, pkt.XID,
		indent, pkt.QuenchID)
}

func (pkt *QnchDelRqst) String() string {
	return pkt.IString("")
}

func (pkt *QnchDelRqst) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.QuenchID, used, err = XdrGetInt64(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

func (pkt *QnchDelRqst) Encode(buffer *bytes.Buffer) (xID uint32) {
	xID = XID()
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutInt32(buffer, int32(xID))
	XdrPutInt64(buffer, pkt.QuenchID)

	return
}

type QnchRply struct {
	XID      uint32
	QuenchID int64
}

func (pkt *QnchRply) ID() int {
	return PacketQnchRply
}

func (pkt *QnchRply) IDString() string {
	return "QnchRply"
}

func (pkt *QnchRply) IString(indent string) string {
	return fmt.Sprintf(
		"%sXID: %d\n"+
			"%sQuenchID: %d\n",
		indent, pkt.XID,
		indent, pkt.QuenchID)
}

func (pkt *QnchRply) String() string {
	return pkt.IString("")
}

func (pkt *QnchRply) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.XID, used, err = XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	pkt.QuenchID, used, err = XdrGetInt64(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

func (pkt *QnchRply) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutInt32(buffer, int32(pkt.XID))
	XdrPutInt64(buffer, pkt.QuenchID)
}

type SubAddNotify struct {
	SecureQuenchIDs   []int64
	InsecureQuenchIDs []int64
	TermID            uint64
	SubExpr           SubAST
}

func (pkt *SubAddNotify) ID() int {
	return PacketSubAddNotify
}

func (pkt *SubAddNotify) IDString() string {
	return "SubAddNotify"
}

func (pkt *SubAddNotify) IString(indent string) string {
	return fmt.Sprintf(
		"%sSecureQuenchIDs: %v\n"+
			"%sInsecureQuenchIDs: %v\n"+
			"%sTermID: %d\n"+
			"%sSubExpr: %v\n",
		indent, pkt.SecureQuenchIDs,
		indent, pkt.InsecureQuenchIDs,
		indent, pkt.TermID,
		indent, pkt.SubExpr)
}

func (pkt *SubAddNotify) String() string {
	return pkt.IString("")
}

func (pkt *SubAddNotify) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	secureQidsCount, used, err := XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	for i := uint32(0); i < secureQidsCount; i++ {
		pkt.SecureQuenchIDs[i], used, err = XdrGetInt64(bytes[offset:])
		if err != nil {
			return err
		}
		offset += used
	}

	insecureQidsCount, used, err := XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	for i := uint32(0); i < insecureQidsCount; i++ {
		pkt.InsecureQuenchIDs[i], used, err = XdrGetInt64(bytes[offset:])
		if err != nil {
			return err
		}
		offset += used
	}

	pkt.TermID, used, err = XdrGetUint64(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	// FIXME
	// pkt.SubExpr

	return nil
}

func (pkt *SubAddNotify) Encode(buffer *bytes.Buffer) {
	XdrPutUint32(buffer, uint32(len(pkt.SecureQuenchIDs)))
	for i := 0; i < len(pkt.SecureQuenchIDs); i++ {
		XdrPutInt64(buffer, pkt.SecureQuenchIDs[i])
	}

	XdrPutUint32(buffer, uint32(len(pkt.InsecureQuenchIDs)))
	for i := 0; i < len(pkt.InsecureQuenchIDs); i++ {
		XdrPutInt64(buffer, pkt.InsecureQuenchIDs[i])
	}

	XdrPutUint64(buffer, pkt.TermID)

	// FIXME
	// SubExpr
}

type SubModNotify struct {
	SecureQuenchIDs   []int64
	InsecureQuenchIDs []int64
	TermID            uint64
	SubExpr           SubAST
}

func (pkt *SubModNotify) ID() int {
	return PacketSubModNotify
}

func (pkt *SubModNotify) IDString() string {
	return "SubModNotify"
}

func (pkt *SubModNotify) IString(indent string) string {
	return fmt.Sprintf(
		"%sSecureQuenchIDs: %v\n"+
			"%sInsecureQuenchIDs: %v\n"+
			"%sTermID: %d\n"+
			"%sSubExpr: %v\n",
		indent, pkt.SecureQuenchIDs,
		indent, pkt.InsecureQuenchIDs,
		indent, pkt.TermID,
		indent, pkt.SubExpr)
}

func (pkt *SubModNotify) String() string {
	return pkt.IString("")
}

func (pkt *SubModNotify) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	secureQidsCount, used, err := XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	for i := uint32(0); i < secureQidsCount; i++ {
		pkt.SecureQuenchIDs[i], used, err = XdrGetInt64(bytes[offset:])
		if err != nil {
			return err
		}
		offset += used
	}

	insecureQidsCount, used, err := XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	for i := uint32(0); i < insecureQidsCount; i++ {
		pkt.InsecureQuenchIDs[i], used, err = XdrGetInt64(bytes[offset:])
		if err != nil {
			return err
		}
		offset += used
	}

	pkt.TermID, used, err = XdrGetUint64(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	// FIXME
	// pkt.SubExpr

	return nil
}

func (pkt *SubModNotify) Encode(buffer *bytes.Buffer) {
	XdrPutUint32(buffer, uint32(len(pkt.SecureQuenchIDs)))
	for i := 0; i < len(pkt.SecureQuenchIDs); i++ {
		XdrPutInt64(buffer, pkt.SecureQuenchIDs[i])
	}

	XdrPutUint32(buffer, uint32(len(pkt.InsecureQuenchIDs)))
	for i := 0; i < len(pkt.InsecureQuenchIDs); i++ {
		XdrPutInt64(buffer, pkt.InsecureQuenchIDs[i])
	}

	XdrPutUint64(buffer, pkt.TermID)

	// FIXME
	// SubExpr
}

type SubDelNotify struct {
	QuenchIDs []int64
	TermID    uint64
}

func (pkt *SubDelNotify) ID() int {
	return PacketSubDelNotify
}

func (pkt *SubDelNotify) IDString() string {
	return "SubDelNotify"
}

func (pkt *SubDelNotify) IString(indent string) string {
	return fmt.Sprintf(
		"%sQuenchIDs: %v\n"+
			"%sTermID: %v\n",
		indent, pkt.QuenchIDs,
		indent, pkt.TermID)
}

func (pkt *SubDelNotify) String() string {
	return pkt.IString("")
}

func (pkt *SubDelNotify) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	qidCount, used, err := XdrGetUint32(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	for i := uint32(0); i < qidCount; i++ {
		pkt.QuenchIDs[i], used, err = XdrGetInt64(bytes[offset:])
		if err != nil {
			return err
		}
		offset += used
	}

	pkt.TermID, used, err = XdrGetUint64(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

func (pkt *SubDelNotify) Encode(buffer *bytes.Buffer) {
	XdrPutUint32(buffer, uint32(len(pkt.QuenchIDs)))
	for i := 0; i < len(pkt.QuenchIDs); i++ {
		XdrPutInt64(buffer, pkt.QuenchIDs[i])
	}

	XdrPutUint64(buffer, pkt.TermID)
}
