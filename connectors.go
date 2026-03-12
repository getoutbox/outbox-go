package outbox

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"time"

	"connectrpc.com/connect"
	longrunningpb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"
	outboxv1connect "github.com/getoutbox/outbox-go/gen/outbox/v1/outboxv1connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// operationsGetFunc is the signature used to poll a long-running operation.
type operationsGetFunc func(context.Context, *connect.Request[longrunningpb.GetOperationRequest]) (*connect.Response[longrunningpb.Operation], error)

// ConnectorsService provides operations on Connectors.
type ConnectorsService struct {
	client       outboxv1connect.ConnectorServiceClient
	getOperation operationsGetFunc
	// pollInterval controls how long to wait between LRO polls.
	// A value of 0 skips the sleep (used in tests).
	pollInterval time.Duration
}

// applyConnectorChannelConfig sets the ChannelConfig oneof field on dst from cfg and returns
// the proto field name suitable for use in an update mask. Returns "" if cfg is nil or not
// assignable to the ChannelConfig field (i.e., not a *outboxv1.Connector_Xxx oneof wrapper).
func applyConnectorChannelConfig(dst *outboxv1.Connector, cfg any) string {
	if cfg == nil {
		return ""
	}
	field := reflect.ValueOf(dst).Elem().FieldByName("ChannelConfig")
	cfgVal := reflect.ValueOf(cfg)
	if !cfgVal.Type().AssignableTo(field.Type()) {
		return ""
	}
	field.Set(cfgVal)
	msg := dst.ProtoReflect()
	oneof := msg.Descriptor().Oneofs().ByName("channel_config")
	for i := 0; i < oneof.Fields().Len(); i++ {
		fd := oneof.Fields().Get(i)
		if msg.Has(fd) {
			return string(fd.Name())
		}
	}
	return ""
}

// CreateConnectorInput holds parameters for creating a Connector.
//
// Set ChannelConfig to one of the *outboxv1.Connector_Xxx oneof wrapper types, e.g.:
//
//	&outboxv1.Connector_SlackBot{SlackBot: &outboxv1.SlackBotConfig{BotToken: "xoxb-..."}}
type CreateConnectorInput struct {
	Tags                []string
	RequestID           string
	ChannelConfig       any // *outboxv1.Connector_Xxx oneof wrapper
	ConsentAcknowledged bool
}

// UpdateConnectorInput holds parameters for updating a Connector.
// nil/zero fields are not sent to the server.
//
// Set ChannelConfig to one of the *outboxv1.Connector_Xxx oneof wrapper types to update
// the channel-specific configuration. The update mask field is inferred automatically.
type UpdateConnectorInput struct {
	ID            string
	Tags          []string // nil = don't update
	ChannelConfig any      // nil = don't update; *outboxv1.Connector_Xxx oneof wrapper
	// WebhookURL, if non-empty, updates the connector's webhook URL.
	// To remove a webhook URL, use a separate API call; empty string means "no change".
	WebhookURL string
}

// ListConnectorsOptions configures a List request.
type ListConnectorsOptions struct {
	PageSize  int32
	PageToken string
	Filter    string
	OrderBy   string
}

// ListConnectorsResult is the paginated result of a List call.
type ListConnectorsResult struct {
	Items         []Connector
	NextPageToken string
	TotalSize     int64
}

// Create creates a new Connector. For OAuth-based channels, the returned
// CreateConnectorResult will have a non-empty AuthorizationURL — redirect
// the user there to complete setup. Poll GetConnector until State is ACTIVE.
func (s *ConnectorsService) Create(ctx context.Context, input CreateConnectorInput) (*CreateConnectorResult, error) {
	connector := &outboxv1.Connector{Tags: input.Tags}
	applyConnectorChannelConfig(connector, input.ChannelConfig)
	res, err := s.client.CreateConnector(ctx, connect.NewRequest(&outboxv1.CreateConnectorRequest{
		Connector:           connector,
		RequestId:           input.RequestID,
		ConsentAcknowledged: input.ConsentAcknowledged,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Connector == nil {
		return nil, errEmpty("CreateConnector")
	}
	c := mapConnector(res.Msg.Connector)
	return &CreateConnectorResult{
		Connector:        &c,
		AuthorizationURL: res.Msg.AuthorizationUrl,
	}, nil
}

// Get retrieves a Connector by its ID.
func (s *ConnectorsService) Get(ctx context.Context, id string) (*Connector, error) {
	res, err := s.client.GetConnector(ctx, connect.NewRequest(&outboxv1.GetConnectorRequest{
		Name: "connectors/" + id,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Connector == nil {
		return nil, errEmpty("GetConnector")
	}
	c := mapConnector(res.Msg.Connector)
	return &c, nil
}

// List returns a paginated list of Connectors.
func (s *ConnectorsService) List(ctx context.Context, opts *ListConnectorsOptions) (*ListConnectorsResult, error) {
	r := &outboxv1.ListConnectorsRequest{}
	if opts != nil {
		r.PageSize = opts.PageSize
		r.PageToken = opts.PageToken
		r.Filter = opts.Filter
		r.OrderBy = opts.OrderBy
	}
	res, err := s.client.ListConnectors(ctx, connect.NewRequest(r))
	if err != nil {
		return nil, err
	}
	items := make([]Connector, len(res.Msg.Connectors))
	for i, c := range res.Msg.Connectors {
		items[i] = mapConnector(c)
	}
	return &ListConnectorsResult{
		Items:         items,
		NextPageToken: res.Msg.NextPageToken,
		TotalSize:     int64(res.Msg.TotalSize),
	}, nil
}

// Update updates a Connector. Only the fields indicated by non-nil/non-zero values in
// input are sent to the server via field mask.
func (s *ConnectorsService) Update(ctx context.Context, input UpdateConnectorInput) (*Connector, error) {
	connector := &outboxv1.Connector{Name: "connectors/" + input.ID}
	var paths []string

	if input.Tags != nil {
		connector.Tags = input.Tags
		paths = append(paths, "tags")
	}
	if input.WebhookURL != "" {
		connector.WebhookUrl = input.WebhookURL
		paths = append(paths, "webhook_url")
	}
	if input.ChannelConfig != nil {
		if fieldName := applyConnectorChannelConfig(connector, input.ChannelConfig); fieldName != "" {
			paths = append(paths, fieldName)
		}
	}

	var mask *fieldmaskpb.FieldMask
	if len(paths) > 0 {
		mask = &fieldmaskpb.FieldMask{Paths: paths}
	}

	res, err := s.client.UpdateConnector(ctx, connect.NewRequest(&outboxv1.UpdateConnectorRequest{
		Connector:  connector,
		UpdateMask: mask,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Connector == nil {
		return nil, errEmpty("UpdateConnector")
	}
	c := mapConnector(res.Msg.Connector)
	return &c, nil
}

// Delete deletes a Connector by its ID.
func (s *ConnectorsService) Delete(ctx context.Context, id string) error {
	_, err := s.client.DeleteConnector(ctx, connect.NewRequest(&outboxv1.DeleteConnectorRequest{
		Name: "connectors/" + id,
	}))
	return err
}

// Activate transitions a Connector to STATE_ACTIVE.
func (s *ConnectorsService) Activate(ctx context.Context, id string) (*Connector, error) {
	res, err := s.client.ActivateConnector(ctx, connect.NewRequest(&outboxv1.ActivateConnectorRequest{
		Name: "connectors/" + id,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Connector == nil {
		return nil, errEmpty("ActivateConnector")
	}
	c := mapConnector(res.Msg.Connector)
	return &c, nil
}

// Deactivate transitions a Connector out of STATE_ACTIVE.
func (s *ConnectorsService) Deactivate(ctx context.Context, id string) (*Connector, error) {
	res, err := s.client.DeactivateConnector(ctx, connect.NewRequest(&outboxv1.DeactivateConnectorRequest{
		Name: "connectors/" + id,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Connector == nil {
		return nil, errEmpty("DeactivateConnector")
	}
	c := mapConnector(res.Msg.Connector)
	return &c, nil
}

// Reauthorize triggers a new OAuth flow for the given connector.
// Returns an error with code FAILED_PRECONDITION for static-credential channels.
func (s *ConnectorsService) Reauthorize(ctx context.Context, id string) (*ReauthorizeResult, error) {
	res, err := s.client.ReauthorizeConnector(ctx, connect.NewRequest(&outboxv1.ReauthorizeConnectorRequest{
		Name: "connectors/" + id,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Connector == nil {
		return nil, errEmpty("ReauthorizeConnector")
	}
	c := mapConnector(res.Msg.Connector)
	return &ReauthorizeResult{
		Connector:        &c,
		AuthorizationURL: res.Msg.AuthorizationUrl,
	}, nil
}

// Verify submits a verification code for a connector (e.g. Telegram/Signal).
// password is the Telegram 2FA cloud password; pass "" if not needed.
func (s *ConnectorsService) Verify(ctx context.Context, id, code, password string) (*Connector, error) {
	res, err := s.client.VerifyConnector(ctx, connect.NewRequest(&outboxv1.VerifyConnectorRequest{
		Name:     "connectors/" + id,
		Code:     code,
		Password: password,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Connector == nil {
		return nil, errEmpty("VerifyConnector")
	}
	c := mapConnector(res.Msg.Connector)
	return &c, nil
}

// Detach detaches the provisioned resource from a managed Connector.
func (s *ConnectorsService) Detach(ctx context.Context, id string) (*Connector, error) {
	res, err := s.client.DetachProvisionedResource(ctx, connect.NewRequest(&outboxv1.DetachProvisionedResourceRequest{
		Name: "connectors/" + id,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Connector == nil {
		return nil, errEmpty("DetachProvisionedResource")
	}
	c := mapConnector(res.Msg.Connector)
	return &c, nil
}

// CreateManagedConnectorInput holds parameters for CreateManaged.
type CreateManagedConnectorInput struct {
	Channel    string            // e.g. "sms" or "email". Required.
	Filters    map[string]string // Optional.
	WebhookURL string            // Optional.
	Tags       []string          // Optional.
	RequestID  string            // Optional idempotency token.
}

// CreateManaged provisions a managed Connector for the given channel.
// It calls CreateManagedConnector (LRO) and polls until complete or ctx is cancelled.
func (s *ConnectorsService) CreateManaged(ctx context.Context, input CreateManagedConnectorInput) (*Connector, error) {
	res, err := s.client.CreateManagedConnector(ctx, connect.NewRequest(&outboxv1.CreateManagedConnectorRequest{
		Channel:    input.Channel,
		Filters:    input.Filters,
		WebhookUrl: input.WebhookURL,
		Tags:       input.Tags,
		RequestId:  input.RequestID,
	}))
	if err != nil {
		return nil, err
	}
	op := res.Msg
	for !op.Done {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if s.pollInterval > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(s.pollInterval):
			}
		}
		pollRes, err := s.getOperation(ctx, connect.NewRequest(&longrunningpb.GetOperationRequest{
			Name: op.Name,
		}))
		if err != nil {
			return nil, err
		}
		op = pollRes.Msg
	}
	if opErr := op.GetError(); opErr != nil {
		return nil, fmt.Errorf("outbox: CreateManaged: operation failed: %s (code %d)", opErr.Message, opErr.Code)
	}
	resp := op.GetResponse()
	if resp == nil {
		return nil, fmt.Errorf("outbox: CreateManaged: operation completed with no result")
	}
	var connector outboxv1.Connector
	if err := resp.UnmarshalTo(&connector); err != nil {
		return nil, fmt.Errorf("outbox: CreateManaged: unmarshal response: %w", err)
	}
	c := mapConnector(&connector)
	return &c, nil
}

func mapConnector(p *outboxv1.Connector) Connector {
	provisionedResources := make([]string, 0, len(p.GetProvisionedResources()))
	for _, r := range p.GetProvisionedResources() {
		provisionedResources = append(provisionedResources, ParseID(r))
	}
	return Connector{
		ID:                   ParseID(p.GetName()),
		Kind:                 p.GetKind(),
		State:                p.GetState(),
		Readiness:            p.GetReadiness(),
		ProvisionedResources: provisionedResources,
		WebhookURL:           p.GetWebhookUrl(),
		DisplayName:          p.GetDisplayName(),
		Tags:                 slices.Clone(p.GetTags()),
		ChannelConfig:        p.GetChannelConfig(),
		ErrorMessage:         p.GetErrorMessage(),
		CreateTime:           protoTime(p.GetCreateTime()),
		UpdateTime:           protoTime(p.GetUpdateTime()),
	}
}
