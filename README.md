# Go push notification client

A work in progress.

`go get github.com/pranavraja/apns`

Implements The 'Enhanced Notification Format' (see [Communicating With APS](http://developer.apple.com/library/mac/#documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CommunicatingWIthAPS/CommunicatingWIthAPS.html#//apple_ref/doc/uid/TP40008194-CH101-SW1)), with handling of failures.

The algorithm used is similar to the one used in [PushSharp](https://github.com/Redth/PushSharp) - see [the problem with Apple's push notification service](http://redth.info/the-problem-with-apples-push-notification-ser/).

# Usage

Send a batch of notifications over a connection, ensuring that failed notifications don't affect the rest of the queue:

```go
package main

import (
    "github.com/pranavraja/apns"
    "os"
)

func main() {
    host := os.Getenv("APNS_HOST") // e.g. gateway.push.sandbox.apple.com:2195
    certFile := os.Getenv("CERT_FILE") // e.g. cert.pem
    keyFile := os.Getenv("KEY_FILE") // e.g. cert.private.pem
    queue := []apns.NotificationAndPayload{apns.MakeNotification(1, "aef4429b", "message")}
    // Blocks until notifications are all sent
    apns.ConnectAndSend(host, certFile, keyFile, queue)
}
```

Or alternatively, send the notifications one by one through channels, and listen for failures yourself:

```go
package main

import (
    "github.com/pranavraja/apns"
    "os"
)

func main() {
    host := os.Getenv("APNS_HOST")
    certFile := os.Getenv("CERT_FILE")
    keyFile := os.Getenv("KEY_FILE")
    write, read, err := apns.ConnectToApns(host, certFile, keyFile)
    if err != nil {
        panic(err)
    }
    write <- apns.MakeNotification(1, "aef4429b", "message")
    failure := <-read
    if failure.Identifier != 0 {
        panic("#" + failure.Identifier + " failed.")
    }
}
```

# Known issues / caveats

- The certificate and the private key can't be in the same file or `tls.LoadX509KeyPair` will freak out (maybe someone can fix this in the tls module?)
- Notification payloads exceeding 256 bytes won't be delivered by Apple

# Todo

- Configurable timeout for reading failures from Apple (currently hardcoded as 2 seconds)
- Connection pooling
- More validation and error handling (yeah, like this will ever happen)

Pull requests encouraged =)
