package outbox

import outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"

// AccountSource indicates how an Account was created.
type AccountSource = outboxv1.Account_Source

const (
	AccountSourceUnspecified AccountSource = outboxv1.Account_SOURCE_UNSPECIFIED
	AccountSourceAPI         AccountSource = outboxv1.Account_SOURCE_API
	AccountSourceAuto        AccountSource = outboxv1.Account_SOURCE_AUTO
)

// ConnectorState represents the lifecycle state of a Connector.
type ConnectorState = outboxv1.ConnectorState

const (
	ConnectorStateUnspecified ConnectorState = outboxv1.ConnectorState_CONNECTOR_STATE_UNSPECIFIED
	ConnectorStateActive      ConnectorState = outboxv1.ConnectorState_CONNECTOR_STATE_ACTIVE
	ConnectorStateInactive    ConnectorState = outboxv1.ConnectorState_CONNECTOR_STATE_INACTIVE
	ConnectorStateAuthorizing ConnectorState = outboxv1.ConnectorState_CONNECTOR_STATE_AUTHORIZING
	ConnectorStateError       ConnectorState = outboxv1.ConnectorState_CONNECTOR_STATE_ERROR
)

// ConnectorKind indicates whether a Connector is managed, user, or bot.
type ConnectorKind = outboxv1.ConnectorKind

const (
	ConnectorKindUnspecified ConnectorKind = outboxv1.ConnectorKind_CONNECTOR_KIND_UNSPECIFIED
	ConnectorKindManaged     ConnectorKind = outboxv1.ConnectorKind_CONNECTOR_KIND_MANAGED
	ConnectorKindUser        ConnectorKind = outboxv1.ConnectorKind_CONNECTOR_KIND_USER
	ConnectorKindBot         ConnectorKind = outboxv1.ConnectorKind_CONNECTOR_KIND_BOT
)

// ConnectorReadiness indicates whether a Connector's provisioned resources are ready.
type ConnectorReadiness = outboxv1.ConnectorReadiness

const (
	ConnectorReadinessUnspecified       ConnectorReadiness = outboxv1.ConnectorReadiness_CONNECTOR_READINESS_UNSPECIFIED
	ConnectorReadinessReady             ConnectorReadiness = outboxv1.ConnectorReadiness_CONNECTOR_READINESS_READY
	ConnectorReadinessPendingCompliance ConnectorReadiness = outboxv1.ConnectorReadiness_CONNECTOR_READINESS_PENDING_COMPLIANCE
	ConnectorReadinessResourceNotActive ConnectorReadiness = outboxv1.ConnectorReadiness_CONNECTOR_READINESS_RESOURCE_NOT_ACTIVE
	ConnectorReadinessResourceSuspended ConnectorReadiness = outboxv1.ConnectorReadiness_CONNECTOR_READINESS_RESOURCE_SUSPENDED
)

// ProvisionedResourceState indicates the state of a provisioned resource associated with a connector.
// This enum is exported for completeness with the API schema; individual states are not currently
// surfaced through the Connector type's ProvisionedResources field.
type ProvisionedResourceState = outboxv1.ProvisionedResourceState

const (
	ProvisionedResourceStateUnspecified  ProvisionedResourceState = outboxv1.ProvisionedResourceState_PROVISIONED_RESOURCE_STATE_UNSPECIFIED
	ProvisionedResourceStatePending      ProvisionedResourceState = outboxv1.ProvisionedResourceState_PROVISIONED_RESOURCE_STATE_PENDING
	ProvisionedResourceStateProvisioning ProvisionedResourceState = outboxv1.ProvisionedResourceState_PROVISIONED_RESOURCE_STATE_PROVISIONING
	ProvisionedResourceStateActive       ProvisionedResourceState = outboxv1.ProvisionedResourceState_PROVISIONED_RESOURCE_STATE_ACTIVE
	ProvisionedResourceStateSuspended    ProvisionedResourceState = outboxv1.ProvisionedResourceState_PROVISIONED_RESOURCE_STATE_SUSPENDED
	ProvisionedResourceStateReleased     ProvisionedResourceState = outboxv1.ProvisionedResourceState_PROVISIONED_RESOURCE_STATE_RELEASED
	ProvisionedResourceStateFailed       ProvisionedResourceState = outboxv1.ProvisionedResourceState_PROVISIONED_RESOURCE_STATE_FAILED
	ProvisionedResourceStateCancelling   ProvisionedResourceState = outboxv1.ProvisionedResourceState_PROVISIONED_RESOURCE_STATE_CANCELLING
	ProvisionedResourceStatePorting      ProvisionedResourceState = outboxv1.ProvisionedResourceState_PROVISIONED_RESOURCE_STATE_PORTING
	ProvisionedResourceStatePortFailed   ProvisionedResourceState = outboxv1.ProvisionedResourceState_PROVISIONED_RESOURCE_STATE_PORT_FAILED
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

// TemplateStatus is the approval state of a Template.
type TemplateStatus = outboxv1.TemplateStatus

const (
	TemplateStatusUnspecified TemplateStatus = outboxv1.TemplateStatus_TEMPLATE_STATUS_UNSPECIFIED
	TemplateStatusPending     TemplateStatus = outboxv1.TemplateStatus_TEMPLATE_STATUS_PENDING
	TemplateStatusApproved    TemplateStatus = outboxv1.TemplateStatus_TEMPLATE_STATUS_APPROVED
	TemplateStatusRejected    TemplateStatus = outboxv1.TemplateStatus_TEMPLATE_STATUS_REJECTED
	TemplateStatusPaused      TemplateStatus = outboxv1.TemplateStatus_TEMPLATE_STATUS_PAUSED
	TemplateStatusDisabled    TemplateStatus = outboxv1.TemplateStatus_TEMPLATE_STATUS_DISABLED
)

// TemplateCategory classifies the purpose of a Template.
type TemplateCategory = outboxv1.Template_Category

const (
	TemplateCategoryUnspecified    TemplateCategory = outboxv1.Template_CATEGORY_UNSPECIFIED
	TemplateCategoryUtility        TemplateCategory = outboxv1.Template_CATEGORY_UTILITY
	TemplateCategoryMarketing      TemplateCategory = outboxv1.Template_CATEGORY_MARKETING
	TemplateCategoryAuthentication TemplateCategory = outboxv1.Template_CATEGORY_AUTHENTICATION
)
