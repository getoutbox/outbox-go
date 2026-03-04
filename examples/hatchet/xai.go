//go:build xai

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

func main() {
	ctx := context.Background()

	_, err := outboxClient.Destinations.Create(ctx, outbox.CreateDestinationInput{
		DestinationID: "hatchet-xai",
		DisplayName:   "Hatchet xAI agent",
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
