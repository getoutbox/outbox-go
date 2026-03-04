//go:build xai

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var (
	outboxClient = outbox.New(os.Getenv("OUTBOX_API_KEY"))
	xaiClient    = openai.NewClient(
		option.WithAPIKey(os.Getenv("XAI_API_KEY")),
		option.WithBaseURL("https://api.x.ai/v1"),
	)
)

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
		ConnectorID: input.ConnectorID,
		AccountID:   input.AccountID,
		Typing:      true,
	})

	completion, err := xaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    "grok-4",
		Messages: msgs,
	})

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: input.ConnectorID,
		AccountID:   input.AccountID,
		Typing:      false,
	})

	if err != nil {
		return fmt.Errorf("ai completion: %w", err)
	}

	reply := completion.Choices[0].Message.Content
	_, err = outboxClient.Messages.Send(ctx, outbox.SendMessageInput{
		ConnectorID: input.ConnectorID,
		RecipientID: input.AccountID,
		Parts:       []outbox.MessagePart{outbox.TextPart(reply)},
	})
	return err
}

func registerDestination(ctx context.Context) {
	_, err := outboxClient.Destinations.Create(ctx, outbox.CreateDestinationInput{
		DestinationID: "temporal-xai",
		DisplayName:   "Temporal (xAI)",
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
