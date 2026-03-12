package outbox_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	outbox "github.com/getoutbox/outbox-go"
	outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"
	"google.golang.org/protobuf/proto"
)

func TestParseID(t *testing.T) {
	tests := []struct{ in, want string }{
		{"connectors/abc", "abc"},
		{"accounts/xyz", "xyz"},
		{"a/b/c", "c"},
		{"abc", "abc"},
		{"", ""},
	}
	for _, tt := range tests {
		got := outbox.ParseID(tt.in)
		if got != tt.want {
			t.Errorf("ParseID(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTextPart(t *testing.T) {
	p := outbox.TextPart("hello")
	if p.ContentType != "text/plain" {
		t.Errorf("ContentType = %q, want text/plain", p.ContentType)
	}
	if string(p.Content) != "hello" {
		t.Errorf("Content = %q, want hello", p.Content)
	}
}

func TestTextContent(t *testing.T) {
	p := outbox.MessagePart{Content: []byte("world")}
	got, err := outbox.TextContent(p)
	if err != nil {
		t.Fatal(err)
	}
	if got != "world" {
		t.Errorf("got %q, want world", got)
	}

	_, err = outbox.TextContent(outbox.MessagePart{})
	if err == nil {
		t.Error("expected error for nil content, got nil")
	}
}

func TestVerify(t *testing.T) {
	body := []byte(`{"delivery_id":"abc"}`)
	secret := "mysecret"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	ok, err := outbox.Verify(body, secret, sig)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected valid signature")
	}

	// Wrong signature: flip first two bytes.
	wrongSig := "00" + sig[2:]
	ok, err = outbox.Verify(body, secret, wrongSig)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected invalid signature")
	}

	// Malformed hex.
	_, err = outbox.Verify(body, secret, "notvalidhex!")
	if err == nil {
		t.Error("expected error for invalid hex signature")
	}
}

func TestParseDeliveryEvent_JSON_Message(t *testing.T) {
	// protojson oneof fields are inlined at the top level (no "event" wrapper).
	// "content" bytes are base64-encoded; "aGVsbG8=" = "hello".
	body := []byte(`{
		"connector": "connectors/abc",
		"message": {
			"name": "messages/xyz",
			"direction": 1,
			"parts": [{"contentType": "text/plain", "content": "aGVsbG8="}]
		}
	}`)
	event, err := outbox.ParseDeliveryEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	me, ok := event.(*outbox.MessageDeliveryEvent)
	if !ok {
		t.Fatalf("expected *MessageDeliveryEvent, got %T", event)
	}
	if me.ConnectorID != "abc" {
		t.Errorf("ConnectorID = %q, want abc", me.ConnectorID)
	}
	if me.Message.ID != "xyz" {
		t.Errorf("Message.ID = %q, want xyz", me.Message.ID)
	}
	if len(me.Message.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(me.Message.Parts))
	}
	if string(me.Message.Parts[0].Content) != "hello" {
		t.Errorf("Part content = %q, want hello", me.Message.Parts[0].Content)
	}
}

func TestParseDeliveryEvent_JSON_Unknown(t *testing.T) {
	body := []byte(`{"connector": "connectors/xyz"}`)
	event, err := outbox.ParseDeliveryEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	unk, ok := event.(*outbox.UnknownDeliveryEvent)
	if !ok {
		t.Fatalf("expected *UnknownDeliveryEvent, got %T", event)
	}
	if unk.ConnectorID != "xyz" {
		t.Errorf("ConnectorID = %q, want xyz", unk.ConnectorID)
	}
}

func TestParseDeliveryEvent_InvalidJSON(t *testing.T) {
	_, err := outbox.ParseDeliveryEvent([]byte(`{not valid json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseDeliveryEvent_JSON_EnvelopeFields(t *testing.T) {
	body := []byte(`{
		"connector": "connectors/conn1",
		"deliveryId": "del1",
		"destination": "destinations/dest1",
		"message": {"name": "messages/m1"}
	}`)
	event, err := outbox.ParseDeliveryEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	me, ok := event.(*outbox.MessageDeliveryEvent)
	if !ok {
		t.Fatalf("expected *MessageDeliveryEvent, got %T", event)
	}
	if me.ConnectorID != "conn1" {
		t.Errorf("ConnectorID = %q, want conn1", me.ConnectorID)
	}
	if me.DeliveryID != "del1" {
		t.Errorf("DeliveryID = %q, want del1", me.DeliveryID)
	}
	if me.DestinationID != "dest1" {
		t.Errorf("DestinationID = %q, want dest1", me.DestinationID)
	}
}

func TestParseDeliveryEvent_JSON_DeliveryUpdate(t *testing.T) {
	body := []byte(`{
		"connector": "connectors/conn1",
		"deliveryUpdate": {
			"message": "messages/msg1",
			"status": 2
		}
	}`)
	event, err := outbox.ParseDeliveryEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	due, ok := event.(*outbox.DeliveryUpdateDeliveryEvent)
	if !ok {
		t.Fatalf("expected *DeliveryUpdateDeliveryEvent, got %T", event)
	}
	if due.ConnectorID != "conn1" {
		t.Errorf("ConnectorID = %q, want conn1", due.ConnectorID)
	}
	if due.Delivery.MessageID != "msg1" {
		t.Errorf("Delivery.MessageID = %q, want msg1", due.Delivery.MessageID)
	}
	if due.Delivery.Status != outbox.MessageDeliveryStatusDelivered {
		t.Errorf("Delivery.Status = %v, want delivered", due.Delivery.Status)
	}
}

func TestParseDeliveryEvent_JSON_ReadReceipt(t *testing.T) {
	body := []byte(`{
		"connector": "connectors/conn1",
		"readReceipt": {
			"account": {"name": "accounts/acc1"},
			"messages": ["messages/m1", "messages/m2"]
		}
	}`)
	event, err := outbox.ParseDeliveryEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	rre, ok := event.(*outbox.ReadReceiptDeliveryEvent)
	if !ok {
		t.Fatalf("expected *ReadReceiptDeliveryEvent, got %T", event)
	}
	if rre.ConnectorID != "conn1" {
		t.Errorf("ConnectorID = %q, want conn1", rre.ConnectorID)
	}
	if rre.ReadReceipt.Account == nil {
		t.Fatal("ReadReceipt.Account is nil")
	}
	if rre.ReadReceipt.Account.ID != "acc1" {
		t.Errorf("Account.ID = %q, want acc1", rre.ReadReceipt.Account.ID)
	}
	if len(rre.ReadReceipt.MessageIDs) != 2 {
		t.Fatalf("MessageIDs len = %d, want 2", len(rre.ReadReceipt.MessageIDs))
	}
	if rre.ReadReceipt.MessageIDs[0] != "m1" {
		t.Errorf("MessageIDs[0] = %q, want m1", rre.ReadReceipt.MessageIDs[0])
	}
	if rre.ReadReceipt.MessageIDs[1] != "m2" {
		t.Errorf("MessageIDs[1] = %q, want m2", rre.ReadReceipt.MessageIDs[1])
	}
}

func TestParseDeliveryEvent_JSON_TypingIndicator(t *testing.T) {
	body := []byte(`{
		"connector": "connectors/conn1",
		"typingIndicator": {
			"account": {"name": "accounts/acc1"},
			"typing": true,
			"contentType": "text/plain"
		}
	}`)
	event, err := outbox.ParseDeliveryEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	tie, ok := event.(*outbox.TypingIndicatorDeliveryEvent)
	if !ok {
		t.Fatalf("expected *TypingIndicatorDeliveryEvent, got %T", event)
	}
	if tie.ConnectorID != "conn1" {
		t.Errorf("ConnectorID = %q, want conn1", tie.ConnectorID)
	}
	if !tie.TypingIndicator.Typing {
		t.Error("Typing = false, want true")
	}
	if tie.TypingIndicator.ContentType != "text/plain" {
		t.Errorf("ContentType = %q, want text/plain", tie.TypingIndicator.ContentType)
	}
}

func TestParseDeliveryEvent_ProtoBinary(t *testing.T) {
	p := &outboxv1.DeliveryEvent{
		DeliveryId:  "proto-del",
		Connector:   "connectors/proto-conn",
		Destination: "destinations/proto-dest",
		Event: &outboxv1.DeliveryEvent_Message{
			Message: &outboxv1.Message{
				Name:      "messages/proto-msg",
				Direction: outboxv1.Message_DIRECTION_INBOUND,
			},
		},
	}
	body, err := proto.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	event, err := outbox.ParseDeliveryEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	me, ok := event.(*outbox.MessageDeliveryEvent)
	if !ok {
		t.Fatalf("expected *MessageDeliveryEvent, got %T", event)
	}
	if me.ConnectorID != "proto-conn" {
		t.Errorf("ConnectorID = %q, want proto-conn", me.ConnectorID)
	}
	if me.DeliveryID != "proto-del" {
		t.Errorf("DeliveryID = %q, want proto-del", me.DeliveryID)
	}
	if me.DestinationID != "proto-dest" {
		t.Errorf("DestinationID = %q, want proto-dest", me.DestinationID)
	}
	if me.Message.ID != "proto-msg" {
		t.Errorf("Message.ID = %q, want proto-msg", me.Message.ID)
	}
	if me.Message.Direction != outbox.MessageDirectionInbound {
		t.Errorf("Message.Direction = %v, want inbound", me.Message.Direction)
	}
}

func TestParseDeliveryEvent_JSON_DeliveryUpdate_WithErrors(t *testing.T) {
	body := []byte(`{
		"connector": "connectors/conn1",
		"deliveryUpdate": {
			"message": "messages/msg1",
			"status": 3,
			"errorCode": "RATE_LIMITED",
			"errorMessage": "Too many requests"
		}
	}`)
	event, err := outbox.ParseDeliveryEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	due, ok := event.(*outbox.DeliveryUpdateDeliveryEvent)
	if !ok {
		t.Fatalf("expected *DeliveryUpdateDeliveryEvent, got %T", event)
	}
	if due.Delivery.ErrorCode == nil {
		t.Fatal("ErrorCode is nil, want RATE_LIMITED")
	}
	if *due.Delivery.ErrorCode != "RATE_LIMITED" {
		t.Errorf("ErrorCode = %q, want RATE_LIMITED", *due.Delivery.ErrorCode)
	}
	if due.Delivery.ErrorMessage == nil {
		t.Fatal("ErrorMessage is nil, want 'Too many requests'")
	}
	if *due.Delivery.ErrorMessage != "Too many requests" {
		t.Errorf("ErrorMessage = %q, want 'Too many requests'", *due.Delivery.ErrorMessage)
	}
}

func TestParseID_TrailingSlash(t *testing.T) {
	got := outbox.ParseID("connectors/")
	if got != "" {
		t.Errorf("ParseID(%q) = %q, want empty string", "connectors/", got)
	}
}

func TestTextContent_EmptySlice(t *testing.T) {
	p := outbox.MessagePart{Content: []byte{}}
	got, err := outbox.TextContent(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestTextPartRoundtrip(t *testing.T) {
	part := outbox.TextPart("hello world")
	got, err := outbox.TextContent(part)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello world" {
		t.Errorf("got %q, want 'hello world'", got)
	}
}

func TestVerify_EmptyBodyEmptySecret(t *testing.T) {
	// Verify that an empty body and empty secret can still produce a valid HMAC.
	// The expected sig is HMAC-SHA256 of "" with key "".
	import_crypto_hmac := func() string {
		mac := hmac.New(sha256.New, []byte(""))
		mac.Write([]byte{})
		return hex.EncodeToString(mac.Sum(nil))
	}
	sig := import_crypto_hmac()
	ok, err := outbox.Verify([]byte{}, "", sig)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected valid signature for empty body/secret")
	}
}
