package notification

import (
	"bytes"
	"testing"
)

func TestSerializeNotification(t *testing.T) {
	var notification Notification
	payload := []byte("payload")
	notification = Notification{1, 1, 0, 32, [32]byte{'a'}}
	b, _ := NotificationToBytes(notification, payload)
	expected := []byte{1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0x20, 'a', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 'p', 'a', 'y', 'l', 'o', 'a', 'd'}
	if !bytes.Equal(b.Bytes(), expected) {
		t.Errorf("\nActual:    %x\n Expected: %x", b.Bytes(), expected)
	}
}

func TestMakeNotification(t *testing.T) {
	n := MakeNotification(1, "ae91fa", "payload")
	expectedToken := []byte{0xAE, 0x91, 0xFA, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	actualToken := n.Notification.Token
	if !bytes.Equal(actualToken[:], expectedToken) {
		t.Errorf("Actual: %x, Expected: %x", actualToken, expectedToken)
	}
}

func TestApsPayload(t *testing.T) {
	payload, _ := ApsPayload("message")
	if !bytes.Equal(payload, []byte("{\"aps\":{\"alert\":\"message\"}}")) {
		t.Errorf("Payload: %s", payload)
	}
}
