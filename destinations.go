package outbox

import (
	"context"
	"reflect"
	"slices"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"
	outboxv1connect "github.com/getoutbox/outbox-go/gen/outbox/v1/outboxv1connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// DestinationsService provides operations on Destinations.
type DestinationsService struct {
	client outboxv1connect.DestinationServiceClient
}

// applyDestinationTarget sets the Target oneof field on dst from target and returns
// the proto field name suitable for use in an update mask. Returns "" if target is nil or not
// assignable to the Target field (i.e., not a *outboxv1.Destination_Xxx oneof wrapper).
func applyDestinationTarget(dst *outboxv1.Destination, target any) string {
	if target == nil {
		return ""
	}
	field := reflect.ValueOf(dst).Elem().FieldByName("Target")
	targetVal := reflect.ValueOf(target)
	if !targetVal.Type().AssignableTo(field.Type()) {
		return ""
	}
	field.Set(targetVal)
	msg := dst.ProtoReflect()
	oneof := msg.Descriptor().Oneofs().ByName("target")
	for i := 0; i < oneof.Fields().Len(); i++ {
		fd := oneof.Fields().Get(i)
		if msg.Has(fd) {
			return string(fd.Name())
		}
	}
	return ""
}

// CreateDestinationInput holds parameters for creating a Destination.
//
// Set Target to one of the *outboxv1.Destination_Xxx oneof wrapper types, e.g.:
//
//	&outboxv1.Destination_Webhook{Webhook: &outboxv1.WebhookTarget{Url: "https://..."}}
type CreateDestinationInput struct {
	DisplayName   string
	EventTypes    []DestinationEventType
	Filter        string
	PayloadFormat DestinationPayloadFormat
	RequestID     string
	DestinationID string // optional client-assigned ID
	Target        any    // *outboxv1.Destination_Xxx oneof wrapper
}

// UpdateDestinationInput holds parameters for updating a Destination.
// Pointer fields: nil means don't update.
//
// Set Target to one of the *outboxv1.Destination_Xxx oneof wrapper types to update
// the destination-specific configuration. The update mask field is inferred automatically.
type UpdateDestinationInput struct {
	ID            string
	DisplayName   *string
	Filter        *string
	PayloadFormat *DestinationPayloadFormat
	EventTypes    []DestinationEventType // nil = don't update
	Target        any                    // nil = don't update; *outboxv1.Destination_Xxx oneof wrapper
}

// ListDestinationsOptions configures a List request.
type ListDestinationsOptions struct {
	PageSize  int32
	PageToken string
	Filter    string
	OrderBy   string
}

// ListDestinationsResult is the paginated result of a List call.
type ListDestinationsResult struct {
	Items         []Destination
	NextPageToken string
	TotalSize     int64
}

// Create creates a new Destination.
func (s *DestinationsService) Create(ctx context.Context, input CreateDestinationInput) (*Destination, error) {
	dest := &outboxv1.Destination{
		DisplayName:   input.DisplayName,
		EventTypes:    input.EventTypes,
		Filter:        input.Filter,
		PayloadFormat: input.PayloadFormat,
	}
	applyDestinationTarget(dest, input.Target)
	res, err := s.client.CreateDestination(ctx, connect.NewRequest(&outboxv1.CreateDestinationRequest{
		Destination:   dest,
		RequestId:     input.RequestID,
		DestinationId: input.DestinationID,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Destination == nil {
		return nil, errEmpty("CreateDestination")
	}
	d := mapDestination(res.Msg.Destination)
	return &d, nil
}

// Get retrieves a Destination by its ID.
func (s *DestinationsService) Get(ctx context.Context, id string) (*Destination, error) {
	res, err := s.client.GetDestination(ctx, connect.NewRequest(&outboxv1.GetDestinationRequest{
		Name: "destinations/" + id,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Destination == nil {
		return nil, errEmpty("GetDestination")
	}
	d := mapDestination(res.Msg.Destination)
	return &d, nil
}

// List returns a paginated list of Destinations.
func (s *DestinationsService) List(ctx context.Context, opts *ListDestinationsOptions) (*ListDestinationsResult, error) {
	r := &outboxv1.ListDestinationsRequest{}
	if opts != nil {
		r.PageSize = opts.PageSize
		r.PageToken = opts.PageToken
		r.Filter = opts.Filter
		r.OrderBy = opts.OrderBy
	}
	res, err := s.client.ListDestinations(ctx, connect.NewRequest(r))
	if err != nil {
		return nil, err
	}
	items := make([]Destination, len(res.Msg.Destinations))
	for i, d := range res.Msg.Destinations {
		items[i] = mapDestination(d)
	}
	return &ListDestinationsResult{
		Items:         items,
		NextPageToken: res.Msg.NextPageToken,
		TotalSize:     int64(res.Msg.TotalSize),
	}, nil
}

// Update updates a Destination. Only the fields indicated by non-nil values in
// input are sent to the server via field mask.
func (s *DestinationsService) Update(ctx context.Context, input UpdateDestinationInput) (*Destination, error) {
	dest := &outboxv1.Destination{Name: "destinations/" + input.ID}
	var paths []string

	if input.DisplayName != nil {
		dest.DisplayName = *input.DisplayName
		paths = append(paths, "display_name")
	}
	if input.Filter != nil {
		dest.Filter = *input.Filter
		paths = append(paths, "filter")
	}
	if input.PayloadFormat != nil {
		dest.PayloadFormat = *input.PayloadFormat
		paths = append(paths, "payload_format")
	}
	if input.EventTypes != nil {
		dest.EventTypes = input.EventTypes
		paths = append(paths, "event_types")
	}
	if input.Target != nil {
		if fieldName := applyDestinationTarget(dest, input.Target); fieldName != "" {
			paths = append(paths, fieldName)
		}
	}

	var mask *fieldmaskpb.FieldMask
	if len(paths) > 0 {
		mask = &fieldmaskpb.FieldMask{Paths: paths}
	}

	res, err := s.client.UpdateDestination(ctx, connect.NewRequest(&outboxv1.UpdateDestinationRequest{
		Destination: dest,
		UpdateMask:  mask,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Destination == nil {
		return nil, errEmpty("UpdateDestination")
	}
	d := mapDestination(res.Msg.Destination)
	return &d, nil
}

// Delete deletes a Destination by its ID.
func (s *DestinationsService) Delete(ctx context.Context, id string) error {
	_, err := s.client.DeleteDestination(ctx, connect.NewRequest(&outboxv1.DeleteDestinationRequest{
		Name: "destinations/" + id,
	}))
	return err
}

// Test tests the connectivity of a Destination.
func (s *DestinationsService) Test(ctx context.Context, id string) (*TestDestinationResult, error) {
	res, err := s.client.TestDestination(ctx, connect.NewRequest(&outboxv1.TestDestinationRequest{
		Name: "destinations/" + id,
	}))
	if err != nil {
		return nil, err
	}
	return &TestDestinationResult{
		Success:        res.Msg.Success,
		ErrorMessage:   res.Msg.ErrorMessage,
		HTTPStatusCode: res.Msg.HttpStatusCode,
		LatencyMS:      res.Msg.LatencyMs,
	}, nil
}

// ListTestResults returns the recent test results for a Destination.
func (s *DestinationsService) ListTestResults(ctx context.Context, id string, pageSize int32) ([]DestinationTestResultItem, error) {
	res, err := s.client.ListDestinationTestResults(ctx, connect.NewRequest(&outboxv1.ListDestinationTestResultsRequest{
		Name:     "destinations/" + id,
		PageSize: pageSize,
	}))
	if err != nil {
		return nil, err
	}
	items := make([]DestinationTestResultItem, len(res.Msg.Results))
	for i, r := range res.Msg.Results {
		item := DestinationTestResultItem{
			Success:        r.GetSuccess(),
			HTTPStatusCode: r.GetHttpStatusCode(),
			LatencyMS:      r.GetLatencyMs(),
			TestTime:       protoTime(r.GetTestTime()),
		}
		if msg := r.GetErrorMessage(); msg != "" {
			item.ErrorMessage = &msg
		}
		items[i] = item
	}
	return items, nil
}

// ValidateFilter validates a CEL filter expression against a sample of recent events.
func (s *DestinationsService) ValidateFilter(ctx context.Context, filter string, sampleSize int32) (*ValidateFilterResult, error) {
	res, err := s.client.ValidateDestinationFilter(ctx, connect.NewRequest(&outboxv1.ValidateDestinationFilterRequest{
		Filter:     filter,
		SampleSize: sampleSize,
	}))
	if err != nil {
		return nil, err
	}
	result := &ValidateFilterResult{
		Valid:        res.Msg.GetValid(),
		MatchedCount: res.Msg.GetMatchedCount(),
		TotalCount:   res.Msg.GetTotalCount(),
	}
	if msg := res.Msg.GetErrorMessage(); msg != "" {
		result.ErrorMessage = &msg
	}
	return result, nil
}

// ListenOptions configures a Listen call.
type ListenOptions struct {
	// ResumeCursor resumes from a known position after process restart.
	ResumeCursor string
	// MaxEvents is the max events per poll (1-100, default 10).
	MaxEvents int32
	// WaitSeconds is the long-poll timeout per request (1-30, default 5).
	WaitSeconds int32
}

// EventIterator iterates over delivery events from a local destination.
// Call Next() to advance; Event() returns the current event; Err() returns
// any error that stopped the iteration.
type EventIterator struct {
	ctx    context.Context
	svc    *DestinationsService
	destID string
	opts   ListenOptions

	cursor  string
	buf     []*outboxv1.DeliveryEvent
	current DeliveryEvent
	err     error
}

// Listen returns an iterator that long-polls for events from a local destination.
// The iterator runs until ctx is cancelled or an unrecoverable error occurs.
// Pass ListenOptions.ResumeCursor to resume from a known position.
func (s *DestinationsService) Listen(ctx context.Context, id string, opts ...ListenOptions) *EventIterator {
	var o ListenOptions
	if len(opts) > 0 {
		o = opts[0]
	}
	return &EventIterator{
		ctx:    ctx,
		svc:    s,
		destID: id,
		opts:   o,
		cursor: o.ResumeCursor,
	}
}

// Next advances the iterator. Returns true if an event is available via Event().
// Returns false when iteration ends (context cancelled or error).
func (it *EventIterator) Next() bool {
	for {
		if len(it.buf) > 0 {
			p := it.buf[0]
			it.buf = it.buf[1:]
			it.current = mapDeliveryEvent(p)
			return true
		}
		if it.ctx.Err() != nil {
			return false
		}
		req := &outboxv1.PollEventsRequest{
			Name:   "destinations/" + it.destID,
			Cursor: it.cursor,
		}
		if it.opts.MaxEvents > 0 {
			req.MaxEvents = it.opts.MaxEvents
		}
		if it.opts.WaitSeconds > 0 {
			req.WaitSeconds = it.opts.WaitSeconds
		}
		res, err := it.svc.client.PollEvents(it.ctx, connect.NewRequest(req))
		if err != nil {
			if it.ctx.Err() != nil {
				return false
			}
			it.err = err
			return false
		}
		it.cursor = res.Msg.GetCursor()
		it.buf = res.Msg.GetEvents()
	}
}

// Event returns the current DeliveryEvent. Only valid after a true return from Next().
func (it *EventIterator) Event() DeliveryEvent { return it.current }

// Cursor returns the cursor from the last poll response.
// Store this value to resume iteration after a process restart.
func (it *EventIterator) Cursor() string { return it.cursor }

// Err returns the error that stopped iteration, if any.
// Returns nil if iteration stopped due to context cancellation.
func (it *EventIterator) Err() error { return it.err }

func mapDestination(p *outboxv1.Destination) Destination {
	d := Destination{
		ID:              ParseID(p.GetName()),
		DisplayName:     p.GetDisplayName(),
		State:           p.GetState(),
		EventTypes:      slices.Clone(p.GetEventTypes()),
		Filter:          p.GetFilter(),
		PayloadFormat:   p.GetPayloadFormat(),
		Target:          p.GetTarget(),
		LastTestSuccess: p.GetLastTestSuccess(),
		CreateTime:      protoTime(p.GetCreateTime()),
		UpdateTime:      protoTime(p.GetUpdateTime()),
	}
	if t := protoTime(p.GetLastTestTime()); !t.IsZero() {
		d.LastTestTime = &t
	}
	return d
}
