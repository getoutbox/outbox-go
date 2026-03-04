//go:build anthropic

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/aws/aws-lambda-go/lambda"
	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
)

var (
	outboxClient    = outbox.New(os.Getenv("OUTBOX_API_KEY"))
	anthropicClient = anthropic.NewClient()
)

type Event struct {
	Body string `json:"body"`
}

func registerDestination(ctx context.Context) {
	_, err := outboxClient.Destinations.Create(ctx, outbox.CreateDestinationInput{
		DestinationID: "lambda-anthropic",
		DisplayName:   "Lambda Anthropic agent",
		EventTypes:    []outbox.DestinationEventType{outbox.DestinationEventTypeMessage},
		Target: &outboxv1.Destination_Lambda{
			Lambda: &outboxv1.LambdaTarget{
				Url: os.Getenv("LAMBDA_FUNCTION_URL"),
			},
		},
	})
	if err != nil {
		log.Printf("register destination: %v", err)
	}
}

func processEvent(ctx context.Context, body []byte) error {
	event, err := outbox.ParseDeliveryEvent(body)
	if err != nil {
		return fmt.Errorf("parse event: %w", err)
	}
	me, ok := event.(*outbox.MessageDeliveryEvent)
	if !ok {
		return nil
	}
	msg := me.Message
	if msg.Account == nil || len(msg.Parts) == 0 {
		return nil
	}
	connectorID := me.ConnectorID
	accountID := msg.Account.ID

	if err := outboxClient.Messages.MarkRead(ctx, outbox.MarkReadInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		MessageIDs:  []string{msg.ID},
	}); err != nil {
		return fmt.Errorf("mark read: %w", err)
	}

	history, err := outboxClient.Messages.History(ctx, outbox.HistoryInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		PageSize:    20,
	})
	if err != nil {
		return fmt.Errorf("history: %w", err)
	}

	var msgs []anthropic.MessageParam
	for _, m := range history.Items {
		if len(m.Parts) == 0 {
			continue
		}
		text, _ := outbox.TextContent(m.Parts[0])
		if m.Direction == outbox.MessageDirectionInbound {
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(text)))
		} else {
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(text)))
		}
	}

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		Typing:      true,
	})

	resp, err := anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-sonnet-4-6-20251001",
		MaxTokens: 1024,
		Messages:  msgs,
	})

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		Typing:      false,
	})
	if err != nil {
		return fmt.Errorf("ai completion: %w", err)
	}

	var reply string
	for _, block := range resp.Content {
		if block.Type == "text" {
			reply = block.Text
			break
		}
	}

	_, err = outboxClient.Messages.Send(ctx, outbox.SendMessageInput{
		ConnectorID: connectorID,
		RecipientID: accountID,
		Parts:       []outbox.MessagePart{outbox.TextPart(reply)},
	})
	return err
}

func handler(ctx context.Context, event Event) (map[string]int, error) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(event.Body), &payload); err != nil {
		log.Printf("parse body: %v", err)
		return map[string]int{"statusCode": 400}, nil
	}
	if err := processEvent(ctx, []byte(event.Body)); err != nil {
		log.Printf("process event: %v", err)
	}
	return map[string]int{"statusCode": 200}, nil
}

func main() {
	// Call registerDestination once to set up the Lambda destination:
	// registerDestination(context.Background())
	lambda.Start(handler)
}
