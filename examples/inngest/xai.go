//go:build xai

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	"github.com/inngest/inngestgo"
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

func registerDestination(ctx context.Context) {
	_, err := outboxClient.Destinations.Create(ctx, outbox.CreateDestinationInput{
		DestinationID: "inngest-xai",
		DisplayName:   "Inngest xAI agent",
		EventTypes:    []outbox.DestinationEventType{outbox.DestinationEventTypeMessage},
		Target: &outboxv1.Destination_Inngest{
			Inngest: &outboxv1.InngestTarget{
				Url:      getEnv("INNGEST_EVENT_URL", "https://inn.gs/e/"),
				EventKey: os.Getenv("INNGEST_EVENT_KEY"),
			},
		},
	})
	if err != nil {
		log.Fatalf("create destination: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	registerDestination(context.Background())

	client, err := inngestgo.NewClient(inngestgo.ClientOpts{AppID: "outbox-agent"})
	if err != nil {
		log.Fatalf("inngest client: %v", err)
	}

	_, err = inngestgo.CreateFunction(
		client,
		inngestgo.FunctionOpts{ID: "handle-message"},
		inngestgo.EventTrigger("outbox/message", nil),
		func(ctx context.Context, input inngestgo.Input[map[string]any]) (any, error) {
			body, err := json.Marshal(input.Event.Data)
			if err != nil {
				return nil, err
			}
			return nil, processEvent(ctx, body)
		},
	)
	if err != nil {
		log.Fatalf("create function: %v", err)
	}

	log.Fatal(http.ListenAndServe(":"+getEnv("PORT", "3000"), client.Serve()))
}
