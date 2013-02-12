# Go push notification client

A work in progress, may be rebased without warning!

`go get github.com/pranavraja/apns`

Implements The 'Enhanced Notification Format' (see [Communicating With APS](http://developer.apple.com/library/mac/#documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/CommunicatingWIthAPS/CommunicatingWIthAPS.html#//apple_ref/doc/uid/TP40008194-CH101-SW1)), with handling of failures.

The algorithm used is similar to the one used in [PushSharp](https://github.com/Redth/PushSharp) - see [the problem with Apple's push notification service](http://redth.info/the-problem-with-apples-push-notification-ser/).

# Usage

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
    queue := []apns.NotificationAndPayload{apns.MakeNotification(1, "aef4429b", "message")}
    // Blocks until notifications are all sent
    apns.ConnectAndSend(host, certFile, keyFile, queue)
}
```

