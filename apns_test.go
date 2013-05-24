package apns

import (
	"./notification"
	"net"
	"testing"
	"time"
)

func BenchmarkNotificationSend(b *testing.B) {
	queue := make([]notification.NotificationAndPayload, b.N)
	for i := 0; i < b.N; i++ {
		queue[i] = notification.MakeNotification(i, "04049bc60fc0a90ab23619c6a33e017ab6a9ea17de42b5eb008ed1f51a0eacee", "hi iphone")
	}
	ConnectAndSend("gateway.sandbox.push.apple.com:2195", "dev.pem", "dev.private.pem", queue)
}

func TestQueue(t *testing.T) {
	queue := NewQueue().Add(1, "a", "payload").Add(2, "b", "payload2").Add(3, "b", "payload2")
	if len(queue) != 3 {
		t.Errorf("queue has wrong number of elements: ", queue)
	}
	queue = queue.ResetAfter(2)
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
