package outbox

import outboxv1 "github.com/getoutbox/outbox-go/outboxv1"

// AccountSource indicates how an Account was created.
type AccountSource = outboxv1.Account_Source

const (
	AccountSourceUnspecified AccountSource = outboxv1.Account_SOURCE_UNSPECIFIED
	AccountSourceAPI         AccountSource = outboxv1.Account_SOURCE_API
	AccountSourceAuto        AccountSource = outboxv1.Account_SOURCE_AUTO
)

// ConnectorState represents the lifecycle state of a Connector.
type ConnectorState = outboxv1.Connector_State

const (
	ConnectorStateUnspecified ConnectorState = outboxv1.Connector_STATE_UNSPECIFIED
	ConnectorStateActive      ConnectorState = outboxv1.Connector_STATE_ACTIVE
	ConnectorStateInactive    ConnectorState = outboxv1.Connector_STATE_INACTIVE
	ConnectorStateAuthorizing ConnectorState = outboxv1.Connector_STATE_AUTHORIZING
	ConnectorStateError       ConnectorState = outboxv1.Connector_STATE_ERROR
)

// MessageDirection indicates whether a Message was sent or received.
type MessageDirection = outboxv1.Message_Direction

const (
	MessageDirectionUnspecified MessageDirection = outboxv1.Message_DIRECTION_UNSPECIFIED
	MessageDirectionInbound     MessageDirection = outboxv1.Message_DIRECTION_INBOUND
	MessageDirectionOutbound    MessageDirection = outboxv1.Message_DIRECTION_OUTBOUND
)

// MessageDeliveryStatus represents the delivery state of a sent Message.
type MessageDeliveryStatus = outboxv1.MessageDelivery_Status

const (
	MessageDeliveryStatusUnspecified MessageDeliveryStatus = outboxv1.MessageDelivery_STATUS_UNSPECIFIED
	MessageDeliveryStatusPending     MessageDeliveryStatus = outboxv1.MessageDelivery_STATUS_PENDING
	MessageDeliveryStatusDelivered   MessageDeliveryStatus = outboxv1.MessageDelivery_STATUS_DELIVERED
	MessageDeliveryStatusDisplayed   MessageDeliveryStatus = outboxv1.MessageDelivery_STATUS_DISPLAYED
	MessageDeliveryStatusProcessed   MessageDeliveryStatus = outboxv1.MessageDelivery_STATUS_PROCESSED
	MessageDeliveryStatusFailed      MessageDeliveryStatus = outboxv1.MessageDelivery_STATUS_FAILED
	MessageDeliveryStatusExpired     MessageDeliveryStatus = outboxv1.MessageDelivery_STATUS_EXPIRED
)

// MessagePartDisposition describes how a MessagePart should be rendered.
type MessagePartDisposition = outboxv1.MessagePart_Disposition

const (
	MessagePartDispositionUnspecified MessagePartDisposition = outboxv1.MessagePart_DISPOSITION_UNSPECIFIED
	MessagePartDispositionRender      MessagePartDisposition = outboxv1.MessagePart_DISPOSITION_RENDER
	MessagePartDispositionReaction    MessagePartDisposition = outboxv1.MessagePart_DISPOSITION_REACTION
	MessagePartDispositionAttachment  MessagePartDisposition = outboxv1.MessagePart_DISPOSITION_ATTACHMENT
	MessagePartDispositionInline      MessagePartDisposition = outboxv1.MessagePart_DISPOSITION_INLINE
)

// MessageDeletionScope controls who a message deletion applies to.
type MessageDeletionScope = outboxv1.Message_DeletionScope

const (
	MessageDeletionScopeUnspecified MessageDeletionScope = outboxv1.Message_DELETION_SCOPE_UNSPECIFIED
	MessageDeletionScopeForSender   MessageDeletionScope = outboxv1.Message_DELETION_SCOPE_FOR_SENDER
	MessageDeletionScopeForEveryone MessageDeletionScope = outboxv1.Message_DELETION_SCOPE_FOR_EVERYONE
)

// DestinationState represents the operational state of a Destination.
type DestinationState = outboxv1.Destination_State

const (
	DestinationStateUnspecified DestinationState = outboxv1.Destination_STATE_UNSPECIFIED
	DestinationStateActive      DestinationState = outboxv1.Destination_STATE_ACTIVE
	DestinationStatePaused      DestinationState = outboxv1.Destination_STATE_PAUSED
	DestinationStateDegraded    DestinationState = outboxv1.Destination_STATE_DEGRADED
)

// DestinationEventType enumerates the event categories a Destination can subscribe to.
type DestinationEventType = outboxv1.Destination_EventType

const (
	DestinationEventTypeUnspecified     DestinationEventType = outboxv1.Destination_EVENT_TYPE_UNSPECIFIED
	DestinationEventTypeMessage         DestinationEventType = outboxv1.Destination_EVENT_TYPE_MESSAGE
	DestinationEventTypeDeliveryUpdate  DestinationEventType = outboxv1.Destination_EVENT_TYPE_DELIVERY_UPDATE
	DestinationEventTypeReadReceipt     DestinationEventType = outboxv1.Destination_EVENT_TYPE_READ_RECEIPT
	DestinationEventTypeTypingIndicator DestinationEventType = outboxv1.Destination_EVENT_TYPE_TYPING_INDICATOR
)

// DestinationPayloadFormat controls the serialization format of event payloads.
type DestinationPayloadFormat = outboxv1.Destination_PayloadFormat

const (
	DestinationPayloadFormatUnspecified DestinationPayloadFormat = outboxv1.Destination_PAYLOAD_FORMAT_UNSPECIFIED
	DestinationPayloadFormatJSON        DestinationPayloadFormat = outboxv1.Destination_PAYLOAD_FORMAT_JSON
	DestinationPayloadFormatProtoBinary DestinationPayloadFormat = outboxv1.Destination_PAYLOAD_FORMAT_PROTO_BINARY
)
