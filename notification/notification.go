package notification

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
)

type Notification struct {
	RequestType uint8
	Identifier  uint32
	Expiry      uint32
	TokenLength uint16
	Token       [32]byte
}

type NotificationAndPayload struct {
	Notification Notification
	Payload      string
}

type NotificationFailure struct {
	FailureType uint8
	Status      uint8
	Identifier  uint32
}

func DeviceTokenAsBinary(token string) ([32]byte, error) {
	decoded, err := hex.DecodeString(token)
	b := [32]byte{}
	copy(b[:], decoded)
	return b, err
}

func MakeNotification(identifier int, token string, payload string) NotificationAndPayload {
	binaryToken, _ := DeviceTokenAsBinary(token)
	return NotificationAndPayload{Notification{1, uint32(identifier), 0, 32, binaryToken}, payload}
}

func NotificationToBytes(n Notification, payload []byte) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, &n); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint16(len(payload))); err != nil {
		return nil, err
	}
	buf.Write(payload)
	return buf, nil
}

func NotificationFailureFromBytes(resp *bytes.Buffer) NotificationFailure {
	var f NotificationFailure
	binary.Read(resp, binary.BigEndian, &f)
	return f
}

func ApsPayload(payload string) ([]byte, error) {
	type tree map[string]interface{}
	jsonPayload := tree{"aps": tree{"alert": payload}}
	return json.Marshal(jsonPayload)
}

func ResetAfter(identifier uint32, queue []NotificationAndPayload) []NotificationAndPayload {
	for index, n := range queue {
		if n.Notification.Identifier > identifier {
			return queue[index:]
		}
	}
	return []NotificationAndPayload{}
}
