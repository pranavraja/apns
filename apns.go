package apns

import (
	"./notification"
	"bytes"
	"crypto/tls"
	"net"
	"time"
)

func Connect(host string, certFile string, keyFile string) (conn *tls.Conn, err error) {
	conf := new(tls.Config)
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return
	}
	conf.Certificates = append(conf.Certificates, cert)
	return tls.Dial("tcp", host, conf)
}

func Channels(conn net.Conn) (writeChannel chan notification.NotificationAndPayload, readChannel chan notification.NotificationFailure, err error) {
	readChannel = make(chan notification.NotificationFailure, 0)
	writeChannel = make(chan notification.NotificationAndPayload, 100)
	go func() {
		for {
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			failure := make([]byte, 6)
			_, err := conn.Read(failure)
			if err != nil {
				close(readChannel)
				break
			}
			readChannel <- notification.NotificationFailureFromBytes(bytes.NewBuffer(failure))
		}
	}()
	go func() {
		for {
			n := <-writeChannel
			apsPayload, _ := notification.ApsPayload(n.Payload)
			notificationBytes, _ := notification.NotificationToBytes(n.Notification, apsPayload)
			conn.Write(notificationBytes.Bytes())
		}
	}()
	return
}

func SendNotifications(write chan notification.NotificationAndPayload, read chan notification.NotificationFailure, queue []notification.NotificationAndPayload) {
	for _, n := range queue {
		write <- n
	}
	failure := <-read
	if failure.Identifier != 0 {
		SendNotifications(write, read, notification.ResetAfter(failure.Identifier, queue))
	}
}

func ConnectAndSend(host string, certFile string, keyFile string, queue []notification.NotificationAndPayload) (err error) {
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
