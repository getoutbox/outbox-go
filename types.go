package outbox

import (
	"time"
)

// Account represents an end-user account on a Channel.
type Account struct {
	ID         string
	ContactID  string
	ExternalID string
	Source     AccountSource
	Metadata   map[string]string
	CreateTime time.Time
	UpdateTime time.Time
}

// Connector bridges an Account to a Channel.
// ChannelConfig holds the channel-specific configuration; its concrete type is
// one of the Connector_* wrapper types defined in the outboxv1 package (e.g.
// *outboxv1.Connector_SlackBot). Use a type switch to access channel details.
type Connector struct {
	ID                   string
	Kind                 ConnectorKind
	State                ConnectorState
	Readiness            ConnectorReadiness
	ProvisionedResources []string
	WebhookURL           string
	DisplayName          string
	Tags                 []string
	ChannelConfig        any // outboxv1.isConnector_ChannelConfig (unexported interface)
	// ErrorMessage contains a human-readable error detail when State is ERROR.
	ErrorMessage string
	CreateTime   time.Time
	UpdateTime   time.Time
}

// ReauthorizeResult is returned by Connectors.Reauthorize.
type ReauthorizeResult struct {
	Connector *Connector
	// AuthorizationURL is the OAuth URL to redirect the user to, if applicable.
	AuthorizationURL string
}

// CreateConnectorResult is returned by Connectors.Create.
type CreateConnectorResult struct {
	Connector *Connector
	// AuthorizationURL is set when the channel requires OAuth authorization.
	// Redirect the user to this URL to complete setup.
	// Empty for direct-creation channels.
	AuthorizationURL string
}

// MessagePart is a single content fragment within a Message.
type MessagePart struct {
	ContentType string
	Disposition MessagePartDisposition
	Content     []byte // nil for URL-sourced parts
	URL         string // empty for inline parts
	Filename    string
}

// Message represents an inbound or outbound message on a Connector.
type Message struct {
	ID                string
	Account           *Account
	RecipientID       string
	Parts             []MessagePart
	Metadata          map[string]string
	Direction         MessageDirection
	DeletionScope     MessageDeletionScope
	EditNumber        int64
	ReplyToMessageID  string
	GroupID           string
	ReplacedMessageID *string
	CreateTime        time.Time
	DeliverTime       time.Time
	DeleteTime        time.Time
}

// MessageDelivery records the delivery status of a sent Message.
type MessageDelivery struct {
	MessageID string
	Account   *Account
	Status    MessageDeliveryStatus
	// ErrorCode is nil when no error occurred. Non-nil for failed deliveries.
	ErrorCode *string
	// ErrorMessage is nil when no error occurred. Non-nil for failed deliveries.
	ErrorMessage     *string
	StatusChangeTime time.Time
}

// ReadReceiptEvent indicates that an Account has read one or more messages.
type ReadReceiptEvent struct {
	Account    *Account
	MessageIDs []string
	Timestamp  time.Time
}

// TypingIndicatorEvent signals that an Account has started or stopped typing.
type TypingIndicatorEvent struct {
	Account     *Account
	Typing      bool
	ContentType string
	Timestamp   time.Time
}

// Destination is a push target that receives events from all Connectors.
// Target holds the destination-specific configuration; its concrete type is
// one of the Destination_* wrapper types defined in the outboxv1 package (e.g.
// *outboxv1.Destination_Webhook). Use a type switch to access target details.
type Destination struct {
	ID            string
	DisplayName   string
	State         DestinationState
	EventTypes    []DestinationEventType
	Filter        string
	PayloadFormat DestinationPayloadFormat
	Target        any // outboxv1.isDestination_Target (unexported interface)
	// LastTestTime is the time of the most recent connectivity test, or nil if never tested.
	LastTestTime *time.Time
	// LastTestSuccess reports whether the most recent connectivity test succeeded.
	LastTestSuccess bool
	CreateTime      time.Time
	UpdateTime      time.Time
}

// TestDestinationResult holds the outcome of a Destination connectivity test.
type TestDestinationResult struct {
	Success bool
	// ErrorMessage is nil when the test succeeded. Non-nil on failure.
	ErrorMessage   *string
	HTTPStatusCode int32
	LatencyMS      int64
}

// DestinationTestResultItem is a single entry from ListTestResults.
type DestinationTestResultItem struct {
	Success bool
	// ErrorMessage is nil when no error occurred. Non-nil for failed tests.
	ErrorMessage   *string
	HTTPStatusCode int32
	LatencyMS      int64
	TestTime       time.Time
}

// ValidateFilterResult holds the outcome of a ValidateFilter call.
type ValidateFilterResult struct {
	Valid bool
	// ErrorMessage is nil when Valid is true. Non-nil when the filter is invalid.
	ErrorMessage *string
	MatchedCount int32
	TotalCount   int32
}

// Template is a pre-approved message template (e.g. WhatsApp template).
// Templates are immutable after creation.
type Template struct {
	ID              string
	ConnectorID     string
	TemplateName    string
	Language        string
	Category        TemplateCategory
	ComponentsJSON  string
	Status          TemplateStatus
	RejectionReason string
	ExternalID      string
	CreateTime      time.Time
	UpdateTime      time.Time
}

// ListTemplatesResult is the paginated result of Templates.List.
type ListTemplatesResult struct {
	Items         []Template
	NextPageToken string
	TotalSize     int64
}

// DeliveryEvent is a sealed interface for all event types delivered to a
// Destination webhook. Use a type switch on the concrete types below.
type DeliveryEvent interface{ deliveryEvent() }

// DeliveryEventEnvelope contains envelope metadata present in every delivery event.
type DeliveryEventEnvelope struct {
	// DeliveryID is the unique ID of this delivery attempt.
	DeliveryID string
	// DestinationID is the ID of the Destination this event was sent to.
	DestinationID string
	// ConnectorID is the ID of the Connector that generated the event.
	ConnectorID string
	// EnqueueTime is when the event was enqueued for delivery.
	EnqueueTime time.Time
}

// MessageDeliveryEvent is delivered when a new Message arrives on a Connector.
type MessageDeliveryEvent struct {
	DeliveryEventEnvelope
	Message Message
}

func (*MessageDeliveryEvent) deliveryEvent() {}

// DeliveryUpdateDeliveryEvent is delivered when a Message delivery status changes.
type DeliveryUpdateDeliveryEvent struct {
	DeliveryEventEnvelope
	Delivery MessageDelivery
}

func (*DeliveryUpdateDeliveryEvent) deliveryEvent() {}

// ReadReceiptDeliveryEvent is delivered when messages have been read.
type ReadReceiptDeliveryEvent struct {
	DeliveryEventEnvelope
	ReadReceipt ReadReceiptEvent
}

func (*ReadReceiptDeliveryEvent) deliveryEvent() {}

// TypingIndicatorDeliveryEvent is delivered when a typing indicator changes.
type TypingIndicatorDeliveryEvent struct {
	DeliveryEventEnvelope
	TypingIndicator TypingIndicatorEvent
}

func (*TypingIndicatorDeliveryEvent) deliveryEvent() {}

// UnknownDeliveryEvent is delivered when the event type is not recognised.
type UnknownDeliveryEvent struct {
	DeliveryEventEnvelope
}

func (*UnknownDeliveryEvent) deliveryEvent() {}
