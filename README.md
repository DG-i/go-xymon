go-xymon
========

A Xymon library for receiving Xymon messages and sending check results.

At the moment there's only a reader for the `xymond_channel` tool. The reader processes all known message types and works with every channel.

```go
package main

import (
	"fmt"

	"github.com/DG-i/go-xymon/channels"
)

type Handler struct{}

func (h *Handler) MessageHandler(msg channels.Message) error {
	fmt.Printf("%+v", msg)
	return nil
}
func (h *Handler) ErrorHandler(err error) { fmt.Printf("%+v", err) }

func main() {
	channelReader := channels.NewReader(&Handler{})
	channelReader.Run()
}
```