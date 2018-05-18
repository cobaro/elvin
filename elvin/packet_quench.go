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
	// FIXME
}

type QnchAddRqst struct {
	Xid             uint32
	Names           []string
	DeliverInsecure bool
	Keys            []Keyset
}

func (pkt *QnchAddRqst) Id() int {
	return PacketQnchAddRqst
}

func (pkt *QnchAddRqst) IdString() string {
	return "QnchAddRqst"
}

func (pkt *QnchAddRqst) IString(indent string) string {
	return fmt.Sprintf(
		"%sXid: %d\n"+
			"%sNames: %v\n"+
			"%sDeliverInsecure: %v\n"+
			"%sKeys: %v\n",
		indent, pkt.Xid,
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

	pkt.Xid, used = XdrGetUint32(bytes[offset:])
	offset += used

	nameCount, used := XdrGetUint32(bytes[offset:])
	offset += used
	for i := uint32(0); i < nameCount; i++ {
		pkt.Names[i], used = XdrGetString(bytes[offset:])
		offset += used
	}

	pkt.DeliverInsecure, used = XdrGetBool(bytes[offset:])
	offset += used

	pkt.Keys, used, err = XdrGetKeys(bytes[offset:])
	if err != nil {
		return err
	}
	offset += used

	return nil
}

func (pkt *QnchAddRqst) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, pkt.Id())
	XdrPutUint32(buffer, uint32(len(pkt.Names)))
	for i := 0; i < len(pkt.Names); i++ {
		XdrPutString(buffer, pkt.Names[i])
	}
	XdrPutBool(buffer, pkt.DeliverInsecure)
	XdrPutKeys(buffer, pkt.Keys)
}

type QnchModRqst struct {
	Xid             uint32
	QuenchId        uint64
	NamesAdd        []string
	NamesDel        []string
	DeliverInsecure bool
	AddKeys         []Keyset
	DelKeys         []Keyset
}

func (pkt *QnchModRqst) Id() int {
	return PacketQnchAddRqst
}

func (pkt *QnchModRqst) IdString() string {
	return "QnchModRqst"
}

func (pkt *QnchModRqst) IString(indent string) string {
	return fmt.Sprintf(
		"%sXid: %d\n"+
			"%sQuenchId: %d\n"+
			"%sNamesAdd: %v\n"+
			"%sNamesDel: %v\n"+
			"%sDeliverInsecure: %v\n"+
			"%sAddKeys: %v\n"+
			"%sDelKeys: %v\n",
		indent, pkt.Xid,
		indent, pkt.QuenchId,
		indent, pkt.NamesAdd,
		indent, pkt.NamesDel,
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

	pkt.Xid, used = XdrGetUint32(bytes[offset:])
	offset += used

	pkt.QuenchId, used = XdrGetUint64(bytes[offset:])
	offset += used

	namesAddCount, used := XdrGetUint32(bytes[offset:])
	offset += used
	for i := uint32(0); i < namesAddCount; i++ {
		pkt.NamesAdd[i], used = XdrGetString(bytes[offset:])
		offset += used
	}

	namesDelCount, used := XdrGetUint32(bytes[offset:])
	offset += used
	for i := uint32(0); i < namesDelCount; i++ {
		pkt.NamesDel[i], used = XdrGetString(bytes[offset:])
		offset += used
	}

	pkt.DeliverInsecure, used = XdrGetBool(bytes[offset:])
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

func (pkt *QnchModRqst) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, pkt.Id())
	XdrPutUint64(buffer, pkt.QuenchId)
	XdrPutUint32(buffer, uint32(len(pkt.NamesAdd)))
	for i := 0; i < len(pkt.NamesAdd); i++ {
		XdrPutString(buffer, pkt.NamesAdd[i])
	}
	XdrPutUint32(buffer, uint32(len(pkt.NamesDel)))
	for i := 0; i < len(pkt.NamesDel); i++ {
		XdrPutString(buffer, pkt.NamesDel[i])
	}
	XdrPutBool(buffer, pkt.DeliverInsecure)
	XdrPutKeys(buffer, pkt.AddKeys)
	XdrPutKeys(buffer, pkt.DelKeys)
}

type QnchDelRqst struct {
	Xid      uint32
	QuenchId uint64
}

func (pkt *QnchDelRqst) Id() int {
	return PacketQnchDelRqst
}

func (pkt *QnchDelRqst) IdString() string {
	return "QnchDelRqst"
}

func (pkt *QnchDelRqst) IString(indent string) string {
	return fmt.Sprintf(
		"%sXid: %d\n"+
			"%sQuenchId: %d\n",
		indent, pkt.Xid,
		indent, pkt.QuenchId)
}

func (pkt *QnchDelRqst) String() string {
	return pkt.IString("")
}

func (pkt *QnchDelRqst) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.Xid, used = XdrGetUint32(bytes[offset:])
	offset += used

	pkt.QuenchId, used = XdrGetUint64(bytes[offset:])
	offset += used

	return nil
}

func (pkt *QnchDelRqst) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, pkt.Id())
	XdrPutUint64(buffer, pkt.QuenchId)
}

type QnchRply struct {
	Xid      uint32
	QuenchId uint64
}

func (pkt *QnchRply) Id() int {
	return PacketQnchRply
}

func (pkt *QnchRply) IdString() string {
	return "QnchRply"
}

func (pkt *QnchRply) IString(indent string) string {
	return fmt.Sprintf(
		"%sXid: %d\n"+
			"%sQuenchId: %d\n",
		indent, pkt.Xid,
		indent, pkt.QuenchId)
}

func (pkt *QnchRply) String() string {
	return pkt.IString("")
}

func (pkt *QnchRply) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	pkt.Xid, used = XdrGetUint32(bytes[offset:])
	offset += used

	pkt.QuenchId, used = XdrGetUint64(bytes[offset:])
	offset += used

	return nil
}

func (pkt *QnchRply) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, pkt.Id())
	XdrPutUint64(buffer, pkt.QuenchId)
}

type SubAddNotify struct {
	SecureQuenchIds   []uint64
	InsecureQuenchIds []uint64
	TermId            uint64
	SubExpr           SubAST
}

func (pkt *SubAddNotify) Id() int {
	return PacketSubAddNotify
}

func (pkt *SubAddNotify) IdString() string {
	return "SubAddNotify"
}

func (pkt *SubAddNotify) IString(indent string) string {
	return fmt.Sprintf(
		"%sSecureQuenchIds: %v\n"+
			"%sInsecureQuenchIds: %v\n"+
			"%sTermId: %d\n"+
			"%sSubExpr: %v\n",
		indent, pkt.SecureQuenchIds,
		indent, pkt.InsecureQuenchIds,
		indent, pkt.TermId,
		indent, pkt.SubExpr)
}

func (pkt *SubAddNotify) String() string {
	return pkt.IString("")
}

func (pkt *SubAddNotify) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	secureQidsCount, used := XdrGetUint32(bytes[offset:])
	offset += used
	for i := uint32(0); i < secureQidsCount; i++ {
		pkt.SecureQuenchIds[i], used = XdrGetUint64(bytes[offset:])
		offset += used
	}

	insecureQidsCount, used := XdrGetUint32(bytes[offset:])
	offset += used
	for i := uint32(0); i < insecureQidsCount; i++ {
		pkt.InsecureQuenchIds[i], used = XdrGetUint64(bytes[offset:])
		offset += used
	}

	pkt.TermId, used = XdrGetUint64(bytes[offset:])
	offset += used

	// FIXME
	// pkt.SubExpr

	return nil
}

func (pkt *SubAddNotify) Encode(buffer *bytes.Buffer) {
	XdrPutUint32(buffer, uint32(len(pkt.SecureQuenchIds)))
	for i := 0; i < len(pkt.SecureQuenchIds); i++ {
		XdrPutUint64(buffer, pkt.SecureQuenchIds[i])
	}

	XdrPutUint32(buffer, uint32(len(pkt.InsecureQuenchIds)))
	for i := 0; i < len(pkt.InsecureQuenchIds); i++ {
		XdrPutUint64(buffer, pkt.InsecureQuenchIds[i])
	}

	XdrPutUint64(buffer, pkt.TermId)

	// FIXME
	// SubExpr
}

type SubModNotify struct {
	SecureQuenchIds   []uint64
	InsecureQuenchIds []uint64
	TermId            uint64
	SubExpr           SubAST
}

func (pkt *SubModNotify) Id() int {
	return PacketSubModNotify
}

func (pkt *SubModNotify) IdString() string {
	return "SubModNotify"
}

func (pkt *SubModNotify) IString(indent string) string {
	return fmt.Sprintf(
		"%sSecureQuenchIds: %v\n"+
			"%sInsecureQuenchIds: %v\n"+
			"%sTermId: %d\n"+
			"%sSubExpr: %v\n",
		indent, pkt.SecureQuenchIds,
		indent, pkt.InsecureQuenchIds,
		indent, pkt.TermId,
		indent, pkt.SubExpr)
}

func (pkt *SubModNotify) String() string {
	return pkt.IString("")
}

func (pkt *SubModNotify) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	secureQidsCount, used := XdrGetUint32(bytes[offset:])
	offset += used
	for i := uint32(0); i < secureQidsCount; i++ {
		pkt.SecureQuenchIds[i], used = XdrGetUint64(bytes[offset:])
		offset += used
	}

	insecureQidsCount, used := XdrGetUint32(bytes[offset:])
	offset += used
	for i := uint32(0); i < insecureQidsCount; i++ {
		pkt.InsecureQuenchIds[i], used = XdrGetUint64(bytes[offset:])
		offset += used
	}

	pkt.TermId, used = XdrGetUint64(bytes[offset:])
	offset += used

	// FIXME
	// pkt.SubExpr

	return nil
}

func (pkt *SubModNotify) Encode(buffer *bytes.Buffer) {
	XdrPutUint32(buffer, uint32(len(pkt.SecureQuenchIds)))
	for i := 0; i < len(pkt.SecureQuenchIds); i++ {
		XdrPutUint64(buffer, pkt.SecureQuenchIds[i])
	}

	XdrPutUint32(buffer, uint32(len(pkt.InsecureQuenchIds)))
	for i := 0; i < len(pkt.InsecureQuenchIds); i++ {
		XdrPutUint64(buffer, pkt.InsecureQuenchIds[i])
	}

	XdrPutUint64(buffer, pkt.TermId)

	// FIXME
	// SubExpr
}

type SubDelNotify struct {
	QuenchIds []uint64
	TermId    uint64
}

func (pkt *SubDelNotify) Id() int {
	return PacketSubDelNotify
}

func (pkt *SubDelNotify) IdString() string {
	return "SubDelNotify"
}

func (pkt *SubDelNotify) IString(indent string) string {
	return fmt.Sprintf(
		"%sQuenchIds: %v\n"+
			"%sTermId: %v\n",
		indent, pkt.QuenchIds,
		indent, pkt.TermId)
}

func (pkt *SubDelNotify) String() string {
	return pkt.IString("")
}

func (pkt *SubDelNotify) Decode(bytes []byte) (err error) {
	var used int
	offset := 4 // header

	qidCount, used := XdrGetUint32(bytes[offset:])
	offset += used
	for i := uint32(0); i < qidCount; i++ {
		pkt.QuenchIds[i], used = XdrGetUint64(bytes[offset:])
		offset += used
	}

	pkt.TermId, used = XdrGetUint64(bytes[offset:])
	offset += used

	return nil
}

func (pkt *SubDelNotify) Encode(buffer *bytes.Buffer) {
	XdrPutUint32(buffer, uint32(len(pkt.QuenchIds)))
	for i := 0; i < len(pkt.QuenchIds); i++ {
		XdrPutUint64(buffer, pkt.QuenchIds[i])
	}

	XdrPutUint64(buffer, pkt.TermId)
}
