// Copyright 2018 Cobaro Pty Ltd. All Rights Reserved.

// This file is part of elvind
//
// elvind is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// elvind is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with elvind. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

const (
	PacketReserved          = 0
	PacketSvrRqst           = 16
	PacketSvrAdvt           = 17
	PacketSvrAdvtClose      = 18
	PacketUnotify           = 32
	PacketNack              = 48
	PacketConnRqst          = 49
	PacketConnRply          = 50
	PacketDisconnRqst       = 51
	PacketDisconnRply       = 52
	PacketDisconn           = 53
	PacketSecRqst           = 54
	PacketSecRply           = 55
	PacketNotifyEmit        = 56
	PacketNotifyDeliver     = 57
	PacketSubAddRqst        = 58
	PacketSubModRqst        = 59
	PacketSubDelRqst        = 60
	PacketSubRply           = 61
	PacketDropWarn          = 62
	PacketTestConn          = 63
	PacketConfConn          = 64
	PacketAck               = 65
	PacketStatusUpdate      = 66
	PacketAuthRqst          = 67
	PacketAuthCont          = 68
	PacketAuthAck           = 69
	PacketQosRqst           = 70
	PacketQosRply           = 71
	PacketQnchAddRqst       = 80
	PacketQnchModRqst       = 81
	PacketQnchDelRqst       = 82
	PacketQnchRply          = 83
	PacketSubAddNotify      = 84
	PacketSubModNotify      = 85
	PacketSubDelNotify      = 86
	PacketActivate          = 128
	PacketStandby           = 129
	PacketRestart           = 130
	PacketShutdown          = 131
	PacketServerReport      = 132
	PacketServerNack        = 133
	PacketServerStatsReport = 134
	PacketClstJoinRqst      = 160
	PacketClstJoinRply      = 161
	PacketClstTerms         = 162
	PacketClstNotify        = 163
	PacketClstRedir         = 164
	PacketClstLeave         = 165
	PacketFedConnRqst       = 192
	PacketFedConnRply       = 193
	PacketFedSubReplace     = 194
	PacketFedNotify         = 195
	PacketFedSubDiff        = 196
	PacketFailoverConnRqst  = 224
	PacketFailoverConnRply  = 225
	PacketFailoverMaster    = 226
)

// In a protocol packet the type is encoded
func PacketId(bytes []byte) int {
	return int(binary.BigEndian.Uint32(bytes[0:4]))
}

// FIXME: This goes away once each packet is coded and has an IdString()
func PacketIdString(packetId int) string {
	switch packetId {
	case PacketReserved:
		return "Reserved"
	case PacketSvrRqst:
		return "SvrRqst"
	case PacketSvrAdvt:
		return "SvrAdvt"
	case PacketSvrAdvtClose:
		return "SvrAdvtClose"
	case PacketUnotify:
		return "Unotify"
	case PacketNack:
		return "Nack"
	case PacketConnRqst:
		return "ConnRqst"
	case PacketConnRply:
		return "ConnRply"
	case PacketDisconnRqst:
		return "DisconnRqst"
	case PacketDisconnRply:
		return "DisconnRply"
	case PacketDisconn:
		return "Disconn"
	case PacketSecRqst:
		return "SecRqst"
	case PacketSecRply:
		return "SecRply"
	case PacketNotifyEmit:
		return "NotifyEmit"
	case PacketNotifyDeliver:
		return "NotifyDeliver"
	case PacketSubAddRqst:
		return "SubAddRqst"
	case PacketSubModRqst:
		return "SubModRqst"
	case PacketSubDelRqst:
		return "SubDelRqst"
	case PacketSubRply:
		return "SubRply"
	case PacketDropWarn:
		return "DropWarn"
	case PacketTestConn:
		return "TestConn"
	case PacketConfConn:
		return "ConfConn"
	case PacketAck:
		return "Ack"
	case PacketStatusUpdate:
		return "StatusUpdate"
	case PacketAuthRqst:
		return "AuthRqst"
	case PacketAuthCont:
		return "AuthCont"
	case PacketAuthAck:
		return "AuthAck"
	case PacketQosRqst:
		return "QosRqst"
	case PacketQosRply:
		return "QosRply"
	case PacketQnchAddRqst:
		return "QnchAddRqst"
	case PacketQnchModRqst:
		return "QnchModRqst"
	case PacketQnchDelRqst:
		return "QnchDelRqst"
	case PacketQnchRply:
		return "QnchRply"
	case PacketSubAddNotify:
		return "SubAddNotify"
	case PacketSubModNotify:
		return "SubModNotify"
	case PacketSubDelNotify:
		return "SubDelNotify"
	case PacketActivate:
		return "Activate"
	case PacketStandby:
		return "Standby"
	case PacketRestart:
		return "Restart"
	case PacketShutdown:
		return "Shutdown"
	case PacketServerReport:
		return "ServerReport"
	case PacketServerNack:
		return "ServerNack"
	case PacketServerStatsReport:
		return "ServerStatsReport"
	case PacketClstJoinRqst:
		return "ClstJoinRqst"
	case PacketClstJoinRply:
		return "ClstJoinRply"
	case PacketClstTerms:
		return "ClstTerms"
	case PacketClstNotify:
		return "ClstNotify"
	case PacketClstRedir:
		return "ClstRedir"
	case PacketClstLeave:
		return "ClstLeave"
	case PacketFedConnRqst:
		return "FedConnRqst"
	case PacketFedConnRply:
		return "FedConnRply"
	case PacketFedSubReplace:
		return "FedSubReplace"
	case PacketFedNotify:
		return "FedNotify"
	case PacketFedSubDiff:
		return "FedSubDiff"
	case PacketFailoverConnRqst:
		return "FailoverConnRqst"
	case PacketFailoverConnRply:
		return "FailoverConnRply"
	case PacketFailoverMaster:
		return "FailoverMaster"
	default:
		return "Unknown"
	}
}

// All packets must implement these
type Packet interface {
	Id() int
	IdString() string
	String() string
	Decode(bytes []byte) (err error)
	Encode() (bytes []byte, err error)
}

// Notification element types
const (
	NotificationReserved = iota
	NotificationInt32    // marshalled as xdr_long
	NotificationInt64    // marshalled as xdr_hyper
	NotificationFloat64  // marshalled as xdr_double
	NotificationString   // marshalled as xdr_string
	NotificationOpaque   // marshalled as xdr_opaque
)

type Key struct {
	FIXME int
}

//
// FIXME: These will get split out soonish
//

// Get an xdr marshalled 32 bit signed int
func XdrGetInt32(bytes []byte) (i int, used int) {
	return int(binary.BigEndian.Uint32(bytes)), 4
}

// Get an xdr marshalled 32 bit unsigned int
func XdrGetUint32(bytes []byte) (u uint32, used int) {
	return binary.BigEndian.Uint32(bytes), 4
}

// Get an xdr marshalled 64 bit signed int
func XdrGetInt64(bytes []byte) (i int64, used int) {
	return int64(binary.BigEndian.Uint64(bytes)), 8
}

// Get an xdr marshalled 64 bit unsigned int
func XdrGetUint64(bytes []byte) (u uint64, used int) {
	return binary.BigEndian.Uint64(bytes), 8
}

// Get an xdr marshalled string
func XdrGetString(bytes []byte) (s string, used int) {
	// string length
	length, used := XdrGetInt32(bytes)
	// name
	return string(bytes[used : used+length]), used + length + (3 - (length+3)%4) // strings use 4 byte boundaries
}

// Get an xdr marshalled list of opaque bytes
func XdrGetOpaque(bytes []byte) (b []byte, used int) {
	// string length
	length, used := XdrGetInt32(bytes)
	// name
	return bytes[used : used+length], used + length + (3 - (length+3)%4) // opaques use 4 byte boundaries
}

// Get an xdr marshalled 64 bit floating point
func XdrGetFloat64(bytes []byte) (f float64, used int) {
	return math.Float64frombits(uint64(bytes[0]) | uint64(bytes[1])<<8 |
		uint64(bytes[2])<<16 | uint64(bytes[3])<<24 |
		uint64(bytes[4])<<32 | uint64(bytes[5])<<40 |
		uint64(bytes[6])<<48 | uint64(bytes[7])<<56), 8
}

// Get an xdr marshalled Elvin Notification
func XdrGetNotification(bytes []byte) (nfn map[string]interface{}, used int, err error) {
	nfn = make(map[string]interface{})
	offset := 0

	// Number of elements
	elementCount, used := XdrGetUint32(bytes[offset:])
	offset += used

	for elementCount > 0 {
		name, used := XdrGetString(bytes[offset:])
		offset += used

		// Type of value
		elementType, used := XdrGetInt32(bytes[offset:])
		offset += used

		// values
		switch elementType {
		case NotificationInt32:
			nfn[name], used = XdrGetInt32(bytes[offset:])
			offset += used
			break

		case NotificationInt64:
			nfn[name], used = XdrGetUint64(bytes[offset:])
			offset += used
			break

		case NotificationFloat64:
			nfn[name], used = XdrGetFloat64(bytes[offset:])
			offset += used
			break

		case NotificationString:
			var used int
			nfn[name], used = XdrGetString(bytes[offset:])
			offset += used
			break

		case NotificationOpaque:
			var used int
			nfn[name], used = XdrGetOpaque(bytes[offset:])
			offset += used
			break

		default:
			return nil, 0, errors.New("Marshalling failed: unknown element type")
		}

		elementCount--
	}

	return nfn, offset, err
}

// Fixme: implement
// Get an xdr marshalled keyset
func XdrGetKeys(bytes []byte) (keys [][]byte, used int, err error) {
	return nil, 0, nil

	offset := 0

	// Number of elements
	fmt.Println("BYTES", bytes)
	elementCount, used := XdrGetUint32(bytes[offset:])
	offset += used
	fmt.Println("elementCount =", elementCount)

	keys = make([][]byte, elementCount)
	fmt.Println("now", keys)

	for elementCount > 0 {
		fmt.Println("EC", elementCount)
		keys[elementCount-1], used = XdrGetOpaque(bytes[offset:])
		offset += used
		elementCount--
	}
	fmt.Println("then", keys)

	return keys, offset, nil
}
