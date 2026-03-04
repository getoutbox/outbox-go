package outbox

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
)

// TestListen_YieldsBufferedEvent verifies that Next() drains the pre-populated
// buffer before making any RPC call, and Event() returns the mapped event.
func TestListen_YieldsBufferedEvent(t *testing.T) {
	svc := newDestinationSvc(&mockDestinationClient{})
	it := svc.Listen(context.Background(), "dest-1")

	proto := &outboxv1.DeliveryEvent{
		DeliveryId:  "del-abc",
		Connector:   "connectors/conn-1",
		Destination: "destinations/dest-1",
	}
	it.buf = []*outboxv1.DeliveryEvent{proto}

	if !it.Next() {
		t.Fatal("Next() = false, want true")
	}
	event := it.Event()
	if event == nil {
		t.Fatal("Event() = nil, want non-nil")
	}
	unk, ok := event.(*UnknownDeliveryEvent)
	if !ok {
		t.Fatalf("Event() type = %T, want *UnknownDeliveryEvent", event)
	}
	if unk.DeliveryID != "del-abc" {
		t.Errorf("DeliveryID = %q, want del-abc", unk.DeliveryID)
	}
	if unk.ConnectorID != "conn-1" {
		t.Errorf("ConnectorID = %q, want conn-1", unk.ConnectorID)
	}
	if it.Err() != nil {
		t.Errorf("Err() = %v, want nil", it.Err())
	}
}

// TestListen_ContextCancelled_StopsIteration verifies that Next() returns false
// without error when the context is cancelled and the buffer is empty.
func TestListen_ContextCancelled_StopsIteration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	svc := newDestinationSvc(&mockDestinationClient{})
	it := svc.Listen(ctx, "dest-1")

	if it.Next() {
		t.Error("Next() = true, want false for cancelled context")
	}
	if it.Err() != nil {
		t.Errorf("Err() = %v, want nil for context cancellation", it.Err())
	}
}

// TestListen_RPCError_StopsWithError verifies that when PollEvents returns an
// error (and ctx is not cancelled), Next() returns false and Err() is non-nil.
func TestListen_RPCError_StopsWithError(t *testing.T) {
	rpcErr := errors.New("network failure")
	svc := newDestinationSvc(&mockDestinationClient{
		pollEvents: func(_ context.Context, _ *connect.Request[outboxv1.PollEventsRequest]) (*connect.Response[outboxv1.PollEventsResponse], error) {
			return nil, rpcErr
		},
	})
	it := svc.Listen(context.Background(), "dest-1")

	if it.Next() {
		t.Error("Next() = true, want false after RPC error")
	}
	if it.Err() == nil {
		t.Fatal("Err() = nil, want non-nil after RPC error")
	}
	if !errors.Is(it.Err(), rpcErr) {
		t.Errorf("Err() = %v, want %v", it.Err(), rpcErr)
	}
}

// TestListen_EmptyResponse_PollsAgain verifies that when PollEvents returns an
// empty event list, Next() calls PollEvents again until events are available.
func TestListen_EmptyResponse_PollsAgain(t *testing.T) {
	callCount := 0
	svc := newDestinationSvc(&mockDestinationClient{
		pollEvents: func(_ context.Context, _ *connect.Request[outboxv1.PollEventsRequest]) (*connect.Response[outboxv1.PollEventsResponse], error) {
			callCount++
			if callCount == 1 {
				// First call: return empty response.
				return connect.NewResponse(&outboxv1.PollEventsResponse{
					Cursor: "cursor-1",
					Events: nil,
				}), nil
			}
			// Second call: return one event.
			return connect.NewResponse(&outboxv1.PollEventsResponse{
				Cursor: "cursor-2",
				Events: []*outboxv1.DeliveryEvent{
					{DeliveryId: "del-1"},
				},
			}), nil
		},
	})
	it := svc.Listen(context.Background(), "dest-1")

	if !it.Next() {
		t.Fatal("Next() = false, want true")
	}
	if callCount != 2 {
		t.Errorf("PollEvents called %d times, want 2", callCount)
	}
}

// TestListen_CursorPropagated verifies that the cursor from a PollEvents
// response is included in the subsequent request.
func TestListen_CursorPropagated(t *testing.T) {
	var secondCursor string
	callCount := 0
	svc := newDestinationSvc(&mockDestinationClient{
		pollEvents: func(_ context.Context, req *connect.Request[outboxv1.PollEventsRequest]) (*connect.Response[outboxv1.PollEventsResponse], error) {
			callCount++
			if callCount == 1 {
				return connect.NewResponse(&outboxv1.PollEventsResponse{
					Cursor: "cursor-abc",
					Events: []*outboxv1.DeliveryEvent{{DeliveryId: "del-1"}},
				}), nil
			}
			secondCursor = req.Msg.GetCursor()
			return connect.NewResponse(&outboxv1.PollEventsResponse{
				Cursor: "cursor-def",
				Events: []*outboxv1.DeliveryEvent{{DeliveryId: "del-2"}},
			}), nil
		},
	})
	it := svc.Listen(context.Background(), "dest-1")

	// Consume first event (triggers first PollEvents call).
	if !it.Next() {
		t.Fatal("first Next() = false, want true")
	}
	// Consume second event (triggers second PollEvents call).
	if !it.Next() {
		t.Fatal("second Next() = false, want true")
	}
	if secondCursor != "cursor-abc" {
		t.Errorf("second request cursor = %q, want cursor-abc", secondCursor)
	}
	if it.Cursor() != "cursor-def" {
		t.Errorf("Cursor() = %q, want cursor-def", it.Cursor())
	}
}

// TestListen_ResumeCursor_SentInFirstRequest verifies that ListenOptions.ResumeCursor
// is passed as the cursor in the very first PollEvents call.
func TestListen_ResumeCursor_SentInFirstRequest(t *testing.T) {
	var firstCursor string
	svc := newDestinationSvc(&mockDestinationClient{
		pollEvents: func(_ context.Context, req *connect.Request[outboxv1.PollEventsRequest]) (*connect.Response[outboxv1.PollEventsResponse], error) {
			firstCursor = req.Msg.GetCursor()
			return connect.NewResponse(&outboxv1.PollEventsResponse{
				Cursor: "cursor-new",
				Events: []*outboxv1.DeliveryEvent{{DeliveryId: "del-1"}},
			}), nil
		},
	})
	it := svc.Listen(context.Background(), "dest-1", ListenOptions{
		ResumeCursor: "cursor-resume",
	})

	if !it.Next() {
		t.Fatal("Next() = false, want true")
	}
	if firstCursor != "cursor-resume" {
		t.Errorf("first request cursor = %q, want cursor-resume", firstCursor)
	}
}

// TestListen_ResourceNameFormat verifies that PollEvents is called with the
// correct resource name format "destinations/{id}".
func TestListen_ResourceNameFormat(t *testing.T) {
	var gotName string
	svc := newDestinationSvc(&mockDestinationClient{
		pollEvents: func(_ context.Context, req *connect.Request[outboxv1.PollEventsRequest]) (*connect.Response[outboxv1.PollEventsResponse], error) {
			gotName = req.Msg.GetName()
			return connect.NewResponse(&outboxv1.PollEventsResponse{
				Events: []*outboxv1.DeliveryEvent{{DeliveryId: "del-1"}},
			}), nil
		},
	})
	it := svc.Listen(context.Background(), "my-dest-id")

	if !it.Next() {
		t.Fatal("Next() = false, want true")
	}
	if gotName != "destinations/my-dest-id" {
		t.Errorf("Name = %q, want destinations/my-dest-id", gotName)
	}
}

// TestListen_OptionsForwarded verifies that MaxEvents and WaitSeconds from
// ListenOptions are forwarded to the PollEvents request.
func TestListen_OptionsForwarded(t *testing.T) {
	var gotReq *outboxv1.PollEventsRequest
	svc := newDestinationSvc(&mockDestinationClient{
		pollEvents: func(_ context.Context, req *connect.Request[outboxv1.PollEventsRequest]) (*connect.Response[outboxv1.PollEventsResponse], error) {
			gotReq = req.Msg
			return connect.NewResponse(&outboxv1.PollEventsResponse{
				Events: []*outboxv1.DeliveryEvent{{DeliveryId: "del-1"}},
			}), nil
		},
	})
	it := svc.Listen(context.Background(), "dest-1", ListenOptions{
		MaxEvents:   50,
		WaitSeconds: 20,
	})

	if !it.Next() {
		t.Fatal("Next() = false, want true")
	}
	if gotReq.GetMaxEvents() != 50 {
		t.Errorf("MaxEvents = %d, want 50", gotReq.GetMaxEvents())
	}
	if gotReq.GetWaitSeconds() != 20 {
		t.Errorf("WaitSeconds = %d, want 20", gotReq.GetWaitSeconds())
	}
}
