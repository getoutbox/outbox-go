# outbox-go

Go SDK for [Outbox](https://outbox.chat) — a unified messaging API for AI agents.

Send and receive messages across channels (Slack, WhatsApp, and more) with a single API.

## Installation

```bash
go get github.com/getoutbox/outbox-go
```

Requires Go 1.25 or later.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	outbox "github.com/getoutbox/outbox-go"
)

func main() {
	client := outbox.New("ob_live_your_api_key")

	result, err := client.Messages.Send(context.Background(), outbox.SendMessageInput{
		ConnectorID: "conn123",
		RecipientID: "acct456",
		Parts:       []outbox.MessagePart{outbox.TextPart("Hello!")},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("sent:", result.Message.ID)
}
```

## Services

The client exposes five namespaces:

| Namespace | Description |
|-----------|-------------|
| `client.Connectors` | Create and manage connectors (one per channel account) |
| `client.Accounts` | Look up or manage end-user accounts |
| `client.Messages` | Send, update, delete, and list messages |
| `client.Destinations` | Configure push targets for delivery events |
| `client.Templates` | Create and manage message templates (e.g. WhatsApp templates) |

## Webhook verification

Verify and parse incoming webhook payloads:

```go
ok, err := outbox.Verify(body, signingSecret, r.Header.Get("X-Outbox-Signature"))
if err != nil || !ok {
	http.Error(w, "unauthorized", http.StatusUnauthorized)
	return
}

event, err := outbox.ParseDeliveryEvent(body)
switch e := event.(type) {
case *outbox.MessageDeliveryEvent:
	text, _ := outbox.TextContent(e.Message.Parts[0])
	fmt.Println("message:", text)
case *outbox.DeliveryUpdateDeliveryEvent:
	fmt.Println("delivery update:", e.Delivery.Status)
case *outbox.ReadReceiptDeliveryEvent:
	fmt.Println("read receipt:", e.ReadReceipt.MessageIDs)
case *outbox.TypingIndicatorDeliveryEvent:
	fmt.Println("typing:", e.TypingIndicator.Typing)
}
```

## Oneof fields

`Connector.ChannelConfig` and `Destination.Target` hold proto oneof wrappers from the `outboxv1` package. Use a type switch:

```go
switch cfg := connector.ChannelConfig.(type) {
case *outboxv1.Connector_SlackBot:
	fmt.Println(cfg.SlackBot.BotToken)
case *outboxv1.Connector_WhatsappBot:
	fmt.Println(cfg.WhatsappBot.AppId)
}
```

## Client options

```go
client := outbox.New(
	"ob_live_your_api_key",
	outbox.WithBaseURL("https://custom-api.example.com"),
	outbox.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
)
```

## Examples

See the [`examples/`](examples/) directory for integration examples with:

EventBridge, Google Pub/Sub, Hatchet, Inngest, Kafka, Lambda, NATS, Restate, SNS, SQS, Temporal, Webhooks

## License

MIT
