package apns

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func BenchmarkNotificationSend(b *testing.B) {
	queue := make([]NotificationAndPayload, b.N)
	for i := 0; i < b.N; i++ {
		queue[i] = MakeNotification(i, "04049bc60fc0a90ab23619c6a33e017ab6a9ea17de42b5eb008ed1f51a0eacee", "hi iphone")
	}
	ConnectAndSend("gateway.sandbox.push.apple.com:2195", "dev.pem", "dev.private.pem", queue)
}

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

func TestResetAfter(t *testing.T) {
	queue := []NotificationAndPayload{MakeNotification(1, "a", "payload"),
		MakeNotification(2, "b", "payload2"),
		MakeNotification(3, "b", "payload2")}
	queue = ResetAfter(2, queue)
	if len(queue) != 1 {
		t.Errorf("queue has too many elements left: ", queue)
	}
	if queue[0].Notification.Identifier != 3 {
		t.Errorf("first identifier != 3")
	}
}

// Implements net.Conn that returns a canned value when read.
// Everything else on this Conn is a no-op
type StubConnection struct {
	Buffer []byte
}

func (conn StubConnection) Read(b []byte) (int, error) {
	copy(b, conn.Buffer)
	return len(conn.Buffer), nil
}
func (conn StubConnection) Close() error {
	return nil
}
func (conn StubConnection) RemoteAddr() net.Addr {
	return nil
}
func (conn StubConnection) LocalAddr() net.Addr {
	return nil
}
func (conn StubConnection) SetReadDeadline(t time.Time) error {
	return nil
}
func (conn StubConnection) SetWriteDeadline(t time.Time) error {
	return nil
}
func (conn StubConnection) SetDeadline(t time.Time) error {
	return nil
}
func (conn StubConnection) Write(b []byte) (int, error) {
	return 0, nil
}

func TestReadsFailures(t *testing.T) {
	_, read, _ := Channels(StubConnection{[]byte{8, 4, 0, 0, 0, 2}})
	failure := <-read
	if failure.Status != 4 {
		t.Errorf("couldn't read failure status")
	}
	if failure.Identifier != 2 {
		t.Errorf("couldn't read failure identifier")
	}
}
