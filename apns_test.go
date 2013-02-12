package apns

import (
    "testing"
    "bytes"
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
    expected := []byte{1,0,0,0,1,0,0,0,0,0,0x20,'a',0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,7,'p','a','y','l','o','a','d'}
    if !bytes.Equal(b.Bytes(), expected) {
        t.Errorf("\nActual:    %x\n Expected: %x", b.Bytes(), expected)
    }
}

func TestReadNotificationFailure(t *testing.T) {
    exampleResponse := bytes.NewBuffer([]byte{8,1,0,0,0,1})
    failure := NotificationFailureFromBytes(exampleResponse)
    if failure.Identifier != 1 {
        t.Errorf("Expected identifier=1")
    }
    if failure.Status != 1 {
        t.Errorf("Expected status=1")
    }
}

func TestMakeNotification(t *testing.T) {
    n := MakeNotification(1, "ae91fa", "payload")
    expectedToken := []byte{0xAE,0x91,0xFA,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}
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
