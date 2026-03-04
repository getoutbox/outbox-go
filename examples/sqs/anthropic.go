//go:build anthropic

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
)

var (
	outboxClient    = outbox.New(os.Getenv("OUTBOX_API_KEY"))
	anthropicClient = anthropic.NewClient()
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

	var msgs []anthropic.MessageParam
	for _, m := range history.Items {
		if len(m.Parts) == 0 {
			continue
		}
		text, _ := outbox.TextContent(m.Parts[0])
		if m.Direction == outbox.MessageDirectionInbound {
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(text)))
		} else {
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(text)))
		}
	}

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		Typing:      true,
	})

	resp, err := anthropicClient.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     "claude-sonnet-4-6-20251001",
		MaxTokens: 1024,
		Messages:  msgs,
	})

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		Typing:      false,
	})
	if err != nil {
		return fmt.Errorf("ai completion: %w", err)
	}

	var reply string
	for _, block := range resp.Content {
		if block.Type == "text" {
			reply = block.Text
			break
		}
	}

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
		DestinationID: "sqs-anthropic",
		DisplayName:   "SQS Anthropic agent",
		EventTypes:    []outbox.DestinationEventType{outbox.DestinationEventTypeMessage},
		Target: &outboxv1.Destination_Sqs{
			Sqs: &outboxv1.SqsTarget{
				QueueUrl:        os.Getenv("SQS_QUEUE_URL"),
				Region:          getEnv("AWS_REGION", "us-east-1"),
				AccessKeyId:     os.Getenv("AWS_ACCESS_KEY_ID"),
				SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			},
		},
	}); err != nil {
		log.Printf("create destination: %v", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("load aws config: %v", err)
	}
	sqsClient := sqs.NewFromConfig(cfg)
	queueURL := os.Getenv("SQS_QUEUE_URL")

	log.Println("polling SQS queue...")
	for {
		out, err := sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20,
		})
		if err != nil {
			log.Printf("receive message: %v", err)
			continue
		}
		for _, sqsMsg := range out.Messages {
			if err := processEvent(ctx, []byte(*sqsMsg.Body)); err != nil {
				log.Printf("process event: %v", err)
			}
			if _, err := sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      aws.String(queueURL),
				ReceiptHandle: sqsMsg.ReceiptHandle,
			}); err != nil {
				log.Printf("delete message: %v", err)
			}
		}
	}
}
