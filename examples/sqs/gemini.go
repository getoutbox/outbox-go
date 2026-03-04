//go:build gemini

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
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
		_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{ConnectorID: connectorID, AccountID: accountID, Typing: false})
		return fmt.Errorf("create chat: %w", err)
	}
	resp, err := chat.Send(ctx, genai.NewPartFromText(userText))

	_ = outboxClient.Messages.SendTypingIndicator(ctx, outbox.TypingInput{
		ConnectorID: connectorID,
		AccountID:   accountID,
		Typing:      false,
	})
	if err != nil {
		return fmt.Errorf("ai completion: %w", err)
	}

	reply := resp.Text()

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
		DestinationID: "sqs-gemini",
		DisplayName:   "SQS Gemini agent",
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
