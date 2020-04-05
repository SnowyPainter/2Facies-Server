package packet

import (
	"strconv"
)

const ( //Error codes
	ConnectionError   = 101
	UnexceptError     = 102
	RoomJoinError     = 301
	RoomLeaveError    = 302
	RoomFull          = 303
	ExistRoom         = 304
	ChatSend          = 401
	ChatConnect       = 402
	ChatRecv          = 403
	NotFound          = 501
	FormatError       = 001
	IncorrectDataType = 002
)

const ( //Socket Headers
	ErrorHeader = "00"

	JoinHeader   = "11"
	LeaveHeader  = "12"
	CreateHeader = "13"

	BroadcastHeader      = "21"
	BroadcastAudioHeader = "22"
	ParticipantsHeader   = "23"
)

const (
	TypeTextBroadcast  = 1
	TypeAudioBroadcast = 2
)

func SockError(code int) []byte {
	body := ErrorHeader + "@" + strconv.Itoa(code)
	return []byte(body)
}
func SockPacket(header string, body []byte) []byte {
	return append([]byte(header+"@"), body...)
}
