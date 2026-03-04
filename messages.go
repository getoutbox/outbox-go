package outbox

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	outboxv1connect "github.com/getoutbox/outbox-go/outboxv1/outboxv1connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// MessagesService provides operations on Messages.
type MessagesService struct {
	client outboxv1connect.MessageServiceClient
}

// SendMessageInput holds parameters for sending a Message.
type SendMessageInput struct {
	ConnectorID      string
	RecipientID      string
	Parts            []MessagePart
	ReplyToMessageID string // optional; empty = no reply
	GroupID          string // optional
	Metadata         map[string]string
	RequestID        string
}

// SendMessageResult holds the result of a Send call.
type SendMessageResult struct {
	Message  Message
	Delivery *MessageDelivery // nil if not returned by the server
}

// UpdateMessageInput holds parameters for updating (editing) a Message.
// Only non-nil fields are included in the update mask.
type UpdateMessageInput struct {
	ID        string
	Parts     []MessagePart     // nil = don't update
	Metadata  map[string]string // nil = don't update
	RequestID string
}

// DeleteMessageInput holds parameters for deleting a Message.
type DeleteMessageInput struct {
	ID            string
	DeletionScope MessageDeletionScope
	RequestID     string
}

// ListMessagesOptions configures a List request.
type ListMessagesOptions struct {
	ConnectorID string
	PageSize    int32
	PageToken   string
	Filter      string
	OrderBy     string
}

// ListMessagesResult is the paginated result of a List call.
type ListMessagesResult struct {
	Items         []Message
	NextPageToken string
	TotalSize     int64
}

// MarkReadInput holds parameters for sending read receipts.
type MarkReadInput struct {
	ConnectorID string
	AccountID   string
	MessageIDs  []string // plain IDs; the SDK builds resource names internally
}

// TypingInput holds parameters for sending a typing indicator.
type TypingInput struct {
	ConnectorID string
	AccountID   string
	// Typing indicates whether the account has started (true) or stopped (false) typing.
	Typing bool
}

// HistoryInput holds parameters for fetching conversation history for a specific
// account on a Connector. Results are always ordered oldest-first by create time.
type HistoryInput struct {
	ConnectorID string
	AccountID   string
	PageSize    int32
	PageToken   string
}

// Send creates and sends a Message on the given Connector.
func (s *MessagesService) Send(ctx context.Context, input SendMessageInput) (*SendMessageResult, error) {
	msg := &outboxv1.Message{
		Recipient: "accounts/" + input.RecipientID,
		Parts:     toProtoParts(input.Parts),
		Metadata:  input.Metadata,
	}
	if input.ReplyToMessageID != "" {
		msg.ReplyTo = "messages/" + input.ReplyToMessageID
	}
	if input.GroupID != "" {
		msg.GroupId = input.GroupID
	}
	res, err := s.client.CreateMessage(ctx, connect.NewRequest(&outboxv1.CreateMessageRequest{
		Connector: "connectors/" + input.ConnectorID,
		Message:   msg,
		RequestId: input.RequestID,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Message == nil {
		return nil, errEmpty("CreateMessage")
	}
	result := &SendMessageResult{
		Message: mapMessage(res.Msg.Message),
	}
	if res.Msg.Delivery != nil {
		d := mapMessageDelivery(res.Msg.Delivery)
		result.Delivery = &d
	}
	return result, nil
}

// Get retrieves a single Message by its plain ID.
// Messages are scoped to a connector; use List with a filter for bulk retrieval.
func (s *MessagesService) Get(ctx context.Context, id string) (*Message, error) {
	if strings.ContainsAny(id, `"`) {
		return nil, fmt.Errorf("outbox: invalid message ID %q", id)
	}
	res, err := s.client.ListMessages(ctx, connect.NewRequest(&outboxv1.ListMessagesRequest{
		Filter:   `name == "messages/` + id + `"`,
		PageSize: 1,
	}))
	if err != nil {
		return nil, err
	}
	if len(res.Msg.Messages) == 0 {
		return nil, errEmpty("GetMessage")
	}
	m := mapMessage(res.Msg.Messages[0])
	return &m, nil
}

// List returns a paginated list of Messages for a Connector.
func (s *MessagesService) List(ctx context.Context, opts *ListMessagesOptions) (*ListMessagesResult, error) {
	r := &outboxv1.ListMessagesRequest{}
	if opts != nil {
		if opts.ConnectorID != "" {
			r.Parent = "connectors/" + opts.ConnectorID
		}
		r.PageSize = opts.PageSize
		r.PageToken = opts.PageToken
		r.Filter = opts.Filter
		r.OrderBy = opts.OrderBy
	}
	res, err := s.client.ListMessages(ctx, connect.NewRequest(r))
	if err != nil {
		return nil, err
	}
	items := make([]Message, len(res.Msg.Messages))
	for i, m := range res.Msg.Messages {
		items[i] = mapMessage(m)
	}
	return &ListMessagesResult{
		Items:         items,
		NextPageToken: res.Msg.NextPageToken,
		TotalSize:     int64(res.Msg.TotalSize),
	}, nil
}

// Update edits a previously sent Message.
func (s *MessagesService) Update(ctx context.Context, input UpdateMessageInput) (*Message, error) {
	msg := &outboxv1.Message{Name: "messages/" + input.ID}
	var paths []string

	if input.Parts != nil {
		msg.Parts = toProtoParts(input.Parts)
		paths = append(paths, "parts")
	}
	if input.Metadata != nil {
		msg.Metadata = input.Metadata
		paths = append(paths, "metadata")
	}

	var mask *fieldmaskpb.FieldMask
	if len(paths) > 0 {
		mask = &fieldmaskpb.FieldMask{Paths: paths}
	}

	res, err := s.client.UpdateMessage(ctx, connect.NewRequest(&outboxv1.UpdateMessageRequest{
		Message:    msg,
		UpdateMask: mask,
		RequestId:  input.RequestID,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Message == nil {
		return nil, errEmpty("UpdateMessage")
	}
	m := mapMessage(res.Msg.Message)
	return &m, nil
}

// Delete deletes a Message.
func (s *MessagesService) Delete(ctx context.Context, input DeleteMessageInput) (*Message, error) {
	res, err := s.client.DeleteMessage(ctx, connect.NewRequest(&outboxv1.DeleteMessageRequest{
		Name:      "messages/" + input.ID,
		Scope:     input.DeletionScope,
		RequestId: input.RequestID,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Message == nil {
		return nil, errEmpty("DeleteMessage")
	}
	m := mapMessage(res.Msg.Message)
	return &m, nil
}

// MarkRead sends read receipts for the given messages on a Connector.
func (s *MessagesService) MarkRead(ctx context.Context, input MarkReadInput) error {
	names := make([]string, len(input.MessageIDs))
	for i, id := range input.MessageIDs {
		names[i] = "messages/" + id
	}
	_, err := s.client.SendReadReceipt(ctx, connect.NewRequest(&outboxv1.SendReadReceiptRequest{
		Connector: "connectors/" + input.ConnectorID,
		Account:   "accounts/" + input.AccountID,
		Messages:  names,
	}))
	return err
}

// SendTypingIndicator sends a typing indicator on a Connector.
func (s *MessagesService) SendTypingIndicator(ctx context.Context, input TypingInput) error {
	typing := input.Typing
	_, err := s.client.SendTypingIndicator(ctx, connect.NewRequest(&outboxv1.SendTypingIndicatorRequest{
		Connector: "connectors/" + input.ConnectorID,
		Account:   "accounts/" + input.AccountID,
		Typing:    &typing,
	}))
	return err
}

// History fetches the conversation history for a specific account on a Connector,
// ordered oldest-first.
func (s *MessagesService) History(ctx context.Context, input HistoryInput) (*ListMessagesResult, error) {
	if strings.ContainsAny(input.AccountID, `"`) {
		return nil, fmt.Errorf("outbox: invalid account ID %q", input.AccountID)
	}
	filter := `account.name == "accounts/` + input.AccountID + `"`
	return s.List(ctx, &ListMessagesOptions{
		ConnectorID: input.ConnectorID,
		PageSize:    input.PageSize,
		PageToken:   input.PageToken,
		Filter:      filter,
		OrderBy:     "create_time asc",
	})
}

// toProtoParts converts a slice of domain MessagePart values to proto MessagePart pointers.
func toProtoParts(parts []MessagePart) []*outboxv1.MessagePart {
	if parts == nil {
		return nil
	}
	out := make([]*outboxv1.MessagePart, len(parts))
	for i, p := range parts {
		pp := &outboxv1.MessagePart{
			ContentType: p.ContentType,
			Disposition: p.Disposition,
			Filename:    p.Filename,
		}
		if p.URL != "" {
			pp.Source = &outboxv1.MessagePart_Url{Url: p.URL}
		} else if p.Content != nil {
			content := make([]byte, len(p.Content))
			copy(content, p.Content)
			pp.Source = &outboxv1.MessagePart_Content{Content: content}
		}
		out[i] = pp
	}
	return out
}
