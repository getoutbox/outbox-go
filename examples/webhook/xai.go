//go:build xai

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
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

func registerDestination(ctx context.Context) {
	_, err := outboxClient.Destinations.Create(ctx, outbox.CreateDestinationInput{
		DestinationID: "webhook-xai",
		DisplayName:   "Webhook xAI agent",
		EventTypes:    []outbox.DestinationEventType{outbox.DestinationEventTypeMessage},
		Target: &outboxv1.Destination_Webhook{
			Webhook: &outboxv1.WebhookTarget{
				Url:           os.Getenv("WEBHOOK_URL"),
				SigningSecret: os.Getenv("OUTBOX_SIGNING_SECRET"),
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
	registerDestination(ctx)

	secret := os.Getenv("OUTBOX_SIGNING_SECRET")

	http.HandleFunc("/inbound", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		sig := r.Header.Get("X-Outbox-Signature")
		ok, err := outbox.Verify(body, secret, sig)
		if err != nil || !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		go func() {
			if err := processEvent(context.Background(), body); err != nil {
				log.Printf("process event: %v", err)
			}
		}()
	})

	port := getEnv("PORT", "3000")
	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
