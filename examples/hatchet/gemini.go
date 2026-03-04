//go:build gemini

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	hatchet "github.com/hatchet-dev/hatchet/sdks/go"
	"google.golang.org/genai"
)

var (
	outboxClient *outbox.Client
	geminiClient *genai.Client
)

func init() {
	outboxClient = outbox.New(os.Getenv("OUTBOX_API_KEY"))
	var err error
	geminiClient, err = genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
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
		_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
			ConnectorID: connectorID,
			AccountID:   accountID,
			Typing:      false,
		})
		return fmt.Errorf("create chat: %w", err)
	}

	resp, err := chat.Send(ctx, genai.NewPartFromText(userText))

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		Typing:      false,
	})
	if err != nil {
		return fmt.Errorf("gemini: %w", err)
	}

	reply := resp.Text()

	_, err = outboxClient.Messages.Send(ctx, outbox.SendMessageInput{
		ConnectorID: connectorID,
		RecipientID: accountID,
		Parts:       []outbox.MessagePart{outbox.TextPart(reply)},
	})
	return err
}

func main() {
	ctx := context.Background()

	_, err := outboxClient.Destinations.Create(ctx, outbox.CreateDestinationInput{
		DestinationID: "hatchet-gemini",
		DisplayName:   "Hatchet Gemini agent",
		EventTypes:    []outbox.DestinationEventType{outbox.DestinationEventTypeMessage},
		Target: &outboxv1.Destination_Hatchet{
			Hatchet: &outboxv1.HatchetTarget{
				Address:      "grpc.hatchet.run:443",
				WorkflowName: "handle-message",
				ApiToken:     os.Getenv("HATCHET_API_TOKEN"),
			},
		},
	})
	if err != nil {
		log.Fatalf("create destination: %v", err)
	}

	hatchetClient, err := hatchet.NewClient()
	if err != nil {
		log.Fatalf("hatchet client: %v", err)
	}

	task := hatchetClient.NewStandaloneTask("handle-message", func(ctx hatchet.Context, input map[string]any) (any, error) {
		body, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		return nil, processEvent(ctx, body)
	})

	worker, err := hatchetClient.NewWorker("outbox-worker", hatchet.WithWorkflows(task))
	if err != nil {
		log.Fatalf("create worker: %v", err)
	}

	if err := worker.StartBlocking(ctx); err != nil {
		log.Fatalf("start worker: %v", err)
	}
}
