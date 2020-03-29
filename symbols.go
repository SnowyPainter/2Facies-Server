package main

const (
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

const (
	errorHeader = "00"

	joinHeader  = "11"
	leaveHeader = "12"
	createHeaer = "13"

	broadcastHeader      = "21"
	broadcastAudioHeader = "22"
	participantsHeader   = "23"
)

const (
	TypeTextBroadcast  = 1
	TypeAudioBroadcast = 2
)
