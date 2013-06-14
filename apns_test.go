package apns

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/pranavraja/apns/notification"
	"testing"
	"time"
)

func ExampleNewService() {
	config := new(tls.Config)
	cert, _ := tls.LoadX509KeyPair("cert.pem", "cert.private.pem")
	// Don't verify certificates (we want to man-in-the-middle this)
	// Obviously, don't do this in production!
	config.InsecureSkipVerify = true
	config.Certificates = append(config.Certificates, cert)
	service := NewService("gateway.sandbox.push.apple.com:2195", config)
	service.Connect()
}

func ExampleQueue_Add() {
	queue := NewQueue().Add(1, "aef4429b", `{"aps":{"alert":"message"}}`).Add(2, "aef4429b", `{"aps":{"alert":"message"}}`)
	fmt.Printf("%v", queue)
}

func ExampleQueue_ResetAfter() {
	queue := NewQueue().Add(1, "aef4429b", `{"aps":{"alert":"message"}}`).Add(2, "aef4429b", `{"aps":{"alert":"message"}}`)
	queue = queue.ResetAfter(1)
	fmt.Printf("remaining identifier: %d", queue[0].Header.Identifier)
	// Output:
	// remaining identifier: 2
}

func ExampleApnsService_SendOne() {
	service, _ := Connect("gateway.sandbox.push.apple.com:2195", "dev.pem", "dev.private.pem")
	service.SendOne(notification.MakeNotification(1, "aef4429b", `{"aps":{"alert":"message"}}`))
	failure, _ := service.ReadInvalid(2 * time.Second)
	fmt.Printf("%v", failure)
}

func ExampleApnsService_SendAll() {
	queue := NewQueue().Add(1, "aef4429b", `{"aps":{"alert":"message"}}`).Add(2, "aef4429b", `{"aps":{"alert":"message"}}`)
	failureTimeout := 2 * time.Second
	service, _ := Connect("gateway.sandbox.push.apple.com:2195", "dev.pem", "dev.private.pem")
	failures, unsent, _ := service.SendAll(queue, failureTimeout)
	fmt.Printf("%v", failures)
	fmt.Printf("%v", unsent)
}

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
	Buffer              *bytes.Buffer
	Written             *bytes.Buffer
	shouldErrorOnRead   bool
	shouldTimeoutOnRead bool
	shouldErrorOnWrite  bool
}

type networkError struct {
	IsTimeout bool
}

func (t networkError) Error() string {
	return "Timed out."
}
func (t networkError) Timeout() bool {
	return t.IsTimeout
}
func (t networkError) Temporary() bool {
	return false
}

func (conn StubConnection) Read(b []byte) (int, error) {
	if conn.shouldTimeoutOnRead {
		return 0, networkError{true}
	}
	if conn.shouldErrorOnRead {
		return 0, networkError{false}
	}
	return conn.Buffer.Read(b)
}
func (conn StubConnection) Close() error {
	return nil
}
func (conn StubConnection) SetReadDeadline(t time.Time) error {
	return nil
}
func (conn StubConnection) Write(b []byte) (int, error) {
	if conn.shouldErrorOnWrite {
		return 0, errors.New("write error")
	}
	return conn.Written.Write(b)
}

func TestSend(t *testing.T) {
	queue := NewQueue().Add(1, "a", "payload").Add(2, "b", "payload2").Add(3, "b", "payload2")
	stubConnection := StubConnection{Written: new(bytes.Buffer), Buffer: new(bytes.Buffer)}
	service := ApnsService{conn: stubConnection}
	service.Send(queue, 2*time.Second)
	if l := stubConnection.Written.Len(); l != 158 {
		t.Errorf("not enough bytes written to the connection, should have been 158 but got %d", l)
	}
}

func TestReadInvalid(t *testing.T) {
	stubConnection := StubConnection{Written: new(bytes.Buffer), Buffer: bytes.NewBuffer([]byte{8, 1, 0, 0, 0, 1})}
	service := ApnsService{conn: stubConnection}
	invalid, err := service.ReadInvalid(2 * time.Second)
	if err != nil {
		t.Error(err)
	}
	if invalid.Identifier != 1 {
		t.Errorf("wrong identifier read: %d", invalid.Identifier)
	}
	if invalid.Status != 1 {
		t.Errorf("wrong status read: %d", invalid.Status)
	}
	if invalid.FailureType != 8 {
		t.Errorf("wrong failure type read: %d", invalid.FailureType)
	}
}

func TestReadInvalid_Error(t *testing.T) {
	stubConnection := StubConnection{Written: new(bytes.Buffer), Buffer: bytes.NewBuffer([]byte{8, 1, 0, 0, 0, 1}), shouldErrorOnRead: true}
	service := ApnsService{conn: stubConnection}
	_, err := service.ReadInvalid(2 * time.Second)
	if err == nil {
		t.Errorf("didn't capture error")
	}
}

func TestReadInvalid_Timeout(t *testing.T) {
	stubConnection := StubConnection{Written: new(bytes.Buffer), Buffer: bytes.NewBuffer([]byte{8, 1, 0, 0, 0, 1}), shouldTimeoutOnRead: true}
	service := ApnsService{conn: stubConnection}
	_, err := service.ReadInvalid(2 * time.Second)
	if err != nil {
		t.Errorf("should have swallowed error: %v", err)
	}
}
