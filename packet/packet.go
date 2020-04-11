package packet

import (
	"bytes"
	"strconv"
	"strings"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
	alpha   = []byte{'@'}
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
	ErrorHeader = "0"

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

type PrivatePacket struct {
	UserId string
	RoomId string
	Header int
	Body   []byte
}
type CreateRoomPacket struct {
	Title           string
	MaxParticipants int
}

func GetHeader(content []byte) (int, string) {
	trim := bytes.TrimSpace(bytes.Replace(content, newline, space, -1))
	div := bytes.Split(trim, alpha) //[0] : header(header, roomId), [1] : userid, [2]: body
	//log.Println("ALL HEADER    :", string(div[0]))
	header := bytes.Split(div[0], space)
	if len(header) > 1 { //Room id information exist
		//log.Println("RoomId :", string(header[1]))
		if v, err := strconv.Atoi(string(header[0])); err == nil {
			return v, string(header[1])
		} else {
			return -1, ""
		}
	}
	if v, err := strconv.Atoi(string(header[0])); err == nil {
		return v, ""
	} else {
		return -1, ""
	}
}
func getSeperateContent(allContent []byte) [][]byte {
	trim := bytes.TrimSpace(bytes.Replace(allContent, newline, space, -1))
	div := bytes.Split(trim, alpha)

	return div
}

func BindPrivatePacket(content []byte) *PrivatePacket {
	header, roomId := GetHeader(content)
	alphaDiv := getSeperateContent(content)
	userId := string(alphaDiv[1])
	body := alphaDiv[2]
	if header != -1 {
		return &PrivatePacket{
			UserId: userId,
			Header: header,
			RoomId: roomId,
			Body:   body,
		}
	}

	return nil

}
func BindCreateRoomPacket(content []byte) *CreateRoomPacket {
	alphaDiv := getSeperateContent(content)
	body := alphaDiv[2]
	data := strings.Split(string(body), " ")
	val, _ := strconv.Atoi(data[1])
	return &CreateRoomPacket{
		Title:           data[0],
		MaxParticipants: val,
	}
}

func SockError(code int) []byte {
	body := ErrorHeader + "@@" + strconv.Itoa(code)
	return []byte(body)
}
func SockPacket(header string, body []byte) []byte {
	return append([]byte(header+"@@"), body...)
}
func SockIdentifyPacket(header string, user string, body []byte) []byte {
	return append([]byte(header+"@"+user+"@"), body...)
}
