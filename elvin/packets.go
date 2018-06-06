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
	"encoding/binary"
)

const (
	PacketReserved            = 0
	PacketSvrRequest          = 16
	PacketSvrAdvt             = 17
	PacketSvrAdvtClose        = 18
	PacketUnotify             = 32
	PacketNack                = 48
	PacketConnRequest         = 49
	PacketConnReply           = 50
	PacketDisconnRequest      = 51
	PacketDisconnReply        = 52
	PacketDisconn             = 53
	PacketSecRequest          = 54
	PacketSecReply            = 55
	PacketNotifyEmit          = 56
	PacketNotifyDeliver       = 57
	PacketSubAddRequest       = 58
	PacketSubModRequest       = 59
	PacketSubDelRequest       = 60
	PacketSubReply            = 61
	PacketDropWarn            = 62
	PacketTestConn            = 63
	PacketConfConn            = 64
	PacketAck                 = 65
	PacketStatusUpdate        = 66
	PacketAuthRequest         = 67
	PacketAuthCont            = 68
	PacketAuthAck             = 69
	PacketQosRequest          = 70
	PacketQosReply            = 71
	PacketQuenchAddRequest    = 80
	PacketQuenchModRequest    = 81
	PacketQuenchDelRequest    = 82
	PacketQuenchReply         = 83
	PacketSubAddNotify        = 84
	PacketSubModNotify        = 85
	PacketSubDelNotify        = 86
	PacketActivate            = 128
	PacketStandby             = 129
	PacketRestart             = 130
	PacketShutdown            = 131
	PacketServerReport        = 132
	PacketServerNack          = 133
	PacketServerStatsReport   = 134
	PacketClstJoinRequest     = 160
	PacketClstJoinReply       = 161
	PacketClstTerms           = 162
	PacketClstNotify          = 163
	PacketClstRedir           = 164
	PacketClstLeave           = 165
	PacketFedConnRequest      = 192
	PacketFedConnReply        = 193
	PacketFedSubReplace       = 194
	PacketFedNotify           = 195
	PacketFedSubDiff          = 196
	PacketFailoverConnRequest = 224
	PacketFailoverConnReply   = 225
	PacketFailoverMaster      = 226
)

// In a protocol packet the type is encoded
func PacketID(bytes []byte) int {
	return int(binary.BigEndian.Uint32(bytes[0:4]))
}

// Return a usable string from a Packet ID
func PacketIDString(packetID int) string {
	switch packetID {
	case PacketReserved:
		return "Reserved"
	case PacketSvrRequest:
		return "SvrRequest"
	case PacketSvrAdvt:
		return "SvrAdvt"
	case PacketSvrAdvtClose:
		return "SvrAdvtClose"
	case PacketUnotify:
		return "Unotify"
	case PacketNack:
		return "Nack"
	case PacketConnRequest:
		return "ConnRequest"
	case PacketConnReply:
		return "ConnReply"
	case PacketDisconnRequest:
		return "DisconnRequest"
	case PacketDisconnReply:
		return "DisconnReply"
	case PacketDisconn:
		return "Disconn"
	case PacketSecRequest:
		return "SecRequest"
	case PacketSecReply:
		return "SecReply"
	case PacketNotifyEmit:
		return "NotifyEmit"
	case PacketNotifyDeliver:
		return "NotifyDeliver"
	case PacketSubAddRequest:
		return "SubAddRequest"
	case PacketSubModRequest:
		return "SubModRequest"
	case PacketSubDelRequest:
		return "SubDelRequest"
	case PacketSubReply:
		return "SubReply"
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
	case PacketAuthRequest:
		return "AuthRequest"
	case PacketAuthCont:
		return "AuthCont"
	case PacketAuthAck:
		return "AuthAck"
	case PacketQosRequest:
		return "QosRequest"
	case PacketQosReply:
		return "QosReply"
	case PacketQuenchAddRequest:
		return "QuenchAddRequest"
	case PacketQuenchModRequest:
		return "QuenchModRequest"
	case PacketQuenchDelRequest:
		return "QuenchDelRequest"
	case PacketQuenchReply:
		return "QuenchReply"
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
	case PacketClstJoinRequest:
		return "ClstJoinRequest"
	case PacketClstJoinReply:
		return "ClstJoinReply"
	case PacketClstTerms:
		return "ClstTerms"
	case PacketClstNotify:
		return "ClstNotify"
	case PacketClstRedir:
		return "ClstRedir"
	case PacketClstLeave:
		return "ClstLeave"
	case PacketFedConnRequest:
		return "FedConnRequest"
	case PacketFedConnReply:
		return "FedConnReply"
	case PacketFedSubReplace:
		return "FedSubReplace"
	case PacketFedNotify:
		return "FedNotify"
	case PacketFedSubDiff:
		return "FedSubDiff"
	case PacketFailoverConnRequest:
		return "FailoverConnRequest"
	case PacketFailoverConnReply:
		return "FailoverConnReply"
	case PacketFailoverMaster:
		return "FailoverMaster"
	default:
		return "Unknown"
	}
}

// All packets must implement these
type Packet interface {
	ID() int
	IDString() string
	String() string
	Decode(bytes []byte) (err error)
	Encode(buffer *bytes.Buffer)
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
