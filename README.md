go-xymon
========

A Xymon library for receiving Xymon messages and sending check results.

At the moment there's only a reader for the Xymon page channel.

```go
package main

import (
	"fmt"

	pageChannel "github.com/dg-i/go-xymon/channels/page"
)

func handleMessage(msg pageChannel.Message, errorChan chan<- error) { fmt.Printf("%+v", msg) }
func handleError(err error)                                         { fmt.Printf("%+v", err) }

func main() {
	channelReader := pageChannel.NewReader(handleMessage, handleError, log)
	channelReader.Run()
}
```
