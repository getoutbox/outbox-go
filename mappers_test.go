package outbox

import (
	"testing"
	"time"

	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMapAccount_AllFields(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	p := &outboxv1.Account{
		Name:       "accounts/acc123",
		ContactId:  "contact1",
		ExternalId: "ext1",
		Source:     outboxv1.Account_SOURCE_API,
		Metadata:   map[string]string{"k": "v"},
		CreateTime: timestamppb.New(ts),
		UpdateTime: timestamppb.New(ts.Add(time.Hour)),
	}
	got := mapAccount(p)
	if got.ID != "acc123" {
		t.Errorf("ID = %q, want acc123", got.ID)
	}
	if got.ContactID != "contact1" {
		t.Errorf("ContactID = %q, want contact1", got.ContactID)
	}
	if got.ExternalID != "ext1" {
		t.Errorf("ExternalID = %q, want ext1", got.ExternalID)
	}
	if got.Source != AccountSourceAPI {
		t.Errorf("Source = %v, want AccountSourceAPI", got.Source)
	}
	if got.Metadata["k"] != "v" {
		t.Errorf("Metadata[k] = %q, want v", got.Metadata["k"])
	}
	if !got.CreateTime.Equal(ts) {
		t.Errorf("CreateTime = %v, want %v", got.CreateTime, ts)
	}
	if !got.UpdateTime.Equal(ts.Add(time.Hour)) {
		t.Errorf("UpdateTime = %v, want %v", got.UpdateTime, ts.Add(time.Hour))
	}
}

func TestMapAccount_NilMetadata(t *testing.T) {
	p := &outboxv1.Account{Name: "accounts/a1"}
	got := mapAccount(p)
	if got.Metadata != nil {
		t.Errorf("Metadata = %v, want nil", got.Metadata)
	}
}

func TestMapChannel_WithCapabilities(t *testing.T) {
	p := &outboxv1.Channel{
		Name: "channels/ch1",
		Capabilities: &outboxv1.Channel_Capabilities{
			Groups:                true,
			Reactions:             true,
			Edits:                 false,
			Deletions:             true,
			ReadReceipts:          false,
			TypingIndicators:      true,
			SupportedContentTypes: []string{"text/plain", "image/jpeg"},
		},
	}
	got := mapChannel(p)
	if got.ID != "ch1" {
		t.Errorf("ID = %q, want ch1", got.ID)
	}
	if got.Capabilities == nil {
		t.Fatal("Capabilities is nil")
	}
	if !got.Capabilities.Groups {
		t.Error("Groups should be true")
	}
	if !got.Capabilities.Reactions {
		t.Error("Reactions should be true")
	}
	if got.Capabilities.Edits {
		t.Error("Edits should be false")
	}
	if !got.Capabilities.Deletions {
		t.Error("Deletions should be true")
	}
	if got.Capabilities.ReadReceipts {
		t.Error("ReadReceipts should be false")
	}
	if !got.Capabilities.TypingIndicators {
		t.Error("TypingIndicators should be true")
	}
	if len(got.Capabilities.SupportedContentTypes) != 2 {
		t.Errorf("SupportedContentTypes len = %d, want 2", len(got.Capabilities.SupportedContentTypes))
	}
}

func TestMapChannel_NilCapabilities(t *testing.T) {
	p := &outboxv1.Channel{Name: "channels/ch1"}
	got := mapChannel(p)
	if got.Capabilities != nil {
		t.Errorf("Capabilities = %v, want nil", got.Capabilities)
	}
}

func TestMapConnector_AllFields(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	p := &outboxv1.Connector{
		Name:         "connectors/conn1",
		State:        outboxv1.Connector_STATE_ACTIVE,
		Tags:         []string{"prod", "slack"},
		ErrorMessage: "something went wrong",
		Account: &outboxv1.Account{
			Name:      "accounts/acc1",
			ContactId: "contact1",
		},
		CreateTime: timestamppb.New(ts),
		UpdateTime: timestamppb.New(ts.Add(time.Hour)),
	}
	got := mapConnector(p)
	if got.ID != "conn1" {
		t.Errorf("ID = %q, want conn1", got.ID)
	}
	if got.State != ConnectorStateActive {
		t.Errorf("State = %v, want ConnectorStateActive", got.State)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "prod" || got.Tags[1] != "slack" {
		t.Errorf("Tags = %v, want [prod slack]", got.Tags)
	}
	if got.ErrorMessage != "something went wrong" {
		t.Errorf("ErrorMessage = %q, want 'something went wrong'", got.ErrorMessage)
	}
	if got.Account == nil {
		t.Fatal("Account is nil")
	}
	if got.Account.ID != "acc1" {
		t.Errorf("Account.ID = %q, want acc1", got.Account.ID)
	}
	if !got.CreateTime.Equal(ts) {
		t.Errorf("CreateTime = %v, want %v", got.CreateTime, ts)
	}
}

func TestMapConnector_NilAccount(t *testing.T) {
	p := &outboxv1.Connector{
		Name:  "connectors/conn1",
		State: outboxv1.Connector_STATE_AUTHORIZING,
	}
	got := mapConnector(p)
	if got.Account != nil {
		t.Errorf("Account = %v, want nil", got.Account)
	}
	if got.State != ConnectorStateAuthorizing {
		t.Errorf("State = %v, want ConnectorStateAuthorizing", got.State)
	}
}

func TestMapMessage_BasicFields(t *testing.T) {
	p := &outboxv1.Message{
		Name:      "messages/msg1",
		Recipient: "accounts/acc1",
		Direction: outboxv1.Message_DIRECTION_OUTBOUND,
		Parts: []*outboxv1.MessagePart{
			{
				ContentType: "text/plain",
				Source:      &outboxv1.MessagePart_Content{Content: []byte("hello")},
			},
		},
		Metadata: map[string]string{"key": "val"},
	}
	got := mapMessage(p)
	if got.ID != "msg1" {
		t.Errorf("ID = %q, want msg1", got.ID)
	}
	if got.RecipientID != "acc1" {
		t.Errorf("RecipientID = %q, want acc1", got.RecipientID)
	}
	if got.Direction != MessageDirectionOutbound {
		t.Errorf("Direction = %v, want outbound", got.Direction)
	}
	if len(got.Parts) != 1 {
		t.Fatalf("Parts len = %d, want 1", len(got.Parts))
	}
	if string(got.Parts[0].Content) != "hello" {
		t.Errorf("Parts[0].Content = %q, want hello", got.Parts[0].Content)
	}
	if got.Metadata["key"] != "val" {
		t.Errorf("Metadata[key] = %q, want val", got.Metadata["key"])
	}
}

func TestMapMessage_OptionalFields(t *testing.T) {
	replaced := "messages/old1"
	p := &outboxv1.Message{
		Name:     "messages/msg1",
		ReplyTo:  "messages/reply1",
		GroupId:  "group1",
		Replaced: &replaced,
	}
	got := mapMessage(p)
	if got.ReplyToMessageID != "reply1" {
		t.Errorf("ReplyToMessageID = %q, want reply1", got.ReplyToMessageID)
	}
	if got.GroupID != "group1" {
		t.Errorf("GroupID = %q, want group1", got.GroupID)
	}
	if got.ReplacedMessageID == nil || *got.ReplacedMessageID != "old1" {
		t.Errorf("ReplacedMessageID = %v, want pointer to old1", got.ReplacedMessageID)
	}
}

func TestMapMessage_ReplacedMessageID_EmptyStringPointer(t *testing.T) {
	// Edge case: proto Replaced is a non-nil *string pointing to "" (empty string).
	// ParseID("") returns "", so ReplacedMessageID should be a non-nil *string pointing to "".
	empty := ""
	p := &outboxv1.Message{Name: "messages/msg1", Replaced: &empty}
	got := mapMessage(p)
	if got.ReplacedMessageID == nil {
		t.Fatal("ReplacedMessageID is nil, want non-nil pointer to empty string")
	}
	if *got.ReplacedMessageID != "" {
		t.Errorf("*ReplacedMessageID = %q, want empty string", *got.ReplacedMessageID)
	}
}

func TestMapMessage_EmptyOptionalFields(t *testing.T) {
	p := &outboxv1.Message{Name: "messages/msg1"}
	got := mapMessage(p)
	if got.ReplyToMessageID != "" {
		t.Errorf("ReplyToMessageID = %q, want empty", got.ReplyToMessageID)
	}
	if got.GroupID != "" {
		t.Errorf("GroupID = %q, want empty", got.GroupID)
	}
	if got.ReplacedMessageID != nil {
		t.Errorf("ReplacedMessageID = %v, want nil", got.ReplacedMessageID)
	}
}

func TestMapMessagePart_ContentSource(t *testing.T) {
	content := []byte{1, 2, 3}
	p := &outboxv1.MessagePart{
		ContentType: "application/octet-stream",
		Disposition: outboxv1.MessagePart_DISPOSITION_ATTACHMENT,
		Filename:    "file.bin",
		Source:      &outboxv1.MessagePart_Content{Content: content},
	}
	got := mapMessagePart(p)
	if got.ContentType != "application/octet-stream" {
		t.Errorf("ContentType = %q, want application/octet-stream", got.ContentType)
	}
	if got.Disposition != MessagePartDispositionAttachment {
		t.Errorf("Disposition = %v, want ATTACHMENT", got.Disposition)
	}
	if got.Filename != "file.bin" {
		t.Errorf("Filename = %q, want file.bin", got.Filename)
	}
	if string(got.Content) != string(content) {
		t.Errorf("Content = %v, want %v", got.Content, content)
	}
	// Verify deep copy
	content[0] = 99
	if got.Content[0] == 99 {
		t.Error("Content was not deep-copied")
	}
	if got.URL != "" {
		t.Errorf("URL = %q, want empty", got.URL)
	}
}

func TestMapMessagePart_URLSource(t *testing.T) {
	p := &outboxv1.MessagePart{
		ContentType: "image/jpeg",
		Source:      &outboxv1.MessagePart_Url{Url: "https://example.com/img.jpg"},
	}
	got := mapMessagePart(p)
	if got.URL != "https://example.com/img.jpg" {
		t.Errorf("URL = %q, want https://example.com/img.jpg", got.URL)
	}
	if got.Content != nil {
		t.Errorf("Content = %v, want nil", got.Content)
	}
}

func TestMapMessagePart_NoSource(t *testing.T) {
	p := &outboxv1.MessagePart{ContentType: "text/plain"}
	got := mapMessagePart(p)
	if got.Content != nil {
		t.Errorf("Content = %v, want nil", got.Content)
	}
	if got.URL != "" {
		t.Errorf("URL = %q, want empty", got.URL)
	}
}

func TestToProtoParts_Nil(t *testing.T) {
	got := toProtoParts(nil)
	if got != nil {
		t.Errorf("toProtoParts(nil) = %v, want nil", got)
	}
}

func TestToProtoParts_Empty(t *testing.T) {
	got := toProtoParts([]MessagePart{})
	if got == nil {
		t.Error("toProtoParts([]) = nil, want empty slice")
	}
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

func TestToProtoParts_ContentPart(t *testing.T) {
	parts := []MessagePart{
		{
			ContentType: "text/plain",
			Content:     []byte("hello"),
			Filename:    "note.txt",
			Disposition: MessagePartDispositionRender,
		},
	}
	got := toProtoParts(parts)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	p := got[0]
	if p.ContentType != "text/plain" {
		t.Errorf("ContentType = %q, want text/plain", p.ContentType)
	}
	if p.Filename != "note.txt" {
		t.Errorf("Filename = %q, want note.txt", p.Filename)
	}
	src, ok := p.Source.(*outboxv1.MessagePart_Content)
	if !ok {
		t.Fatalf("Source type = %T, want *outboxv1.MessagePart_Content", p.Source)
	}
	if string(src.Content) != "hello" {
		t.Errorf("Content = %q, want hello", src.Content)
	}
}

func TestToProtoParts_URLPart(t *testing.T) {
	parts := []MessagePart{
		{
			ContentType: "image/jpeg",
			URL:         "https://example.com/img.jpg",
		},
	}
	got := toProtoParts(parts)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	src, ok := got[0].Source.(*outboxv1.MessagePart_Url)
	if !ok {
		t.Fatalf("Source type = %T, want *outboxv1.MessagePart_Url", got[0].Source)
	}
	if src.Url != "https://example.com/img.jpg" {
		t.Errorf("URL = %q, want https://example.com/img.jpg", src.Url)
	}
}

func TestToProtoParts_URLTakesPriority(t *testing.T) {
	// When both URL and Content are set, URL takes priority.
	parts := []MessagePart{
		{
			ContentType: "image/jpeg",
			URL:         "https://example.com/img.jpg",
			Content:     []byte("ignored"),
		},
	}
	got := toProtoParts(parts)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if _, ok := got[0].Source.(*outboxv1.MessagePart_Url); !ok {
		t.Errorf("Source type = %T, want *outboxv1.MessagePart_Url", got[0].Source)
	}
}

func TestToProtoParts_DeepCopy(t *testing.T) {
	original := []byte{1, 2, 3}
	parts := []MessagePart{{Content: original}}
	got := toProtoParts(parts)
	src := got[0].Source.(*outboxv1.MessagePart_Content)
	original[0] = 99
	if src.Content[0] == 99 {
		t.Error("toProtoParts did not deep-copy Content")
	}
}

func TestMapMessageDelivery(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	errCode := "ERR_TIMEOUT"
	errMsg := "delivery timed out"
	p := &outboxv1.MessageDelivery{
		Message:          "messages/msg1",
		Status:           outboxv1.MessageDelivery_STATUS_FAILED,
		StatusChangeTime: timestamppb.New(ts),
		Account:          &outboxv1.Account{Name: "accounts/acc1"},
		ErrorCode:        &errCode,
		ErrorMessage:     &errMsg,
	}
	got := mapMessageDelivery(p)
	if got.MessageID != "msg1" {
		t.Errorf("MessageID = %q, want msg1", got.MessageID)
	}
	if got.Status != MessageDeliveryStatusFailed {
		t.Errorf("Status = %v, want failed", got.Status)
	}
	if !got.StatusChangeTime.Equal(ts) {
		t.Errorf("StatusChangeTime = %v, want %v", got.StatusChangeTime, ts)
	}
	if got.Account == nil {
		t.Fatal("Account is nil")
	}
	if got.Account.ID != "acc1" {
		t.Errorf("Account.ID = %q, want acc1", got.Account.ID)
	}
	if got.ErrorCode == nil || *got.ErrorCode != errCode {
		t.Errorf("ErrorCode = %v, want %q", got.ErrorCode, errCode)
	}
	if got.ErrorMessage == nil || *got.ErrorMessage != errMsg {
		t.Errorf("ErrorMessage = %v, want %q", got.ErrorMessage, errMsg)
	}
}

func TestMapMessageDelivery_NoError(t *testing.T) {
	p := &outboxv1.MessageDelivery{
		Message: "messages/msg1",
		Status:  outboxv1.MessageDelivery_STATUS_DELIVERED,
	}
	got := mapMessageDelivery(p)
	if got.ErrorCode != nil {
		t.Errorf("ErrorCode = %v, want nil", got.ErrorCode)
	}
	if got.ErrorMessage != nil {
		t.Errorf("ErrorMessage = %v, want nil", got.ErrorMessage)
	}
}

func TestMapReadReceipt(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	p := &outboxv1.ReadReceiptEvent{
		Account:   &outboxv1.Account{Name: "accounts/acc1"},
		Messages:  []string{"messages/msg1", "messages/msg2"},
		Timestamp: timestamppb.New(ts),
	}
	got := mapReadReceipt(p)
	if got.Account == nil {
		t.Fatal("Account is nil")
	}
	if got.Account.ID != "acc1" {
		t.Errorf("Account.ID = %q, want acc1", got.Account.ID)
	}
	if len(got.MessageIDs) != 2 {
		t.Fatalf("MessageIDs len = %d, want 2", len(got.MessageIDs))
	}
	if got.MessageIDs[0] != "msg1" {
		t.Errorf("MessageIDs[0] = %q, want msg1", got.MessageIDs[0])
	}
	if got.MessageIDs[1] != "msg2" {
		t.Errorf("MessageIDs[1] = %q, want msg2", got.MessageIDs[1])
	}
	if !got.Timestamp.Equal(ts) {
		t.Errorf("Timestamp = %v, want %v", got.Timestamp, ts)
	}
}

func TestMapTypingIndicator(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	p := &outboxv1.TypingIndicatorEvent{
		Account:     &outboxv1.Account{Name: "accounts/acc1"},
		Typing:      true,
		ContentType: "text/plain",
		Timestamp:   timestamppb.New(ts),
	}
	got := mapTypingIndicator(p)
	if got.Account == nil {
		t.Fatal("Account is nil")
	}
	if got.Account.ID != "acc1" {
		t.Errorf("Account.ID = %q, want acc1", got.Account.ID)
	}
	if !got.Typing {
		t.Error("Typing = false, want true")
	}
	if got.ContentType != "text/plain" {
		t.Errorf("ContentType = %q, want text/plain", got.ContentType)
	}
	if !got.Timestamp.Equal(ts) {
		t.Errorf("Timestamp = %v, want %v", got.Timestamp, ts)
	}
}

func TestProtoTime_Nil(t *testing.T) {
	got := protoTime(nil)
	if !got.IsZero() {
		t.Errorf("protoTime(nil) = %v, want zero time", got)
	}
}

func TestCloneStringMap_Nil(t *testing.T) {
	got := cloneStringMap(nil)
	if got != nil {
		t.Errorf("cloneStringMap(nil) = %v, want nil", got)
	}
}

func TestCloneStringMap_DeepCopy(t *testing.T) {
	original := map[string]string{"k": "v"}
	got := cloneStringMap(original)
	original["k"] = "changed"
	if got["k"] != "v" {
		t.Errorf("cloneStringMap not deep copy: got %q, want v", got["k"])
	}
}

func TestMapAccountPtr_Nil(t *testing.T) {
	got := mapAccountPtr(nil)
	if got != nil {
		t.Errorf("mapAccountPtr(nil) = %v, want nil", got)
	}
}

func TestMapAccountPtr_NonNil(t *testing.T) {
	p := &outboxv1.Account{Name: "accounts/acc1", ContactId: "c1"}
	got := mapAccountPtr(p)
	if got == nil {
		t.Fatal("mapAccountPtr returned nil for non-nil input")
	}
	if got.ID != "acc1" {
		t.Errorf("ID = %q, want acc1", got.ID)
	}
	if got.ContactID != "c1" {
		t.Errorf("ContactID = %q, want c1", got.ContactID)
	}
}

func TestMapDestination_AllFields(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	target := &outboxv1.Destination_Webhook{
		Webhook: &outboxv1.WebhookTarget{Url: "https://example.com"},
	}
	p := &outboxv1.Destination{
		Name:          "destinations/dest1",
		DisplayName:   "My Destination",
		State:         outboxv1.Destination_STATE_ACTIVE,
		EventTypes:    []outboxv1.Destination_EventType{outboxv1.Destination_EVENT_TYPE_MESSAGE},
		Filter:        "direction == INBOUND",
		PayloadFormat: outboxv1.Destination_PAYLOAD_FORMAT_JSON,
		Target:        target,
		CreateTime:    timestamppb.New(ts),
		UpdateTime:    timestamppb.New(ts.Add(time.Hour)),
	}
	got := mapDestination(p)
	if got.ID != "dest1" {
		t.Errorf("ID = %q, want dest1", got.ID)
	}
	if got.DisplayName != "My Destination" {
		t.Errorf("DisplayName = %q, want 'My Destination'", got.DisplayName)
	}
	if got.State != DestinationStateActive {
		t.Errorf("State = %v, want DestinationStateActive", got.State)
	}
	if len(got.EventTypes) != 1 || got.EventTypes[0] != DestinationEventTypeMessage {
		t.Errorf("EventTypes = %v, want [message]", got.EventTypes)
	}
	if got.Filter != "direction == INBOUND" {
		t.Errorf("Filter = %q, want 'direction == INBOUND'", got.Filter)
	}
	if got.PayloadFormat != DestinationPayloadFormatJSON {
		t.Errorf("PayloadFormat = %v, want JSON", got.PayloadFormat)
	}
	if got.Target != target {
		t.Errorf("Target = %v, want %v", got.Target, target)
	}
	if !got.CreateTime.Equal(ts) {
		t.Errorf("CreateTime = %v, want %v", got.CreateTime, ts)
	}
	if !got.UpdateTime.Equal(ts.Add(time.Hour)) {
		t.Errorf("UpdateTime = %v, want %v", got.UpdateTime, ts.Add(time.Hour))
	}
}

func TestMapDestination_NilTarget(t *testing.T) {
	p := &outboxv1.Destination{Name: "destinations/dest1"}
	got := mapDestination(p)
	if got.Target != nil {
		t.Errorf("Target = %v, want nil", got.Target)
	}
}

func TestApplyConnectorChannelConfig(t *testing.T) {
	tests := []struct {
		name      string
		cfg       any
		wantField string
		wantMut   bool // expect dst to be mutated
	}{
		{
			name:      "nil returns empty",
			cfg:       nil,
			wantField: "",
			wantMut:   false,
		},
		{
			name:      "wrong type returns empty",
			cfg:       "not-a-oneof",
			wantField: "",
			wantMut:   false,
		},
		{
			name:      "Connector_Email returns email",
			cfg:       &outboxv1.Connector_Email{Email: &outboxv1.EmailConfig{}},
			wantField: "email",
			wantMut:   true,
		},
		{
			name:      "Connector_Discord returns discord",
			cfg:       &outboxv1.Connector_Discord{Discord: &outboxv1.DiscordConfig{}},
			wantField: "discord",
			wantMut:   true,
		},
		{
			name:      "Connector_Slack returns slack",
			cfg:       &outboxv1.Connector_Slack{Slack: &outboxv1.SlackConfig{}},
			wantField: "slack",
			wantMut:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &outboxv1.Connector{}
			got := applyConnectorChannelConfig(dst, tt.cfg)
			if got != tt.wantField {
				t.Errorf("field = %q, want %q", got, tt.wantField)
			}
			if tt.wantMut && dst.ChannelConfig == nil {
				t.Error("dst.ChannelConfig is nil, want mutated")
			}
			if !tt.wantMut && dst.ChannelConfig != nil {
				t.Errorf("dst.ChannelConfig = %v, want nil (no mutation)", dst.ChannelConfig)
			}
		})
	}
}

func TestApplyDestinationTarget(t *testing.T) {
	tests := []struct {
		name      string
		target    any
		wantField string
		wantMut   bool
	}{
		{
			name:      "nil returns empty",
			target:    nil,
			wantField: "",
			wantMut:   false,
		},
		{
			name:      "wrong type returns empty",
			target:    42,
			wantField: "",
			wantMut:   false,
		},
		{
			name:      "Destination_Webhook returns webhook",
			target:    &outboxv1.Destination_Webhook{Webhook: &outboxv1.WebhookTarget{Url: "https://example.com"}},
			wantField: "webhook",
			wantMut:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := &outboxv1.Destination{}
			got := applyDestinationTarget(dst, tt.target)
			if got != tt.wantField {
				t.Errorf("field = %q, want %q", got, tt.wantField)
			}
			if tt.wantMut && dst.Target == nil {
				t.Error("dst.Target is nil, want mutated")
			}
			if !tt.wantMut && dst.Target != nil {
				t.Errorf("dst.Target = %v, want nil (no mutation)", dst.Target)
			}
		})
	}
}

func TestMapMessage_WithAccount(t *testing.T) {
	p := &outboxv1.Message{
		Name:    "messages/msg1",
		Account: &outboxv1.Account{Name: "accounts/acc1", ContactId: "c1"},
	}
	got := mapMessage(p)
	if got.Account == nil {
		t.Fatal("Account is nil")
	}
	if got.Account.ID != "acc1" {
		t.Errorf("Account.ID = %q, want acc1", got.Account.ID)
	}
	if got.Account.ContactID != "c1" {
		t.Errorf("Account.ContactID = %q, want c1", got.Account.ContactID)
	}
}

func TestMapConnector_EmptyTags(t *testing.T) {
	p := &outboxv1.Connector{Name: "connectors/c1", Tags: []string{}}
	got := mapConnector(p)
	if got.Tags == nil {
		t.Error("Tags is nil, want empty slice")
	}
	if len(got.Tags) != 0 {
		t.Errorf("Tags len = %d, want 0", len(got.Tags))
	}
}

func TestMapReadReceipt_EmptyMessages(t *testing.T) {
	p := &outboxv1.ReadReceiptEvent{
		Messages: []string{},
		Account:  &outboxv1.Account{Name: "accounts/acc1"},
	}
	got := mapReadReceipt(p)
	if got.MessageIDs == nil {
		t.Error("MessageIDs is nil, want empty slice")
	}
	if len(got.MessageIDs) != 0 {
		t.Errorf("MessageIDs len = %d, want 0", len(got.MessageIDs))
	}
}

func TestMapTypingIndicator_NilAccount(t *testing.T) {
	p := &outboxv1.TypingIndicatorEvent{Account: nil, Typing: true}
	got := mapTypingIndicator(p)
	if got.Account != nil {
		t.Errorf("Account = %v, want nil", got.Account)
	}
	if !got.Typing {
		t.Error("Typing = false, want true")
	}
}

func TestMapChannel_EmptySupportedContentTypes(t *testing.T) {
	p := &outboxv1.Channel{
		Name: "channels/ch1",
		Capabilities: &outboxv1.Channel_Capabilities{
			SupportedContentTypes: []string{},
		},
	}
	got := mapChannel(p)
	if got.Capabilities == nil {
		t.Fatal("Capabilities is nil")
	}
	if got.Capabilities.SupportedContentTypes == nil {
		t.Error("SupportedContentTypes is nil, want empty slice")
	}
	if len(got.Capabilities.SupportedContentTypes) != 0 {
		t.Errorf("SupportedContentTypes len = %d, want 0", len(got.Capabilities.SupportedContentTypes))
	}
}

func TestMapDeliveryEvent_Message(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	p := &outboxv1.DeliveryEvent{
		DeliveryId:  "del1",
		Connector:   "connectors/conn1",
		Destination: "destinations/dest1",
		EnqueueTime: timestamppb.New(ts),
		Event: &outboxv1.DeliveryEvent_Message{
			Message: &outboxv1.Message{Name: "messages/msg1"},
		},
	}
	got := mapDeliveryEvent(p)
	me, ok := got.(*MessageDeliveryEvent)
	if !ok {
		t.Fatalf("got %T, want *MessageDeliveryEvent", got)
	}
	if me.DeliveryID != "del1" {
		t.Errorf("DeliveryID = %q, want del1", me.DeliveryID)
	}
	if me.ConnectorID != "conn1" {
		t.Errorf("ConnectorID = %q, want conn1", me.ConnectorID)
	}
	if me.DestinationID != "dest1" {
		t.Errorf("DestinationID = %q, want dest1", me.DestinationID)
	}
	if !me.EnqueueTime.Equal(ts) {
		t.Errorf("EnqueueTime = %v, want %v", me.EnqueueTime, ts)
	}
	if me.Message.ID != "msg1" {
		t.Errorf("Message.ID = %q, want msg1", me.Message.ID)
	}
}

func TestMapDeliveryEvent_DeliveryUpdate(t *testing.T) {
	p := &outboxv1.DeliveryEvent{
		Connector: "connectors/conn1",
		Event: &outboxv1.DeliveryEvent_DeliveryUpdate{
			DeliveryUpdate: &outboxv1.MessageDelivery{
				Message: "messages/msg1",
				Status:  outboxv1.MessageDelivery_STATUS_DELIVERED,
			},
		},
	}
	got := mapDeliveryEvent(p)
	due, ok := got.(*DeliveryUpdateDeliveryEvent)
	if !ok {
		t.Fatalf("got %T, want *DeliveryUpdateDeliveryEvent", got)
	}
	if due.Delivery.MessageID != "msg1" {
		t.Errorf("Delivery.MessageID = %q, want msg1", due.Delivery.MessageID)
	}
	if due.Delivery.Status != MessageDeliveryStatusDelivered {
		t.Errorf("Delivery.Status = %v, want delivered", due.Delivery.Status)
	}
}

func TestMapDeliveryEvent_ReadReceipt(t *testing.T) {
	p := &outboxv1.DeliveryEvent{
		Connector: "connectors/conn1",
		Event: &outboxv1.DeliveryEvent_ReadReceipt{
			ReadReceipt: &outboxv1.ReadReceiptEvent{
				Account:  &outboxv1.Account{Name: "accounts/acc1"},
				Messages: []string{"messages/m1"},
			},
		},
	}
	got := mapDeliveryEvent(p)
	rre, ok := got.(*ReadReceiptDeliveryEvent)
	if !ok {
		t.Fatalf("got %T, want *ReadReceiptDeliveryEvent", got)
	}
	if rre.ReadReceipt.Account == nil || rre.ReadReceipt.Account.ID != "acc1" {
		t.Errorf("ReadReceipt.Account.ID = %v, want acc1", rre.ReadReceipt.Account)
	}
	if len(rre.ReadReceipt.MessageIDs) != 1 || rre.ReadReceipt.MessageIDs[0] != "m1" {
		t.Errorf("MessageIDs = %v, want [m1]", rre.ReadReceipt.MessageIDs)
	}
}

func TestMapDeliveryEvent_TypingIndicator(t *testing.T) {
	p := &outboxv1.DeliveryEvent{
		Connector: "connectors/conn1",
		Event: &outboxv1.DeliveryEvent_TypingIndicator{
			TypingIndicator: &outboxv1.TypingIndicatorEvent{Typing: true},
		},
	}
	got := mapDeliveryEvent(p)
	tie, ok := got.(*TypingIndicatorDeliveryEvent)
	if !ok {
		t.Fatalf("got %T, want *TypingIndicatorDeliveryEvent", got)
	}
	if !tie.TypingIndicator.Typing {
		t.Error("Typing = false, want true")
	}
}

func TestMapDeliveryEvent_Unknown(t *testing.T) {
	p := &outboxv1.DeliveryEvent{
		DeliveryId: "del1",
		Connector:  "connectors/conn1",
	}
	got := mapDeliveryEvent(p)
	unk, ok := got.(*UnknownDeliveryEvent)
	if !ok {
		t.Fatalf("got %T, want *UnknownDeliveryEvent", got)
	}
	if unk.ConnectorID != "conn1" {
		t.Errorf("ConnectorID = %q, want conn1", unk.ConnectorID)
	}
}
