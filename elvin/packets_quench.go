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

// Packet: QuenchAddRequest
type QuenchAddRequest struct {
	XID             uint32
	Names           map[string]bool
	DeliverInsecure bool
	Keys            KeyBlock
}

// Integer value of packet type
func (pkt *QuenchAddRequest) ID() int {
	return PacketQuenchAddRequest
}

// Integer value of packet type
func (pkt *QuenchAddRequest) IDString() string {
	return "QuenchAddRequest"
}

// Pretty print with indent
func (pkt *QuenchAddRequest) IString(indent string) string {
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

// Pretty print without indent so generic ToString() works
func (pkt *QuenchAddRequest) String() string {
	return pkt.IString("")
}

// Decode from a byte array
func (pkt *QuenchAddRequest) Decode(bytes []byte) (err error) {
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

// Encode from a buffer
func (pkt *QuenchAddRequest) Encode(buffer *bytes.Buffer) (xID uint32) {
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

// Packet: QuenchModRequest
type QuenchModRequest struct {
	XID             uint32
	QuenchID        int64
	AddNames        map[string]bool
	DelNames        map[string]bool
	DeliverInsecure bool
	AddKeys         KeyBlock
	DelKeys         KeyBlock
}

// Integer value of packet type
func (pkt *QuenchModRequest) ID() int {
	return PacketQuenchModRequest
}

// Integer value of packet type
func (pkt *QuenchModRequest) IDString() string {
	return "QuenchModRequest"
}

// Pretty print with indent
func (pkt *QuenchModRequest) IString(indent string) string {
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

// Pretty print without indent so generic ToString() works
func (pkt *QuenchModRequest) String() string {
	return pkt.IString("")
}

// Decode from a byte array
func (pkt *QuenchModRequest) Decode(bytes []byte) (err error) {
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

// Encode from a buffer
func (pkt *QuenchModRequest) Encode(buffer *bytes.Buffer) (xID uint32) {
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

// Packet: QuenchDelRequest
type QuenchDelRequest struct {
	XID      uint32
	QuenchID int64
}

// Integer value of packet type
func (pkt *QuenchDelRequest) ID() int {
	return PacketQuenchDelRequest
}

// Integer value of packet type
func (pkt *QuenchDelRequest) IDString() string {
	return "QuenchDelRequest"
}

// Pretty print with indent
func (pkt *QuenchDelRequest) IString(indent string) string {
	return fmt.Sprintf(
		"%sXID: %d\n"+
			"%sQuenchID: %d\n",
		indent, pkt.XID,
		indent, pkt.QuenchID)
}

// Pretty print without indent so generic ToString() works
func (pkt *QuenchDelRequest) String() string {
	return pkt.IString("")
}

// Decode from a byte array
func (pkt *QuenchDelRequest) Decode(bytes []byte) (err error) {
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

// Encode from a buffer
func (pkt *QuenchDelRequest) Encode(buffer *bytes.Buffer) (xID uint32) {
	xID = XID()
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutInt32(buffer, int32(xID))
	XdrPutInt64(buffer, pkt.QuenchID)

	return
}

// Packet: QuenchReply
type QuenchReply struct {
	XID      uint32
	QuenchID int64
}

// Integer value of packet type
func (pkt *QuenchReply) ID() int {
	return PacketQuenchReply
}

// Integer value of packet type
func (pkt *QuenchReply) IDString() string {
	return "QuenchReply"
}

// Pretty print with indent
func (pkt *QuenchReply) IString(indent string) string {
	return fmt.Sprintf(
		"%sXID: %d\n"+
			"%sQuenchID: %d\n",
		indent, pkt.XID,
		indent, pkt.QuenchID)
}

// Pretty print without indent so generic ToString() works
func (pkt *QuenchReply) String() string {
	return pkt.IString("")
}

// Decode from a byte array
func (pkt *QuenchReply) Decode(bytes []byte) (err error) {
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

// Encode from a buffer
func (pkt *QuenchReply) Encode(buffer *bytes.Buffer) {
	XdrPutInt32(buffer, int32(pkt.ID()))
	XdrPutInt32(buffer, int32(pkt.XID))
	XdrPutInt64(buffer, pkt.QuenchID)
}

// Packet: SubAddNotify
type SubAddNotify struct {
	SecureQuenchIDs   []int64
	InsecureQuenchIDs []int64
	TermID            uint64
	SubExpr           SubAST
}

// Integer value of packet type
func (pkt *SubAddNotify) ID() int {
	return PacketSubAddNotify
}

// Integer value of packet type
func (pkt *SubAddNotify) IDString() string {
	return "SubAddNotify"
}

// Pretty print with indent
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

// Pretty print without indent so generic ToString() works
func (pkt *SubAddNotify) String() string {
	return pkt.IString("")
}

// Decode from a byte array
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

// Encode from a buffer
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

// Packet: SubModNotify
type SubModNotify struct {
	SecureQuenchIDs   []int64
	InsecureQuenchIDs []int64
	TermID            uint64
	SubExpr           SubAST
}

// Integer value of packet type
func (pkt *SubModNotify) ID() int {
	return PacketSubModNotify
}

// Integer value of packet type
func (pkt *SubModNotify) IDString() string {
	return "SubModNotify"
}

// Pretty print with indent
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

// Pretty print without indent so generic ToString() works
func (pkt *SubModNotify) String() string {
	return pkt.IString("")
}

// Decode from a byte array
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

// Encode from a buffer
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

// Packet: SubDelNotify
type SubDelNotify struct {
	QuenchIDs []int64
	TermID    uint64
}

// Integer value of packet type
func (pkt *SubDelNotify) ID() int {
	return PacketSubDelNotify
}

// Integer value of packet type
func (pkt *SubDelNotify) IDString() string {
	return "SubDelNotify"
}

// Pretty print with indent
func (pkt *SubDelNotify) IString(indent string) string {
	return fmt.Sprintf(
		"%sQuenchIDs: %v\n"+
			"%sTermID: %v\n",
		indent, pkt.QuenchIDs,
		indent, pkt.TermID)
}

// Pretty print without indent so generic ToString() works
func (pkt *SubDelNotify) String() string {
	return pkt.IString("")
}

// Decode from a byte array
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

// Encode from a buffer
func (pkt *SubDelNotify) Encode(buffer *bytes.Buffer) {
	XdrPutUint32(buffer, uint32(len(pkt.QuenchIDs)))
	for i := 0; i < len(pkt.QuenchIDs); i++ {
		XdrPutInt64(buffer, pkt.QuenchIDs[i])
	}

	XdrPutUint64(buffer, pkt.TermID)
}
