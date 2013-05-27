package notification

import (
	"bytes"
	"testing"
)

func TestSerializeNotification(t *testing.T) {
	var notification Notification
	notification = Notification{header{1, 1, 0, 32, [32]byte{'a'}}, "payload"}
	b, _ := notification.Bytes()
	expected := []byte{1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0x20, 'a', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 'p', 'a', 'y', 'l', 'o', 'a', 'd'}
	if !bytes.Equal(b, expected) {
		t.Errorf("\nActual:    %x\n Expected: %x", b, expected)
	}
}

func TestMakeNotification(t *testing.T) {
	n := MakeNotification(1, "ae91fa", "payload")
	expectedToken := []byte{0xAE, 0x91, 0xFA, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	actualToken := n.Header.Token
	if !bytes.Equal(actualToken[:], expectedToken) {
		t.Errorf("Actual: %x, Expected: %x", actualToken, expectedToken)
	}
}
