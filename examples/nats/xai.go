//go:build xai

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	nats "github.com/nats-io/nats.go"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var (
	outboxClient = outbox.New(os.Getenv("OUTBOX_API_KEY"))
	xaiClient    = openai.NewClient(
		option.WithAPIKey(os.Getenv("XAI_API_KEY")),
		option.WithBaseURL("https://api.x.ai/v1"),
	)
)

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

	var msgs []openai.ChatCompletionMessageParamUnion
	for _, m := range history.Items {
		if len(m.Parts) == 0 {
			continue
		}
		text, _ := outbox.TextContent(m.Parts[0])
		if m.Direction == outbox.MessageDirectionInbound {
			msgs = append(msgs, openai.UserMessage(text))
		} else {
			msgs = append(msgs, openai.AssistantMessage(text))
		}
	}

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		Typing:      true,
	})

	completion, err := xaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    "grok-4",
		Messages: msgs,
	})

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		Typing:      false,
	})
	if err != nil {
		return fmt.Errorf("ai completion: %w", err)
	}

	reply := completion.Choices[0].Message.Content

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

	if _, err := outboxClient.Destinations.Create(ctx, outbox.CreateDestinationInput{
		DestinationID: "nats-xai",
		DisplayName:   "NATS xAI agent",
		EventTypes:    []outbox.DestinationEventType{outbox.DestinationEventTypeMessage},
		Target: &outboxv1.Destination_Nats{
			Nats: &outboxv1.NatsTarget{
				Url:     getEnv("NATS_URL", "nats://localhost:4222"),
				Subject: getEnv("NATS_SUBJECT", "outbox.events"),
			},
		},
	}); err != nil {
		log.Fatalf("create destination: %v", err)
	}

	nc, err := nats.Connect(getEnv("NATS_URL", "nats://localhost:4222"))
	if err != nil {
		log.Fatalf("nats connect: %v", err)
	}
	defer nc.Drain()

	if _, err := nc.Subscribe(getEnv("NATS_SUBJECT", "outbox.events"), func(m *nats.Msg) {
		if err := processEvent(context.Background(), m.Data); err != nil {
			log.Printf("process event: %v", err)
		}
	}); err != nil {
		log.Fatalf("nats subscribe: %v", err)
	}

	log.Printf("Subscribed to NATS subject %s", getEnv("NATS_SUBJECT", "outbox.events"))
	runtime.Goexit()
}
