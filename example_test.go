package outbox_test

import (
	"fmt"
	"log"
	"net/http"

	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"
)

// ExampleNew demonstrates creating an Outbox client.
func ExampleNew() {
	client := outbox.New("ob_live_your_api_key")
	_ = client
}

// ExampleNew_withOptions demonstrates overriding the base URL and HTTP client.
func ExampleNew_withOptions() {
	client := outbox.New(
		"ob_live_your_api_key",
		outbox.WithBaseURL("https://api.outbox.chat"),
		outbox.WithHTTPClient(&http.Client{}),
	)
	_ = client
}

// ExampleParseID demonstrates extracting the plain ID from a resource name.
func ExampleParseID() {
	fmt.Println(outbox.ParseID("connectors/abc"))
	fmt.Println(outbox.ParseID("accounts/xyz"))
	fmt.Println(outbox.ParseID("abc"))
	// Output:
	// abc
	// xyz
	// abc
}

// ExampleTextPart demonstrates creating a text/plain MessagePart.
func ExampleTextPart() {
	part := outbox.TextPart("hello")
	fmt.Println(part.ContentType)
	// Output: text/plain
}

// ExampleTextContent demonstrates decoding a MessagePart's content as text.
func ExampleTextContent() {
	part := outbox.TextPart("hello world")
	text, err := outbox.TextContent(part)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(text)
	// Output: hello world
}

// ExampleVerify demonstrates verifying an incoming webhook signature.
func ExampleVerify() {
	body := []byte(`{"delivery_id":"abc"}`)
	secret := "my-webhook-secret"
	// signature is the value of the X-Outbox-Signature header.
	signature := "not-a-real-sig"

	ok, err := outbox.Verify(body, secret, signature)
	if err != nil {
		// err is returned when signature is not valid hex.
		log.Printf("invalid signature format: %v", err)
		return
	}
	if !ok {
		log.Println("webhook signature mismatch — reject request")
		return
	}
	_ = ok
}

// ExampleParseDeliveryEvent demonstrates parsing a webhook body and dispatching
// on the concrete event type.
func ExampleParseDeliveryEvent() {
	body := []byte(`{
		"connector": "connectors/abc",
		"message": {
			"name": "messages/msg1",
			"direction": 1,
			"parts": [{"contentType": "text/plain", "content": "aGVsbG8="}]
		}
	}`)

	event, err := outbox.ParseDeliveryEvent(body)
	if err != nil {
		log.Fatal(err)
	}

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
	case *outbox.UnknownDeliveryEvent:
		fmt.Println("unknown event from connector:", e.ConnectorID)
	}
	// Output: message: hello
}

// ExampleConnectorsService_Create_channelConfig demonstrates passing a
// channel-specific config when creating a Connector. The ChannelConfig field
// accepts any *outboxv1.Connector_Xxx oneof wrapper.
func ExampleConnectorsService_Create_channelConfig() {
	client := outbox.New("ob_live_your_api_key")
	_ = client

	input := outbox.CreateConnectorInput{
		ChannelConfig: &outboxv1.Connector_SlackBot{
			SlackBot: &outboxv1.SlackBotConfig{
				BotToken: "xoxb-...",
			},
		},
	}
	_ = input
}
