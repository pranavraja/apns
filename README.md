# Go push notification client

A work in progress. API will probably change, so if you use this, do the Right Thing and vendor your dependencies.

`go get github.com/pranavraja/apns`

Implements The 'Enhanced Notification Format' (see [Communicating With APS](http://developer.apple.com/library/mac/#documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CommunicatingWIthAPS/CommunicatingWIthAPS.html#//apple_ref/doc/uid/TP40008194-CH101-SW1)), with handling of 'invalid' notification responses.

# Background

[The problem with Apple's push notification service](http://redth.info/the-problem-with-apples-push-notification-ser/).

An 'invalid' notification is one that has an invalid device token or payload, as opposed to one that is valid but couldn't be sent for some reason.

In summary:

- Apple recommends that you re-use a single TCP connection to send all your push notifications.
- However if we happens to send an invalid notification, apple will stay connected, and silently drop ALL the rest of the notifications you send through that connection, valid or invalid, for all time. This continues until you actually read back the failure response from the socket. 
- However, there is no response for valid notifications, so if you want to know if a particular notification is valid, the only way to tell is to allow a socket read to _timeout_. Obviously we can't do this for every single notification we send or it would be horribly slow.
- The normal issues of network errors still apply, for example, if we send all the notifications and there is a network error reading the response, we just have to assume that all the notifications we sent were valid.
- Notifications that failed delivery (because the user disabled push notifications or wasn't online) aren't even reported through the response from this service. You have to read another service called the "feedback service" which will report these at a later point in time. This project doesn't attempt to solve this.

# Usage


```go
package main

import (
    "github.com/pranavraja/apns"
    "os"
    "time"
)

func main() {
    host := os.Getenv("APNS_HOST") // e.g. gateway.push.sandbox.apple.com:2195
    certFile := os.Getenv("CERT_FILE") // e.g. cert.pem
    keyFile := os.Getenv("KEY_FILE") // e.g. cert.private.pem
    service, err := apns.Connect(host, certFile, keyFile)
    if err != nil {
        panic(err) // Couldn't read certificates, or couldn't connect to Apple for some reason
    }
    queue := apns.NewQueue().Add(1, "aef4429b", `{"aps":{"alert":"message"}}`).Add(2, "aef4429b", `{"aps":{"alert":"message"}}`)
    failureTimeout := 2 * time.Second

    // Send a batch of notifications over a connection. 
    // When Apple reports an invalid notification in the queue, skip that one, reconnect to APNS, 
    // and try to re-send the rest, until the queue is exhausted or there is a network error.
    failures, unsent, err := service.SendAll(queue, failureTimeout)

    // Alternatively, when Apple reports an invalid notification in the queue, return immediately. 
    // Report the failure and the items still to be sent. 
    // The caller should handle re-connecting and attempting to re-send the remaining items.
    // This is useful if you want to employ a more complex batching strategy for performance reasons
    // failure, unsent, err := service.Send(queue, failureTimeout)
}
```

Or alternatively, send the notifications one by one, and listen for failures yourself:

```go
package main

import (
    "github.com/pranavraja/apns"
    "github.com/pranavraja/apns/notification"
    "os"
)

func main() {
    host := os.Getenv("APNS_HOST")
    certFile := os.Getenv("CERT_FILE")
    keyFile := os.Getenv("KEY_FILE")
    service, err := apns.Connect(host, certFile, keyFile)
    if err != nil {
        panic(err)
    }
    // Writing a notification will serialize and send it through the Conn
    err = service.SendOne(notification.MakeNotification(1, "aef4429b", `{"aps":{"alert":"message"}}`))
    if err != nil {
        panic(err)
    }
    // Check if that notification was invalid by reading the response back from APNS. 
    // A timeout in the response means that it was valid. Yeaaaaa
    failure, err := service.ReadInvalid(2 * time.Second)
    if err != nil {
        panic(err)
    }
    if failure.Identifier != 0 {
        panic("#" + failure.Identifier + " failed.")
    }
}
```

# Running the tests

Clone the repo and run `go test`.

# Known issues / caveats

- Notification payloads exceeding 256 bytes won't be delivered by Apple

# Todo

- Connection pooling?
- Tests around error handling

Pull requests encouraged =)
