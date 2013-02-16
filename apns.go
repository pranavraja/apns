package apns

import (
    "net"
    "time"
    "crypto/tls"
    "bytes"
    "encoding/binary"
    "encoding/hex"
    "encoding/json"
)

type Notification struct {
    RequestType uint8
    Identifier uint32
    Expiry uint32
    TokenLength uint16
    Token [32]byte
}

type NotificationAndPayload struct {
    Notification Notification
    Payload string
}

type NotificationFailure struct {
    FailureType uint8
    Status uint8
    Identifier uint32
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
    if err := binary.Write(buf, binary.BigEndian, uint16(len(payload)));
        err != nil {
        return nil, err
    }
    buf.Write(payload)
    return buf, nil
}

func NotificationFailureFromBytes(resp *bytes.Buffer) (NotificationFailure) {
    var f NotificationFailure
    binary.Read(resp, binary.BigEndian, &f)
    return f
}

func ApsPayload(payload string) ([]byte, error) {
    type tree map[string] interface{}
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

func Connect(host string, certFile string, keyFile string) (conn *tls.Conn, err error) {
    conf := new(tls.Config)
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return
    }
    conf.Certificates = append(conf.Certificates, cert)
    return tls.Dial("tcp", host, conf)
}

func Channels(conn net.Conn) (writeChannel chan NotificationAndPayload, readChannel chan NotificationFailure, err error) {
    readChannel = make(chan NotificationFailure, 0)
    writeChannel = make(chan NotificationAndPayload, 100)
    go func () {
        for {
            conn.SetReadDeadline(time.Now().Add(2 * time.Second))
            failure := make([]byte, 6)
            _, err := conn.Read(failure)
            if err != nil {
                close(readChannel)
                break
            }
            readChannel <- NotificationFailureFromBytes(bytes.NewBuffer(failure))
        }
    }()
    go func () {
        for {
            notification := <-writeChannel
            apsPayload, _ := ApsPayload(notification.Payload)
            notificationBytes, _ := NotificationToBytes(notification.Notification, apsPayload)
            conn.Write(notificationBytes.Bytes())
        }
    }()
    return
}

func SendNotifications(write chan NotificationAndPayload, read chan NotificationFailure, queue []NotificationAndPayload) {
    for _, n := range queue {
        write <- n
    }
    failure := <-read
    if failure.Identifier != 0 {
        SendNotifications(write, read, ResetAfter(failure.Identifier, queue))
    }
}

func ConnectAndSend(host string, certFile string, keyFile string, queue []NotificationAndPayload) (err error) {
    conn, err := Connect(host, certFile, keyFile)
    if err != nil {
        return
    }
    write, read, err := Channels(conn)
    if err != nil {
        return
    }
    SendNotifications(write, read, queue)
    return
}
