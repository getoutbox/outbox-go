//go:build gemini

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
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

func main() {
	ctx := context.Background()
	registerDestination(ctx)

	c, err := client.Dial(client.Options{
		HostPort: getEnv("TEMPORAL_ADDRESS", "localhost:7233"),
	})
	if err != nil {
		log.Fatalf("failed to dial temporal: %v", err)
	}
	defer c.Close()

	w := worker.New(c, getEnv("TEMPORAL_TASK_QUEUE", "outbox-task-queue"), worker.Options{})
	w.RegisterWorkflow(MessageWorkflow)
	w.RegisterActivity(processMessage)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("worker error: %v", err)
	}
}

func processMessage(ctx context.Context, input ProcessMessageInput) error {
	if err := outboxClient.Messages.MarkRead(ctx, outbox.MarkReadInput{
		ConnectorID: input.ConnectorID,
		AccountID:   input.AccountID,
		MessageIDs:  []string{input.MessageID},
	}); err != nil {
		return fmt.Errorf("mark read: %w", err)
	}

	history, err := outboxClient.Messages.History(ctx, outbox.HistoryInput{
		ConnectorID: input.ConnectorID,
		AccountID:   input.AccountID,
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
	userText := input.UserText
	if len(chatHistory) > 1 {
		prior = chatHistory[:len(chatHistory)-1]
	}

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: input.ConnectorID,
		AccountID:   input.AccountID,
		Typing:      true,
	})

	chat, err := geminiClient.Chats.Create(ctx, "gemini-2.5-flash", nil, prior)
	if err != nil {
		_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
			ConnectorID: input.ConnectorID,
			AccountID:   input.AccountID,
			Typing:      false,
		})
		return fmt.Errorf("create chat: %w", err)
	}

	resp, err := chat.Send(ctx, genai.NewPartFromText(userText))

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: input.ConnectorID,
		AccountID:   input.AccountID,
		Typing:      false,
	})

	if err != nil {
		return fmt.Errorf("ai completion: %w", err)
	}

	reply := resp.Text()
	_, err = outboxClient.Messages.Send(ctx, outbox.SendMessageInput{
		ConnectorID: input.ConnectorID,
		RecipientID: input.AccountID,
		Parts:       []outbox.MessagePart{outbox.TextPart(reply)},
	})
	return err
}

func registerDestination(ctx context.Context) {
	_, err := outboxClient.Destinations.Create(ctx, outbox.CreateDestinationInput{
		DestinationID: "temporal-gemini",
		DisplayName:   "Temporal (Gemini)",
		EventTypes:    []outbox.DestinationEventType{outbox.DestinationEventTypeMessage},
		Target: &outboxv1.Destination_Temporal{
			Temporal: &outboxv1.TemporalTarget{
				Address:      getEnv("TEMPORAL_ADDRESS", "localhost:7233"),
				Namespace:    getEnv("TEMPORAL_NAMESPACE", "default"),
				TaskQueue:    getEnv("TEMPORAL_TASK_QUEUE", "outbox-task-queue"),
				WorkflowType: "MessageWorkflow",
			},
		},
	})
	if err != nil {
		log.Printf("register destination: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
