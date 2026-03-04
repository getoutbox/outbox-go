//go:build openai

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	openai "github.com/openai/openai-go/v3"
	kafka "github.com/segmentio/kafka-go"
)

var (
	outboxClient = outbox.New(os.Getenv("OUTBOX_API_KEY"))
	openaiClient = openai.NewClient()
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

	completion, err := openaiClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    "gpt-5.2",
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

	_, err := outboxClient.Destinations.Create(ctx, outbox.CreateDestinationInput{
		DestinationID: "kafka-openai",
		DisplayName:   "Kafka OpenAI agent",
		EventTypes:    []outbox.DestinationEventType{outbox.DestinationEventTypeMessage},
		Target: &outboxv1.Destination_Kafka{
			Kafka: &outboxv1.KafkaTarget{
				Brokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
				Topic:   getEnv("KAFKA_TOPIC", "outbox-events"),
			},
		},
	})
	if err != nil {
		log.Fatalf("create destination: %v", err)
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		Topic:   getEnv("KAFKA_TOPIC", "outbox-events"),
		GroupID: "outbox-agent",
	})
	defer r.Close()

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			break
		}
		if err := processEvent(ctx, m.Value); err != nil {
			log.Printf("process event: %v", err)
		}
	}
}
