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
	"encoding/binary"
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
