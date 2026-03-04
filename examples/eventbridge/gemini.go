//go:build gemini

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	"google.golang.org/genai"
)

var (
	outboxClient = outbox.New(os.Getenv("OUTBOX_API_KEY"))
	geminiClient *genai.Client
)

func init() {
	var err error
	geminiClient, err = genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}
}

type EventBridgeEvent struct {
	Detail json.RawMessage `json:"detail"`
}

func registerDestination(ctx context.Context) {
	_, err := outboxClient.Destinations.Create(ctx, outbox.CreateDestinationInput{
		DestinationID: "eventbridge-gemini",
		DisplayName:   "EventBridge Gemini agent",
		EventTypes:    []outbox.DestinationEventType{outbox.DestinationEventTypeMessage},
		Target: &outboxv1.Destination_EventBridge{
			EventBridge: &outboxv1.EventBridgeTarget{
				EventBus:        os.Getenv("EVENTBRIDGE_BUS_NAME"),
				Region:          getEnv("AWS_REGION", "us-east-1"),
				AccessKeyId:     os.Getenv("AWS_ACCESS_KEY_ID"),
				SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			},
		},
	})
	if err != nil {
		log.Printf("register destination: %v", err)
	}
}

func handler(ctx context.Context, event EventBridgeEvent) error {
	return processEvent(ctx, event.Detail)
}

func processEvent(ctx context.Context, body []byte) error {
	ev, err := outbox.ParseDeliveryEvent(body)
	if err != nil {
		return fmt.Errorf("parse event: %w", err)
	}
	me, ok := ev.(*outbox.MessageDeliveryEvent)
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

	var chatHistory []*genai.Content
	for _, m := range history.Items {
		if len(m.Parts) == 0 {
			continue
		}
		text, _ := outbox.TextContent(m.Parts[0])
		role := "user"
		if m.Direction == outbox.MessageDirectionOutbound {
			role = "model"
		}
		chatHistory = append(chatHistory, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{{Text: text}},
		})
	}

	var prior []*genai.Content
	var userText string
	if len(chatHistory) > 0 {
		prior = chatHistory[:len(chatHistory)-1]
		userText, _ = outbox.TextContent(history.Items[len(history.Items)-1].Parts[0])
	}

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		Typing:      true,
	})

	chat, err := geminiClient.Chats.Create(ctx, "gemini-2.5-flash", nil, prior)
	if err != nil {
		_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{ConnectorID: connectorID, AccountID: accountID, Typing: false})
		return fmt.Errorf("create chat: %w", err)
	}
	resp, err := chat.Send(ctx, genai.NewPartFromText(userText))

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		Typing:      false,
	})

	if err != nil {
		return fmt.Errorf("ai completion: %w", err)
	}

	reply := resp.Text()

	_, err = outboxClient.Messages.Send(ctx, outbox.SendMessageInput{
		ConnectorID: connectorID,
		RecipientID: accountID,
		Parts:       []outbox.MessagePart{outbox.TextPart(reply)},
	})
	return err
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	ctx := context.Background()
	registerDestination(ctx)
	lambda.Start(handler)
}
