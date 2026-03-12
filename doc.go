// Package outbox provides a Go client SDK for the Outbox messaging API.
//
// # Getting Started
//
// Create a client and send a message:
//
//	client := outbox.New("your-api-key")
//
//	result, err := client.Messages.Send(ctx, outbox.SendMessageInput{
//	    ConnectorID: "conn123",
//	    RecipientID: "acct456",
//	    Parts:       []outbox.MessagePart{outbox.TextPart("Hello!")},
//	})
//
// # Services
//
// The client exposes five services:
//
//   - [Client.Connectors] — create and manage connectors (one per channel account)
//   - [Client.Accounts] — look up or manage end-user accounts
//   - [Client.Messages] — send, update, delete, and list messages
//   - [Client.Destinations] — configure push targets for events
//   - [Client.Templates] — create and manage message templates (e.g. WhatsApp templates)
//
// # Oneof Fields
//
// [Connector.ChannelConfig] and [Destination.Target] are typed as any and hold
// proto oneof wrappers from the outboxv1 package. Use a type switch to access
// channel-specific details:
//
//	switch cfg := connector.ChannelConfig.(type) {
//	case *outboxv1.Connector_SlackBot:
//	    fmt.Println(cfg.SlackBot.BotToken)
//	case *outboxv1.Connector_WhatsappBot:
//	    fmt.Println(cfg.WhatsappBot.AppId)
//	}
//
// Pass the same oneof wrapper types when creating or updating:
//
//	client.Connectors.Create(ctx, outbox.CreateConnectorInput{
//	    ChannelConfig: &outboxv1.Connector_SlackBot{
//	        SlackBot: &outboxv1.SlackBotConfig{BotToken: "xoxb-..."},
//	    },
//	})
//
// # Webhook Handling
//
// Verify and parse incoming webhook payloads:
//
//	ok, err := outbox.Verify(body, signingSecret, r.Header.Get("X-Outbox-Signature"))
//	if err != nil || !ok {
//	    http.Error(w, "unauthorized", http.StatusUnauthorized)
//	    return
//	}
//
//	event, err := outbox.ParseDeliveryEvent(body)
//	switch e := event.(type) {
//	case *outbox.MessageDeliveryEvent:
//	    fmt.Printf("message from connector %s\n", e.ConnectorID)
//	case *outbox.DeliveryUpdateDeliveryEvent:
//	    if e.Delivery.ErrorCode != nil {
//	        fmt.Printf("delivery failed: %s\n", *e.Delivery.ErrorCode)
//	    }
//	}
package outbox
