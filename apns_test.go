package apns

import (
	"errors"
	"testing"
	"time"
)

func BenchmarkNotificationSend(b *testing.B) {
	queue := NewQueue()
	for i := 0; i < b.N; i++ {
		queue = queue.Add(i, "04049bc60fc0a90ab23619c6a33e017ab6a9ea17de42b5eb008ed1f51a0eacee", "hi iphone")
	}
	service, err := Connect("gateway.sandbox.push.apple.com:2195", "dev.pem", "dev.private.pem")
	if err != nil {
		panic(err)
	}
	_, unsent, err := service.Send(queue, 2*time.Second)
	if err != nil {
		panic(err)
	}
	if len(unsent) != 0 {
		panic("some notifications were not sent due to an error")
	}
}

func TestQueue(t *testing.T) {
	queue := NewQueue().Add(1, "a", "payload").Add(2, "b", "payload2").Add(3, "b", "payload2")
	if len(queue) != 3 {
		t.Errorf("queue has wrong number of elements: ", queue)
	}
	if queue[0].Header.Identifier != 1 {
		t.Errorf("first identifier != 1")
	}
}

type StubConnection struct {
	Buffer             []byte
	Written            []byte
	shouldErrorOnRead  bool
	shouldErrorOnWrite bool
}

func (conn *StubConnection) Read(b []byte) (int, error) {
	if conn.shouldErrorOnRead {
		return 0, errors.New("read error")
	}
	copy(b, conn.Buffer)
	return len(conn.Buffer), nil
}
func (conn *StubConnection) Close() error {
	return nil
}
func (conn *StubConnection) SetReadDeadline(t time.Time) error {
	return nil
}
func (conn *StubConnection) Write(b []byte) (int, error) {
	if conn.shouldErrorOnWrite {
		return 0, errors.New("write error")
	}
	copy(conn.Written, b)
	return len(b), nil
}
