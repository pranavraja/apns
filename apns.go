package apns

import (
	"./notification"
	"bytes"
	"crypto/tls"
	"net"
	"time"
)

type Queue []notification.Notification

func NewQueue() Queue {
	return Queue{}
}

func (queue Queue) Add(identifier int, token string, payload string) Queue {
	return append(queue, notification.MakeNotification(identifier, token, payload))
}

func (queue Queue) ResetAfter(identifier uint32) Queue {
	for index, n := range queue {
		if n.Header.Identifier > identifier {
			return queue[index:]
		}
	}
	return NewQueue()
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

func Channels(conn net.Conn) (writeChannel chan notification.Notification, readChannel chan notification.Failure, err error) {
	readChannel = make(chan notification.Failure, 0)
	writeChannel = make(chan notification.Notification, 100)
	go func() {
		for {
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			failure := make([]byte, 6)
			_, err := conn.Read(failure)
			if err != nil {
				close(readChannel)
				break
			}
			readChannel <- notification.FailureFromBytes(bytes.NewBuffer(failure))
		}
	}()
	go func() {
		for {
			n := <-writeChannel
			notificationBytes, _ := n.Bytes()
			conn.Write(notificationBytes)
		}
	}()
	return
}

func SendNotifications(write chan notification.Notification, read chan notification.Failure, queue Queue) {
	for _, n := range queue {
		write <- n
	}
	failure := <-read
	if failure.Identifier != 0 {
		SendNotifications(write, read, queue.ResetAfter(failure.Identifier))
	}
}

func ConnectAndSend(host string, certFile string, keyFile string, queue Queue) (err error) {
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
