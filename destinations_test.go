package outbox

import (
	"context"
	"slices"
	"testing"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"
)

func newDestinationSvc(m *mockDestinationClient) *DestinationsService {
	return &DestinationsService{client: m}
}

func okDestination(id string) *outboxv1.Destination {
	return &outboxv1.Destination{Name: "destinations/" + id}
}

func TestDestinationsService_Create_Fields(t *testing.T) {
	var got *outboxv1.CreateDestinationRequest
	svc := newDestinationSvc(&mockDestinationClient{
		createDestination: func(_ context.Context, req *connect.Request[outboxv1.CreateDestinationRequest]) (*connect.Response[outboxv1.CreateDestinationResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.CreateDestinationResponse{Destination: okDestination("d1")}), nil
		},
	})
	if _, err := svc.Create(context.Background(), CreateDestinationInput{
		DisplayName:   "My Dest",
		RequestID:     "req-1",
		DestinationID: "custom-id",
	}); err != nil {
		t.Fatal(err)
	}
	if got.Destination.DisplayName != "My Dest" {
		t.Errorf("DisplayName = %q, want My Dest", got.Destination.DisplayName)
	}
	if got.RequestId != "req-1" {
		t.Errorf("RequestId = %q, want req-1", got.RequestId)
	}
	if got.DestinationId != "custom-id" {
		t.Errorf("DestinationId = %q, want custom-id", got.DestinationId)
	}
}

func TestDestinationsService_Create_WithTarget(t *testing.T) {
	var got *outboxv1.CreateDestinationRequest
	svc := newDestinationSvc(&mockDestinationClient{
		createDestination: func(_ context.Context, req *connect.Request[outboxv1.CreateDestinationRequest]) (*connect.Response[outboxv1.CreateDestinationResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.CreateDestinationResponse{Destination: okDestination("d1")}), nil
		},
	})
	target := &outboxv1.Destination_Webhook{Webhook: &outboxv1.WebhookTarget{Url: "https://example.com/hook"}}
	if _, err := svc.Create(context.Background(), CreateDestinationInput{
		DisplayName: "My Webhook",
		Target:      target,
	}); err != nil {
		t.Fatal(err)
	}
	if _, ok := got.Destination.Target.(*outboxv1.Destination_Webhook); !ok {
		t.Errorf("Target type = %T, want *outboxv1.Destination_Webhook", got.Destination.Target)
	}
}

func TestDestinationsService_Create_EmptyResponse(t *testing.T) {
	svc := newDestinationSvc(&mockDestinationClient{
		createDestination: func(_ context.Context, _ *connect.Request[outboxv1.CreateDestinationRequest]) (*connect.Response[outboxv1.CreateDestinationResponse], error) {
			return connect.NewResponse(&outboxv1.CreateDestinationResponse{}), nil
		},
	})
	if _, err := svc.Create(context.Background(), CreateDestinationInput{}); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestDestinationsService_Get_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newDestinationSvc(&mockDestinationClient{
		getDestination: func(_ context.Context, req *connect.Request[outboxv1.GetDestinationRequest]) (*connect.Response[outboxv1.GetDestinationResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.GetDestinationResponse{Destination: okDestination("d1")}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "d1"); err != nil {
		t.Fatal(err)
	}
	if gotName != "destinations/d1" {
		t.Errorf("Name = %q, want destinations/d1", gotName)
	}
}

func TestDestinationsService_Get_EmptyResponse(t *testing.T) {
	svc := newDestinationSvc(&mockDestinationClient{
		getDestination: func(_ context.Context, _ *connect.Request[outboxv1.GetDestinationRequest]) (*connect.Response[outboxv1.GetDestinationResponse], error) {
			return connect.NewResponse(&outboxv1.GetDestinationResponse{}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "d1"); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestDestinationsService_List_PropagatesOptions(t *testing.T) {
	var got *outboxv1.ListDestinationsRequest
	svc := newDestinationSvc(&mockDestinationClient{
		listDestinations: func(_ context.Context, req *connect.Request[outboxv1.ListDestinationsRequest]) (*connect.Response[outboxv1.ListDestinationsResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.ListDestinationsResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), &ListDestinationsOptions{
		PageSize:  5,
		PageToken: "page2",
		Filter:    "state=='ACTIVE'",
		OrderBy:   "display_name asc",
	}); err != nil {
		t.Fatal(err)
	}
	if got.PageSize != 5 {
		t.Errorf("PageSize = %d, want 5", got.PageSize)
	}
	if got.PageToken != "page2" {
		t.Errorf("PageToken = %q, want page2", got.PageToken)
	}
	if got.Filter != "state=='ACTIVE'" {
		t.Errorf("Filter = %q, want state=='ACTIVE'", got.Filter)
	}
	if got.OrderBy != "display_name asc" {
		t.Errorf("OrderBy = %q, want display_name asc", got.OrderBy)
	}
}

func TestDestinationsService_List_NilOptions(t *testing.T) {
	svc := newDestinationSvc(&mockDestinationClient{
		listDestinations: func(_ context.Context, _ *connect.Request[outboxv1.ListDestinationsRequest]) (*connect.Response[outboxv1.ListDestinationsResponse], error) {
			return connect.NewResponse(&outboxv1.ListDestinationsResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestDestinationsService_List_ResultStructure(t *testing.T) {
	svc := newDestinationSvc(&mockDestinationClient{
		listDestinations: func(_ context.Context, _ *connect.Request[outboxv1.ListDestinationsRequest]) (*connect.Response[outboxv1.ListDestinationsResponse], error) {
			return connect.NewResponse(&outboxv1.ListDestinationsResponse{
				Destinations:  []*outboxv1.Destination{okDestination("d1"), okDestination("d2")},
				NextPageToken: "next",
				TotalSize:     2,
			}), nil
		},
	})
	res, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(res.Items))
	}
	if res.Items[0].ID != "d1" {
		t.Errorf("Items[0].ID = %q, want d1", res.Items[0].ID)
	}
	if res.NextPageToken != "next" {
		t.Errorf("NextPageToken = %q, want next", res.NextPageToken)
	}
	if res.TotalSize != 2 {
		t.Errorf("TotalSize = %d, want 2", res.TotalSize)
	}
}

func TestDestinationsService_Update_TargetOnly(t *testing.T) {
	var got *outboxv1.UpdateDestinationRequest
	svc := newDestinationSvc(&mockDestinationClient{
		updateDestination: func(_ context.Context, req *connect.Request[outboxv1.UpdateDestinationRequest]) (*connect.Response[outboxv1.UpdateDestinationResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateDestinationResponse{Destination: okDestination("d1")}), nil
		},
	})
	newTarget := &outboxv1.Destination_Webhook{Webhook: &outboxv1.WebhookTarget{Url: "https://new.example.com/hook"}}
	if _, err := svc.Update(context.Background(), UpdateDestinationInput{ID: "d1", Target: newTarget}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil || !slices.Equal(got.UpdateMask.Paths, []string{"webhook"}) {
		t.Errorf("paths = %v, want [webhook]", got.UpdateMask.Paths)
	}
	if _, ok := got.Destination.Target.(*outboxv1.Destination_Webhook); !ok {
		t.Errorf("Target type = %T, want *outboxv1.Destination_Webhook", got.Destination.Target)
	}
}

func TestDestinationsService_Update_DisplayNameOnly(t *testing.T) {
	var got *outboxv1.UpdateDestinationRequest
	svc := newDestinationSvc(&mockDestinationClient{
		updateDestination: func(_ context.Context, req *connect.Request[outboxv1.UpdateDestinationRequest]) (*connect.Response[outboxv1.UpdateDestinationResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateDestinationResponse{Destination: okDestination("d1")}), nil
		},
	})
	name := "Updated"
	if _, err := svc.Update(context.Background(), UpdateDestinationInput{ID: "d1", DisplayName: &name}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if !slices.Equal(got.UpdateMask.Paths, []string{"display_name"}) {
		t.Errorf("paths = %v, want [display_name]", got.UpdateMask.Paths)
	}
	if got.Destination.DisplayName != "Updated" {
		t.Errorf("DisplayName = %q, want Updated", got.Destination.DisplayName)
	}
}

func TestDestinationsService_Update_FilterOnly(t *testing.T) {
	var got *outboxv1.UpdateDestinationRequest
	svc := newDestinationSvc(&mockDestinationClient{
		updateDestination: func(_ context.Context, req *connect.Request[outboxv1.UpdateDestinationRequest]) (*connect.Response[outboxv1.UpdateDestinationResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateDestinationResponse{Destination: okDestination("d1")}), nil
		},
	})
	f := "connector_id=='c1'"
	if _, err := svc.Update(context.Background(), UpdateDestinationInput{ID: "d1", Filter: &f}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil || !slices.Equal(got.UpdateMask.Paths, []string{"filter"}) {
		t.Errorf("paths = %v, want [filter]", got.UpdateMask.Paths)
	}
}

func TestDestinationsService_Update_PayloadFormatOnly(t *testing.T) {
	var got *outboxv1.UpdateDestinationRequest
	svc := newDestinationSvc(&mockDestinationClient{
		updateDestination: func(_ context.Context, req *connect.Request[outboxv1.UpdateDestinationRequest]) (*connect.Response[outboxv1.UpdateDestinationResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateDestinationResponse{Destination: okDestination("d1")}), nil
		},
	})
	pf := DestinationPayloadFormatProtoBinary
	if _, err := svc.Update(context.Background(), UpdateDestinationInput{ID: "d1", PayloadFormat: &pf}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil || !slices.Equal(got.UpdateMask.Paths, []string{"payload_format"}) {
		t.Errorf("paths = %v, want [payload_format]", got.UpdateMask.Paths)
	}
}

func TestDestinationsService_Update_EventTypesOnly(t *testing.T) {
	var got *outboxv1.UpdateDestinationRequest
	svc := newDestinationSvc(&mockDestinationClient{
		updateDestination: func(_ context.Context, req *connect.Request[outboxv1.UpdateDestinationRequest]) (*connect.Response[outboxv1.UpdateDestinationResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateDestinationResponse{Destination: okDestination("d1")}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateDestinationInput{
		ID:         "d1",
		EventTypes: []DestinationEventType{DestinationEventTypeMessage},
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil || !slices.Equal(got.UpdateMask.Paths, []string{"event_types"}) {
		t.Errorf("paths = %v, want [event_types]", got.UpdateMask.Paths)
	}
}

func TestDestinationsService_Update_NoFields_NilMask(t *testing.T) {
	var got *outboxv1.UpdateDestinationRequest
	svc := newDestinationSvc(&mockDestinationClient{
		updateDestination: func(_ context.Context, req *connect.Request[outboxv1.UpdateDestinationRequest]) (*connect.Response[outboxv1.UpdateDestinationResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateDestinationResponse{Destination: okDestination("d1")}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateDestinationInput{ID: "d1"}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask != nil {
		t.Errorf("expected nil update mask, got %v", got.UpdateMask.Paths)
	}
}

func TestDestinationsService_Update_MultipleFields(t *testing.T) {
	var got *outboxv1.UpdateDestinationRequest
	svc := newDestinationSvc(&mockDestinationClient{
		updateDestination: func(_ context.Context, req *connect.Request[outboxv1.UpdateDestinationRequest]) (*connect.Response[outboxv1.UpdateDestinationResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateDestinationResponse{Destination: okDestination("d1")}), nil
		},
	})
	name := "New"
	filter := "true"
	if _, err := svc.Update(context.Background(), UpdateDestinationInput{
		ID:          "d1",
		DisplayName: &name,
		Filter:      &filter,
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if !slices.Contains(got.UpdateMask.Paths, "display_name") {
		t.Errorf("paths %v missing display_name", got.UpdateMask.Paths)
	}
	if !slices.Contains(got.UpdateMask.Paths, "filter") {
		t.Errorf("paths %v missing filter", got.UpdateMask.Paths)
	}
}

func TestDestinationsService_Delete_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newDestinationSvc(&mockDestinationClient{
		deleteDestination: func(_ context.Context, req *connect.Request[outboxv1.DeleteDestinationRequest]) (*connect.Response[outboxv1.DeleteDestinationResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.DeleteDestinationResponse{}), nil
		},
	})
	if err := svc.Delete(context.Background(), "dest-1"); err != nil {
		t.Fatal(err)
	}
	if gotName != "destinations/dest-1" {
		t.Errorf("Name = %q, want destinations/dest-1", gotName)
	}
}

func TestDestinationsService_Test_Success(t *testing.T) {
	svc := newDestinationSvc(&mockDestinationClient{
		testDestination: func(_ context.Context, _ *connect.Request[outboxv1.TestDestinationRequest]) (*connect.Response[outboxv1.TestDestinationResponse], error) {
			return connect.NewResponse(&outboxv1.TestDestinationResponse{Success: true}), nil
		},
	})
	res, err := svc.Test(context.Background(), "d1")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Error("expected Success=true")
	}
	if res.ErrorMessage != nil {
		t.Errorf("expected nil ErrorMessage, got %q", *res.ErrorMessage)
	}
}

func TestDestinationsService_Test_WithErrorMessage(t *testing.T) {
	msg := "connection refused"
	svc := newDestinationSvc(&mockDestinationClient{
		testDestination: func(_ context.Context, _ *connect.Request[outboxv1.TestDestinationRequest]) (*connect.Response[outboxv1.TestDestinationResponse], error) {
			return connect.NewResponse(&outboxv1.TestDestinationResponse{
				Success:      false,
				ErrorMessage: &msg,
			}), nil
		},
	})
	res, err := svc.Test(context.Background(), "d1")
	if err != nil {
		t.Fatal(err)
	}
	if res.Success {
		t.Error("expected Success=false")
	}
	if res.ErrorMessage == nil {
		t.Fatal("expected non-nil ErrorMessage")
	}
	if *res.ErrorMessage != msg {
		t.Errorf("ErrorMessage = %q, want %q", *res.ErrorMessage, msg)
	}
}

func TestDestinationsService_Test_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newDestinationSvc(&mockDestinationClient{
		testDestination: func(_ context.Context, req *connect.Request[outboxv1.TestDestinationRequest]) (*connect.Response[outboxv1.TestDestinationResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.TestDestinationResponse{Success: true}), nil
		},
	})
	if _, err := svc.Test(context.Background(), "dest-42"); err != nil {
		t.Fatal(err)
	}
	if gotName != "destinations/dest-42" {
		t.Errorf("Name = %q, want destinations/dest-42", gotName)
	}
}

func TestDestinationsService_Test_PopulatesHTTPStatusAndLatency(t *testing.T) {
	svc := newDestinationSvc(&mockDestinationClient{
		testDestination: func(_ context.Context, _ *connect.Request[outboxv1.TestDestinationRequest]) (*connect.Response[outboxv1.TestDestinationResponse], error) {
			return connect.NewResponse(&outboxv1.TestDestinationResponse{
				Success:        true,
				HttpStatusCode: 200,
				LatencyMs:      42,
			}), nil
		},
	})
	res, err := svc.Test(context.Background(), "d1")
	if err != nil {
		t.Fatal(err)
	}
	if res.HTTPStatusCode != 200 {
		t.Errorf("HTTPStatusCode = %d, want 200", res.HTTPStatusCode)
	}
	if res.LatencyMS != 42 {
		t.Errorf("LatencyMS = %d, want 42", res.LatencyMS)
	}
}

func TestDestinationsService_ListTestResults_BuildsResourceName(t *testing.T) {
	var gotReq *outboxv1.ListDestinationTestResultsRequest
	svc := newDestinationSvc(&mockDestinationClient{
		listDestinationTestResults: func(_ context.Context, req *connect.Request[outboxv1.ListDestinationTestResultsRequest]) (*connect.Response[outboxv1.ListDestinationTestResultsResponse], error) {
			gotReq = req.Msg
			return connect.NewResponse(&outboxv1.ListDestinationTestResultsResponse{}), nil
		},
	})
	if _, err := svc.ListTestResults(context.Background(), "d1", 10); err != nil {
		t.Fatal(err)
	}
	if gotReq.Name != "destinations/d1" {
		t.Errorf("Name = %q, want destinations/d1", gotReq.Name)
	}
	if gotReq.PageSize != 10 {
		t.Errorf("PageSize = %d, want 10", gotReq.PageSize)
	}
}

func TestDestinationsService_ListTestResults_MapsResults(t *testing.T) {
	errMsg := "connection refused"
	svc := newDestinationSvc(&mockDestinationClient{
		listDestinationTestResults: func(_ context.Context, _ *connect.Request[outboxv1.ListDestinationTestResultsRequest]) (*connect.Response[outboxv1.ListDestinationTestResultsResponse], error) {
			return connect.NewResponse(&outboxv1.ListDestinationTestResultsResponse{
				Results: []*outboxv1.DestinationTestResult{
					{
						Success:        true,
						HttpStatusCode: 200,
						LatencyMs:      15,
					},
					{
						Success:        false,
						ErrorMessage:   &errMsg,
						HttpStatusCode: 503,
						LatencyMs:      200,
					},
				},
			}), nil
		},
	})
	items, err := svc.ListTestResults(context.Background(), "d1", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if !items[0].Success {
		t.Error("items[0].Success = false, want true")
	}
	if items[0].HTTPStatusCode != 200 {
		t.Errorf("items[0].HTTPStatusCode = %d, want 200", items[0].HTTPStatusCode)
	}
	if items[0].ErrorMessage != nil {
		t.Errorf("items[0].ErrorMessage = %q, want nil", *items[0].ErrorMessage)
	}
	if items[1].Success {
		t.Error("items[1].Success = true, want false")
	}
	if items[1].ErrorMessage == nil {
		t.Fatal("items[1].ErrorMessage is nil, want non-nil")
	}
	if *items[1].ErrorMessage != errMsg {
		t.Errorf("items[1].ErrorMessage = %q, want %q", *items[1].ErrorMessage, errMsg)
	}
	if items[1].HTTPStatusCode != 503 {
		t.Errorf("items[1].HTTPStatusCode = %d, want 503", items[1].HTTPStatusCode)
	}
}

func TestDestinationsService_ValidateFilter_Valid(t *testing.T) {
	var gotReq *outboxv1.ValidateDestinationFilterRequest
	svc := newDestinationSvc(&mockDestinationClient{
		validateDestinationFilter: func(_ context.Context, req *connect.Request[outboxv1.ValidateDestinationFilterRequest]) (*connect.Response[outboxv1.ValidateDestinationFilterResponse], error) {
			gotReq = req.Msg
			return connect.NewResponse(&outboxv1.ValidateDestinationFilterResponse{
				Valid:        true,
				MatchedCount: 8,
				TotalCount:   10,
			}), nil
		},
	})
	res, err := svc.ValidateFilter(context.Background(), "connector_id=='c1'", 10)
	if err != nil {
		t.Fatal(err)
	}
	if gotReq.Filter != "connector_id=='c1'" {
		t.Errorf("Filter = %q, want connector_id=='c1'", gotReq.Filter)
	}
	if gotReq.SampleSize != 10 {
		t.Errorf("SampleSize = %d, want 10", gotReq.SampleSize)
	}
	if !res.Valid {
		t.Error("Valid = false, want true")
	}
	if res.ErrorMessage != nil {
		t.Errorf("ErrorMessage = %q, want nil", *res.ErrorMessage)
	}
	if res.MatchedCount != 8 {
		t.Errorf("MatchedCount = %d, want 8", res.MatchedCount)
	}
	if res.TotalCount != 10 {
		t.Errorf("TotalCount = %d, want 10", res.TotalCount)
	}
}

func TestDestinationsService_ValidateFilter_Invalid(t *testing.T) {
	errMsg := "unknown field: foo"
	svc := newDestinationSvc(&mockDestinationClient{
		validateDestinationFilter: func(_ context.Context, _ *connect.Request[outboxv1.ValidateDestinationFilterRequest]) (*connect.Response[outboxv1.ValidateDestinationFilterResponse], error) {
			return connect.NewResponse(&outboxv1.ValidateDestinationFilterResponse{
				Valid:        false,
				ErrorMessage: &errMsg,
				MatchedCount: 0,
				TotalCount:   10,
			}), nil
		},
	})
	res, err := svc.ValidateFilter(context.Background(), "foo=='bar'", 10)
	if err != nil {
		t.Fatal(err)
	}
	if res.Valid {
		t.Error("Valid = true, want false")
	}
	if res.ErrorMessage == nil {
		t.Fatal("ErrorMessage is nil, want non-nil")
	}
	if *res.ErrorMessage != errMsg {
		t.Errorf("ErrorMessage = %q, want %q", *res.ErrorMessage, errMsg)
	}
}

func TestMapDestination_EventTypesCopied(t *testing.T) {
	p := &outboxv1.Destination{
		Name:       "destinations/d1",
		EventTypes: []DestinationEventType{DestinationEventTypeMessage, DestinationEventTypeReadReceipt},
	}
	d := mapDestination(p)
	// Mutating the proto slice must not affect the mapped destination.
	p.EventTypes[0] = DestinationEventTypeUnspecified
	if d.EventTypes[0] != DestinationEventTypeMessage {
		t.Errorf("EventTypes[0] = %v after proto mutation, want MESSAGE (expected defensive copy)", d.EventTypes[0])
	}
}
