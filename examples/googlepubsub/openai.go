//go:build openai

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/pubsub"
	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	openai "github.com/openai/openai-go/v3"
)

var (
	outboxClient = outbox.New(os.Getenv("OUTBOX_API_KEY"))
	openaiClient = openai.NewClient()
)

func main() {
	ctx := context.Background()

	if _, err := outboxClient.Destinations.Create(ctx, outbox.CreateDestinationInput{
		DestinationID: "pubsub-openai",
		DisplayName:   "Pub/Sub OpenAI agent",
		EventTypes:    []outbox.DestinationEventType{outbox.DestinationEventTypeMessage},
		Target: &outboxv1.Destination_GooglePubSub{
			GooglePubSub: &outboxv1.GooglePubSubTarget{
				ProjectId: os.Getenv("GOOGLE_PROJECT_ID"),
				TopicId:   os.Getenv("PUBSUB_TOPIC_ID"),
			},
		},
	}); err != nil {
		log.Printf("create destination: %v", err)
	}

	client, err := pubsub.NewClient(ctx, os.Getenv("GOOGLE_PROJECT_ID"))
	if err != nil {
		log.Fatalf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(os.Getenv("PUBSUB_SUBSCRIPTION_ID"))
	log.Println("listening for messages...")
	if err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		msg.Ack()
		if err := processEvent(ctx, msg.Data); err != nil {
			log.Printf("processEvent: %v", err)
		}
	}); err != nil {
		log.Fatalf("sub.Receive: %v", err)
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
