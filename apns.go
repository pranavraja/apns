package apns

import (
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"time"

	"github.com/pranavraja/apns/notification"
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

func Connect(host string, certFile string, keyFile string) (*ApnsService, error) {
	conf := new(tls.Config)
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	conf.Certificates = append(conf.Certificates, cert)
	service := &ApnsService{host: host, conf: conf}
	err = service.Connect()
	return service, err
}

type deadlineReader interface {
	SetReadDeadline(t time.Time) error
}

type readWriteCloserWithDeadline interface {
	io.ReadWriteCloser
	deadlineReader
}

type ApnsService struct {
	conn readWriteCloserWithDeadline
	host string
	conf *tls.Config
}

// Create a new service with a custom tls.Config.
// Allows you to load in your own certificates and customize verification behavior
func NewService(host string, conf *tls.Config) *ApnsService {
	return &ApnsService{host: host, conf: conf}
}

func (service *ApnsService) Connect() (err error) {
	if service.conn != nil {
		err = service.conn.Close()
		if err != nil {
			return
		}
	}
	service.conn, err = tls.Dial("tcp", service.host, service.conf)
	return
}

// Assuming we are already connected, send a single notification through the current connection.
func (service *ApnsService) SendOne(n notification.Notification) error {
	notificationBytes, _ := n.Bytes()
	_, err := service.conn.Write(notificationBytes)
	return err
}

func (service *ApnsService) ReadInvalid(timeout time.Duration) (f notification.Invalid, err error) {
	invalid := make([]byte, 6) // Protocol defines apns failures to be 6 bytes long. See http://developer.apple.com/library/ios/#documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/CommunicatingWIthAPS.html#//apple_ref/doc/uid/TP40008194-CH101-SW4
	service.conn.SetReadDeadline(time.Now().Add(timeout))
	_, err = service.conn.Read(invalid)
	if readError, ok := err.(net.Error); ok && readError.Timeout() {
		// Timeouts are actually OK, this means that Apple has nothing to say and therefore the send succeeded
		return f, nil
	}
	if err != nil {
		return
	}
	return notification.InvalidFromBytes(bytes.NewBuffer(invalid)), nil
}

// Assuming we are already connected, send notifications in `queue` through the current conn.
// Returns a notification that was invalid and caused Apple to drop the connection, which should not be re-sent,
// a queue of "unsent" notifications, which are notifications that were not sent due to a prior failure or network error and can be re-tried,
// and a network/connection error if one occured.
func (service *ApnsService) Send(queue Queue, timeToWaitForResponse time.Duration) (invalid notification.Invalid, unsent Queue, err error) {
	for i, notificationToSend := range queue {
		err = service.SendOne(notificationToSend)
		if err != nil {
			// If we errored out while sending i, then the rest of the queue (including i) were unsent
			unsent = queue[i:]
			return
		}
	}
	invalid, err = service.ReadInvalid(timeToWaitForResponse)
	if err != nil {
		// If we get here, there's no way to tell whether the notifications were correct or not, as we got an error when reading the response.
		// So in the spirit of optimism, we declare that there are no unsent notifications, and return the error to the caller
		return
	}
	if invalid.Identifier != 0 {
		unsent = queue.ResetAfter(invalid.Identifier)
	}
	return
}

// Assuming we are already connected, send all notifications in `queue`, retrying after apns connection drops until the entire queue is sent.
// Note: If there is an unexpected network error (i.e. if the machine is offline), this will return unsent items to the caller
func (service *ApnsService) SendAll(queue Queue, timeToWaitForEachResponse time.Duration) (invalids []notification.Invalid, unsent Queue, err error) {
	var invalid notification.Invalid
	for len(queue) > 0 {
		invalid, queue, err = service.Send(queue, timeToWaitForEachResponse)
		if err != nil {
			return invalids, queue, err
		}
		if invalid.Identifier != 0 {
			invalids = append(invalids, invalid)
		}
		if len(queue) > 0 {
			err = service.Connect()
			if err != nil {
				return invalids, queue, err
			}
		}
	}
	return
}

// Close closes the underlying socket
func (service *ApnsService) Close() error {
	return service.conn.Close()
}
