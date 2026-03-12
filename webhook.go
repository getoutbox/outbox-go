package outbox

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"strings"
	"time"

	outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// ParseID extracts the plain ID from a resource name.
// "connectors/abc" → "abc". Returns the full string if no "/" found.
func ParseID(name string) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

// TextPart creates a text/plain MessagePart from a UTF-8 string.
func TextPart(text string) MessagePart {
	return MessagePart{
		ContentType: "text/plain",
		Content:     []byte(text),
	}
}

// TextContent decodes a MessagePart's content bytes as UTF-8.
// Returns an error if Content is nil.
func TextContent(part MessagePart) (string, error) {
	if part.Content == nil {
		return "", fmt.Errorf("outbox: MessagePart has no content")
	}
	return string(part.Content), nil
}

// ParseDeliveryEvent parses a webhook request body into a DeliveryEvent.
// Accepts proto-binary (when destination PayloadFormat is PROTO_BINARY) or JSON.
// Format is detected automatically: if first byte is not '{', treated as proto-binary.
func ParseDeliveryEvent(body []byte) (DeliveryEvent, error) {
	p := &outboxv1.DeliveryEvent{}
	if len(body) > 0 && body[0] != '{' {
		if err := proto.Unmarshal(body, p); err != nil {
			return nil, fmt.Errorf("outbox: parse proto-binary delivery event: %w", err)
		}
	} else {
		opts := protojson.UnmarshalOptions{DiscardUnknown: true}
		if err := opts.Unmarshal(body, p); err != nil {
			return nil, fmt.Errorf("outbox: parse JSON delivery event: %w", err)
		}
	}
	return mapDeliveryEvent(p), nil
}

// mapDeliveryEvent converts an outboxv1.DeliveryEvent proto to a DeliveryEvent.
func mapDeliveryEvent(p *outboxv1.DeliveryEvent) DeliveryEvent {
	envelope := DeliveryEventEnvelope{
		DeliveryID:    p.GetDeliveryId(),
		DestinationID: ParseID(p.GetDestination()),
		ConnectorID:   ParseID(p.GetConnector()),
		EnqueueTime:   protoTime(p.GetEnqueueTime()),
	}
	switch e := p.Event.(type) {
	case *outboxv1.DeliveryEvent_Message:
		return &MessageDeliveryEvent{
			DeliveryEventEnvelope: envelope,
			Message:               mapMessage(e.Message),
		}
	case *outboxv1.DeliveryEvent_DeliveryUpdate:
		return &DeliveryUpdateDeliveryEvent{
			DeliveryEventEnvelope: envelope,
			Delivery:              mapMessageDelivery(e.DeliveryUpdate),
		}
	case *outboxv1.DeliveryEvent_ReadReceipt:
		return &ReadReceiptDeliveryEvent{
			DeliveryEventEnvelope: envelope,
			ReadReceipt:           mapReadReceipt(e.ReadReceipt),
		}
	case *outboxv1.DeliveryEvent_TypingIndicator:
		return &TypingIndicatorDeliveryEvent{
			DeliveryEventEnvelope: envelope,
			TypingIndicator:       mapTypingIndicator(e.TypingIndicator),
		}
	default:
		return &UnknownDeliveryEvent{DeliveryEventEnvelope: envelope}
	}
}

func mapAccount(p *outboxv1.Account) Account {
	return Account{
		ID:         ParseID(p.GetName()),
		ContactID:  p.GetContactId(),
		ExternalID: p.GetExternalId(),
		Source:     p.GetSource(),
		Metadata:   cloneStringMap(p.GetMetadata()),
		CreateTime: protoTime(p.GetCreateTime()),
		UpdateTime: protoTime(p.GetUpdateTime()),
	}
}

func mapAccountPtr(p *outboxv1.Account) *Account {
	if p == nil {
		return nil
	}
	a := mapAccount(p)
	return &a
}

func cloneStringMap(m map[string]string) map[string]string {
	return maps.Clone(m)
}

func mapMessagePart(p *outboxv1.MessagePart) MessagePart {
	part := MessagePart{
		ContentType: p.GetContentType(),
		Disposition: p.GetDisposition(),
		Filename:    p.GetFilename(),
	}
	switch s := p.Source.(type) {
	case *outboxv1.MessagePart_Content:
		if s.Content != nil {
			dst := make([]byte, len(s.Content))
			copy(dst, s.Content)
			part.Content = dst
		}
	case *outboxv1.MessagePart_Url:
		part.URL = s.Url
	}
	return part
}

func mapMessage(p *outboxv1.Message) Message {
	parts := make([]MessagePart, len(p.GetParts()))
	for i, pp := range p.GetParts() {
		parts[i] = mapMessagePart(pp)
	}
	msg := Message{
		ID:            ParseID(p.GetName()),
		Account:       mapAccountPtr(p.GetAccount()),
		RecipientID:   ParseID(p.GetRecipient()),
		Parts:         parts,
		Metadata:      cloneStringMap(p.GetMetadata()),
		Direction:     p.GetDirection(),
		DeletionScope: p.GetScope(),
		EditNumber:    p.GetEditNumber(),
		CreateTime:    protoTime(p.GetCreateTime()),
		DeliverTime:   protoTime(p.GetDeliverTime()),
		DeleteTime:    protoTime(p.GetDeleteTime()),
	}
	if rt := p.GetReplyTo(); rt != "" {
		msg.ReplyToMessageID = ParseID(rt)
	}
	if gid := p.GetGroupId(); gid != "" {
		msg.GroupID = gid
	}
	if p.Replaced != nil {
		id := ParseID(*p.Replaced)
		msg.ReplacedMessageID = &id
	}
	return msg
}

func mapMessageDelivery(p *outboxv1.MessageDelivery) MessageDelivery {
	return MessageDelivery{
		MessageID:        ParseID(p.GetMessage()),
		Account:          mapAccountPtr(p.GetAccount()),
		Status:           p.GetStatus(),
		ErrorCode:        p.ErrorCode,
		ErrorMessage:     p.ErrorMessage,
		StatusChangeTime: protoTime(p.GetStatusChangeTime()),
	}
}

func mapReadReceipt(p *outboxv1.ReadReceiptEvent) ReadReceiptEvent {
	ids := make([]string, len(p.GetMessages()))
	for i, name := range p.GetMessages() {
		ids[i] = ParseID(name)
	}
	return ReadReceiptEvent{
		Account:    mapAccountPtr(p.GetAccount()),
		MessageIDs: ids,
		Timestamp:  protoTime(p.GetTimestamp()),
	}
}

func mapTypingIndicator(p *outboxv1.TypingIndicatorEvent) TypingIndicatorEvent {
	return TypingIndicatorEvent{
		Account:     mapAccountPtr(p.GetAccount()),
		Typing:      p.GetTyping(),
		ContentType: p.GetContentType(),
		Timestamp:   protoTime(p.GetTimestamp()),
	}
}

// Verify checks an HMAC-SHA256 webhook signature using constant-time comparison.
// signature is a lowercase hex string (value of the X-Outbox-Signature header).
// Returns (true, nil) if valid, (false, nil) if signature is wrong,
// (false, err) if the hex cannot be decoded.
func Verify(body []byte, secret, signature string) (bool, error) {
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("outbox: decode signature hex: %w", err)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(mac.Sum(nil), sigBytes), nil
}

// protoTime converts a protobuf Timestamp to a time.Time. Returns zero time if ts is nil.
func protoTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}
