package outbox

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	longrunningpb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"
	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

func newConnectorSvc(m *mockConnectorClient) *ConnectorsService {
	return &ConnectorsService{client: m}
}

func okConnectorResp(id string) *connect.Response[outboxv1.CreateConnectorResponse] {
	return connect.NewResponse(&outboxv1.CreateConnectorResponse{
		Connector: &outboxv1.Connector{Name: "connectors/" + id},
	})
}

func TestConnectorsService_Create_RequestID(t *testing.T) {
	var got *outboxv1.CreateConnectorRequest
	svc := newConnectorSvc(&mockConnectorClient{
		createConnector: func(_ context.Context, req *connect.Request[outboxv1.CreateConnectorRequest]) (*connect.Response[outboxv1.CreateConnectorResponse], error) {
			got = req.Msg
			return okConnectorResp("c1"), nil
		},
	})
	if _, err := svc.Create(context.Background(), CreateConnectorInput{RequestID: "req-1"}); err != nil {
		t.Fatal(err)
	}
	if got.RequestId != "req-1" {
		t.Errorf("RequestId = %q, want req-1", got.RequestId)
	}
}

func TestConnectorsService_Create_Tags(t *testing.T) {
	var got *outboxv1.CreateConnectorRequest
	svc := newConnectorSvc(&mockConnectorClient{
		createConnector: func(_ context.Context, req *connect.Request[outboxv1.CreateConnectorRequest]) (*connect.Response[outboxv1.CreateConnectorResponse], error) {
			got = req.Msg
			return okConnectorResp("c1"), nil
		},
	})
	if _, err := svc.Create(context.Background(), CreateConnectorInput{Tags: []string{"a", "b"}}); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got.Connector.Tags, []string{"a", "b"}) {
		t.Errorf("Tags = %v, want [a b]", got.Connector.Tags)
	}
}

func TestConnectorsService_Create_AuthorizationURL(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		createConnector: func(_ context.Context, _ *connect.Request[outboxv1.CreateConnectorRequest]) (*connect.Response[outboxv1.CreateConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.CreateConnectorResponse{
				Connector:        &outboxv1.Connector{Name: "connectors/c1"},
				AuthorizationUrl: "https://auth.example.com/oauth",
			}), nil
		},
	})
	res, err := svc.Create(context.Background(), CreateConnectorInput{})
	if err != nil {
		t.Fatal(err)
	}
	if res.AuthorizationURL != "https://auth.example.com/oauth" {
		t.Errorf("AuthorizationURL = %q, want https://auth.example.com/oauth", res.AuthorizationURL)
	}
}

func TestConnectorsService_Create_EmptyResponse(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		createConnector: func(_ context.Context, _ *connect.Request[outboxv1.CreateConnectorRequest]) (*connect.Response[outboxv1.CreateConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.CreateConnectorResponse{}), nil
		},
	})
	if _, err := svc.Create(context.Background(), CreateConnectorInput{}); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestConnectorsService_Get_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newConnectorSvc(&mockConnectorClient{
		getConnector: func(_ context.Context, req *connect.Request[outboxv1.GetConnectorRequest]) (*connect.Response[outboxv1.GetConnectorResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.GetConnectorResponse{
				Connector: &outboxv1.Connector{Name: req.Msg.Name},
			}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "my-id"); err != nil {
		t.Fatal(err)
	}
	if gotName != "connectors/my-id" {
		t.Errorf("Name = %q, want connectors/my-id", gotName)
	}
}

func TestConnectorsService_Get_EmptyResponse(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		getConnector: func(_ context.Context, _ *connect.Request[outboxv1.GetConnectorRequest]) (*connect.Response[outboxv1.GetConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.GetConnectorResponse{}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "c1"); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestConnectorsService_List_PropagatesOptions(t *testing.T) {
	var got *outboxv1.ListConnectorsRequest
	svc := newConnectorSvc(&mockConnectorClient{
		listConnectors: func(_ context.Context, req *connect.Request[outboxv1.ListConnectorsRequest]) (*connect.Response[outboxv1.ListConnectorsResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.ListConnectorsResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), &ListConnectorsOptions{
		PageSize:  10,
		PageToken: "tok",
		Filter:    "state==ACTIVE",
		OrderBy:   "create_time",
	}); err != nil {
		t.Fatal(err)
	}
	if got.PageSize != 10 {
		t.Errorf("PageSize = %d, want 10", got.PageSize)
	}
	if got.PageToken != "tok" {
		t.Errorf("PageToken = %q, want tok", got.PageToken)
	}
	if got.Filter != "state==ACTIVE" {
		t.Errorf("Filter = %q, want state==ACTIVE", got.Filter)
	}
	if got.OrderBy != "create_time" {
		t.Errorf("OrderBy = %q, want create_time", got.OrderBy)
	}
}

func TestConnectorsService_List_NilOptions(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		listConnectors: func(_ context.Context, _ *connect.Request[outboxv1.ListConnectorsRequest]) (*connect.Response[outboxv1.ListConnectorsResponse], error) {
			return connect.NewResponse(&outboxv1.ListConnectorsResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestConnectorsService_Update_TagsOnly(t *testing.T) {
	var got *outboxv1.UpdateConnectorRequest
	svc := newConnectorSvc(&mockConnectorClient{
		updateConnector: func(_ context.Context, req *connect.Request[outboxv1.UpdateConnectorRequest]) (*connect.Response[outboxv1.UpdateConnectorResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateConnectorResponse{
				Connector: &outboxv1.Connector{Name: "connectors/c1"},
			}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateConnectorInput{ID: "c1", Tags: []string{"a"}}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if !slices.Equal(got.UpdateMask.Paths, []string{"tags"}) {
		t.Errorf("paths = %v, want [tags]", got.UpdateMask.Paths)
	}
}

func TestConnectorsService_Update_ChannelConfigOnly(t *testing.T) {
	var got *outboxv1.UpdateConnectorRequest
	svc := newConnectorSvc(&mockConnectorClient{
		updateConnector: func(_ context.Context, req *connect.Request[outboxv1.UpdateConnectorRequest]) (*connect.Response[outboxv1.UpdateConnectorResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateConnectorResponse{
				Connector: &outboxv1.Connector{Name: "connectors/c1"},
			}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateConnectorInput{
		ID:            "c1",
		ChannelConfig: &outboxv1.Connector_SlackBot{SlackBot: &outboxv1.SlackBotConfig{BotToken: "xoxb"}},
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if len(got.UpdateMask.Paths) != 1 || got.UpdateMask.Paths[0] != "slack_bot" {
		t.Errorf("paths = %v, want [slack_bot]", got.UpdateMask.Paths)
	}
}

func TestConnectorsService_Update_BothTagsAndChannelConfig(t *testing.T) {
	var got *outboxv1.UpdateConnectorRequest
	svc := newConnectorSvc(&mockConnectorClient{
		updateConnector: func(_ context.Context, req *connect.Request[outboxv1.UpdateConnectorRequest]) (*connect.Response[outboxv1.UpdateConnectorResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateConnectorResponse{
				Connector: &outboxv1.Connector{Name: "connectors/c1"},
			}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateConnectorInput{
		ID:            "c1",
		Tags:          []string{"x"},
		ChannelConfig: &outboxv1.Connector_SlackBot{SlackBot: &outboxv1.SlackBotConfig{}},
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if !slices.Contains(got.UpdateMask.Paths, "tags") {
		t.Errorf("paths %v missing 'tags'", got.UpdateMask.Paths)
	}
	if !slices.Contains(got.UpdateMask.Paths, "slack_bot") {
		t.Errorf("paths %v missing 'slack_bot'", got.UpdateMask.Paths)
	}
}

func TestConnectorsService_Update_NoFields_NilMask(t *testing.T) {
	var got *outboxv1.UpdateConnectorRequest
	svc := newConnectorSvc(&mockConnectorClient{
		updateConnector: func(_ context.Context, req *connect.Request[outboxv1.UpdateConnectorRequest]) (*connect.Response[outboxv1.UpdateConnectorResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateConnectorResponse{
				Connector: &outboxv1.Connector{Name: "connectors/c1"},
			}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateConnectorInput{ID: "c1"}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask != nil {
		t.Errorf("expected nil update mask, got %v", got.UpdateMask.Paths)
	}
}

func TestConnectorsService_Update_WebhookURL(t *testing.T) {
	var got *outboxv1.UpdateConnectorRequest
	svc := newConnectorSvc(&mockConnectorClient{
		updateConnector: func(_ context.Context, req *connect.Request[outboxv1.UpdateConnectorRequest]) (*connect.Response[outboxv1.UpdateConnectorResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateConnectorResponse{
				Connector: &outboxv1.Connector{Name: "connectors/c1"},
			}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateConnectorInput{
		ID:         "c1",
		WebhookURL: "https://example.com/webhook",
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if !slices.Contains(got.UpdateMask.Paths, "webhook_url") {
		t.Errorf("paths = %v, want to contain webhook_url", got.UpdateMask.Paths)
	}
	if got.Connector.WebhookUrl != "https://example.com/webhook" {
		t.Errorf("WebhookUrl = %q, want https://example.com/webhook", got.Connector.WebhookUrl)
	}
}

func TestConnectorsService_Update_EmptyResponse(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		updateConnector: func(_ context.Context, _ *connect.Request[outboxv1.UpdateConnectorRequest]) (*connect.Response[outboxv1.UpdateConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.UpdateConnectorResponse{}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateConnectorInput{ID: "c1"}); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestConnectorsService_Delete_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newConnectorSvc(&mockConnectorClient{
		deleteConnector: func(_ context.Context, req *connect.Request[outboxv1.DeleteConnectorRequest]) (*connect.Response[outboxv1.DeleteConnectorResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.DeleteConnectorResponse{}), nil
		},
	})
	if err := svc.Delete(context.Background(), "my-connector"); err != nil {
		t.Fatal(err)
	}
	if gotName != "connectors/my-connector" {
		t.Errorf("Name = %q, want connectors/my-connector", gotName)
	}
}

func TestConnectorsService_Reauthorize_ReturnsURL(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		reauthorizeConnector: func(_ context.Context, _ *connect.Request[outboxv1.ReauthorizeConnectorRequest]) (*connect.Response[outboxv1.ReauthorizeConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.ReauthorizeConnectorResponse{
				Connector:        &outboxv1.Connector{Name: "connectors/c1"},
				AuthorizationUrl: "https://oauth.example.com/reauth",
			}), nil
		},
	})
	res, err := svc.Reauthorize(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if res.AuthorizationURL != "https://oauth.example.com/reauth" {
		t.Errorf("AuthorizationURL = %q, want https://oauth.example.com/reauth", res.AuthorizationURL)
	}
	if res.Connector == nil {
		t.Error("expected Connector, got nil")
	}
}

func TestConnectorsService_Reauthorize_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newConnectorSvc(&mockConnectorClient{
		reauthorizeConnector: func(_ context.Context, req *connect.Request[outboxv1.ReauthorizeConnectorRequest]) (*connect.Response[outboxv1.ReauthorizeConnectorResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.ReauthorizeConnectorResponse{
				Connector: &outboxv1.Connector{Name: req.Msg.Name},
			}), nil
		},
	})
	if _, err := svc.Reauthorize(context.Background(), "conn-1"); err != nil {
		t.Fatal(err)
	}
	if gotName != "connectors/conn-1" {
		t.Errorf("Name = %q, want connectors/conn-1", gotName)
	}
}

func TestConnectorsService_Activate_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newConnectorSvc(&mockConnectorClient{
		activateConnector: func(_ context.Context, req *connect.Request[outboxv1.ActivateConnectorRequest]) (*connect.Response[outboxv1.ActivateConnectorResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.ActivateConnectorResponse{
				Connector: &outboxv1.Connector{Name: req.Msg.Name},
			}), nil
		},
	})
	if _, err := svc.Activate(context.Background(), "c1"); err != nil {
		t.Fatal(err)
	}
	if gotName != "connectors/c1" {
		t.Errorf("Name = %q, want connectors/c1", gotName)
	}
}

func TestConnectorsService_Activate_ReturnsMappedConnector(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		activateConnector: func(_ context.Context, req *connect.Request[outboxv1.ActivateConnectorRequest]) (*connect.Response[outboxv1.ActivateConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.ActivateConnectorResponse{
				Connector: &outboxv1.Connector{
					Name:  "connectors/c1",
					State: outboxv1.ConnectorState_CONNECTOR_STATE_ACTIVE,
				},
			}), nil
		},
	})
	c, err := svc.Activate(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "c1" {
		t.Errorf("ID = %q, want c1", c.ID)
	}
	if c.State != outboxv1.ConnectorState_CONNECTOR_STATE_ACTIVE {
		t.Errorf("State = %v, want CONNECTOR_STATE_ACTIVE", c.State)
	}
}

func TestConnectorsService_Activate_EmptyResponse(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		activateConnector: func(_ context.Context, _ *connect.Request[outboxv1.ActivateConnectorRequest]) (*connect.Response[outboxv1.ActivateConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.ActivateConnectorResponse{}), nil
		},
	})
	if _, err := svc.Activate(context.Background(), "c1"); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestConnectorsService_Deactivate_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newConnectorSvc(&mockConnectorClient{
		deactivateConnector: func(_ context.Context, req *connect.Request[outboxv1.DeactivateConnectorRequest]) (*connect.Response[outboxv1.DeactivateConnectorResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.DeactivateConnectorResponse{
				Connector: &outboxv1.Connector{Name: req.Msg.Name},
			}), nil
		},
	})
	if _, err := svc.Deactivate(context.Background(), "c1"); err != nil {
		t.Fatal(err)
	}
	if gotName != "connectors/c1" {
		t.Errorf("Name = %q, want connectors/c1", gotName)
	}
}

func TestConnectorsService_Deactivate_ReturnsMappedConnector(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		deactivateConnector: func(_ context.Context, _ *connect.Request[outboxv1.DeactivateConnectorRequest]) (*connect.Response[outboxv1.DeactivateConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.DeactivateConnectorResponse{
				Connector: &outboxv1.Connector{
					Name:  "connectors/c1",
					State: outboxv1.ConnectorState_CONNECTOR_STATE_INACTIVE,
				},
			}), nil
		},
	})
	c, err := svc.Deactivate(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "c1" {
		t.Errorf("ID = %q, want c1", c.ID)
	}
}

func TestConnectorsService_Deactivate_EmptyResponse(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		deactivateConnector: func(_ context.Context, _ *connect.Request[outboxv1.DeactivateConnectorRequest]) (*connect.Response[outboxv1.DeactivateConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.DeactivateConnectorResponse{}), nil
		},
	})
	if _, err := svc.Deactivate(context.Background(), "c1"); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestConnectorsService_CreateManaged_ImmediateDone(t *testing.T) {
	connProto := &outboxv1.Connector{Name: "connectors/c1", Kind: outboxv1.ConnectorKind_CONNECTOR_KIND_MANAGED}
	anyConn, _ := anypb.New(connProto)
	svc := &ConnectorsService{
		client: &mockConnectorClient{
			createManagedConnector: func(_ context.Context, _ *connect.Request[outboxv1.CreateManagedConnectorRequest]) (*connect.Response[longrunningpb.Operation], error) {
				return connect.NewResponse(&longrunningpb.Operation{
					Done:   true,
					Result: &longrunningpb.Operation_Response{Response: anyConn},
				}), nil
			},
		},
	}
	c, err := svc.CreateManaged(context.Background(), CreateManagedConnectorInput{Channel: "sms"})
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "c1" {
		t.Errorf("ID = %q, want c1", c.ID)
	}
}

func TestConnectorsService_CreateManaged_PollsUntilDone(t *testing.T) {
	connProto := &outboxv1.Connector{Name: "connectors/c2"}
	anyConn, _ := anypb.New(connProto)
	calls := 0
	svc := &ConnectorsService{
		pollInterval: 0,
		client: &mockConnectorClient{
			createManagedConnector: func(_ context.Context, _ *connect.Request[outboxv1.CreateManagedConnectorRequest]) (*connect.Response[longrunningpb.Operation], error) {
				return connect.NewResponse(&longrunningpb.Operation{Name: "operations/op1", Done: false}), nil
			},
		},
		getOperation: func(_ context.Context, req *connect.Request[longrunningpb.GetOperationRequest]) (*connect.Response[longrunningpb.Operation], error) {
			calls++
			if req.Msg.Name != "operations/op1" {
				t.Errorf("getOperation Name = %q, want operations/op1", req.Msg.Name)
			}
			if calls < 2 {
				return connect.NewResponse(&longrunningpb.Operation{Name: "operations/op1", Done: false}), nil
			}
			return connect.NewResponse(&longrunningpb.Operation{
				Done:   true,
				Result: &longrunningpb.Operation_Response{Response: anyConn},
			}), nil
		},
	}
	c, err := svc.CreateManaged(context.Background(), CreateManagedConnectorInput{Channel: "sms"})
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "c2" {
		t.Errorf("ID = %q, want c2", c.ID)
	}
}

func TestConnectorsService_CreateManaged_GetOperationError(t *testing.T) {
	svc := &ConnectorsService{
		pollInterval: 0,
		client: &mockConnectorClient{
			createManagedConnector: func(_ context.Context, _ *connect.Request[outboxv1.CreateManagedConnectorRequest]) (*connect.Response[longrunningpb.Operation], error) {
				return connect.NewResponse(&longrunningpb.Operation{Done: false, Name: "operations/op1"}), nil
			},
		},
		getOperation: func(_ context.Context, _ *connect.Request[longrunningpb.GetOperationRequest]) (*connect.Response[longrunningpb.Operation], error) {
			return nil, errors.New("network error")
		},
	}
	_, err := svc.CreateManaged(context.Background(), CreateManagedConnectorInput{Channel: "sms"})
	if err == nil {
		t.Error("expected error when getOperation fails, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "network error") {
		t.Errorf("error = %v, want to contain 'network error'", err)
	}
}

func TestConnectorsService_CreateManaged_OperationError(t *testing.T) {
	svc := &ConnectorsService{
		client: &mockConnectorClient{
			createManagedConnector: func(_ context.Context, _ *connect.Request[outboxv1.CreateManagedConnectorRequest]) (*connect.Response[longrunningpb.Operation], error) {
				return connect.NewResponse(&longrunningpb.Operation{
					Done:   true,
					Result: &longrunningpb.Operation_Error{Error: &statuspb.Status{Code: 13, Message: "internal error"}},
				}), nil
			},
		},
	}
	if _, err := svc.CreateManaged(context.Background(), CreateManagedConnectorInput{Channel: "sms"}); err == nil {
		t.Error("expected error for failed operation, got nil")
	}
}

func TestConnectorsService_CreateManaged_NilResponse(t *testing.T) {
	svc := &ConnectorsService{
		client: &mockConnectorClient{
			createManagedConnector: func(_ context.Context, _ *connect.Request[outboxv1.CreateManagedConnectorRequest]) (*connect.Response[longrunningpb.Operation], error) {
				return connect.NewResponse(&longrunningpb.Operation{
					Done: true,
					// No Result set — simulates Done: true with neither error nor response
				}), nil
			},
		},
	}
	_, err := svc.CreateManaged(context.Background(), CreateManagedConnectorInput{Channel: "sms"})
	if err == nil {
		t.Error("expected error for nil response, got nil")
	}
}

func TestMapConnector_TagsCopied(t *testing.T) {
	p := &outboxv1.Connector{
		Name: "connectors/c1",
		Tags: []string{"a", "b"},
	}
	c := mapConnector(p)
	// Mutating the proto tags must not affect the mapped connector's tags.
	p.Tags[0] = "MUTATED"
	if c.Tags[0] != "a" {
		t.Errorf("Tags[0] = %q after proto mutation, want a (expected defensive copy)", c.Tags[0])
	}
}

func TestConnectorsService_Verify_BuildsResourceName(t *testing.T) {
	var gotReq *outboxv1.VerifyConnectorRequest
	svc := newConnectorSvc(&mockConnectorClient{
		verifyConnector: func(_ context.Context, req *connect.Request[outboxv1.VerifyConnectorRequest]) (*connect.Response[outboxv1.VerifyConnectorResponse], error) {
			gotReq = req.Msg
			return connect.NewResponse(&outboxv1.VerifyConnectorResponse{
				Connector: &outboxv1.Connector{Name: req.Msg.Name},
			}), nil
		},
	})
	if _, err := svc.Verify(context.Background(), "c1", "123456", ""); err != nil {
		t.Fatal(err)
	}
	if gotReq.Name != "connectors/c1" {
		t.Errorf("Name = %q, want connectors/c1", gotReq.Name)
	}
	if gotReq.Code != "123456" {
		t.Errorf("Code = %q, want 123456", gotReq.Code)
	}
}

func TestConnectorsService_Verify_PassesPassword(t *testing.T) {
	var gotPw string
	svc := newConnectorSvc(&mockConnectorClient{
		verifyConnector: func(_ context.Context, req *connect.Request[outboxv1.VerifyConnectorRequest]) (*connect.Response[outboxv1.VerifyConnectorResponse], error) {
			gotPw = req.Msg.Password
			return connect.NewResponse(&outboxv1.VerifyConnectorResponse{
				Connector: &outboxv1.Connector{Name: "connectors/c1"},
			}), nil
		},
	})
	if _, err := svc.Verify(context.Background(), "c1", "123456", "cloudpassword"); err != nil {
		t.Fatal(err)
	}
	if gotPw != "cloudpassword" {
		t.Errorf("Password = %q, want cloudpassword", gotPw)
	}
}

func TestConnectorsService_Verify_EmptyResponse(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		verifyConnector: func(_ context.Context, _ *connect.Request[outboxv1.VerifyConnectorRequest]) (*connect.Response[outboxv1.VerifyConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.VerifyConnectorResponse{}), nil
		},
	})
	if _, err := svc.Verify(context.Background(), "c1", "code", ""); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestConnectorsService_Detach_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newConnectorSvc(&mockConnectorClient{
		detachProvisionedResource: func(_ context.Context, req *connect.Request[outboxv1.DetachProvisionedResourceRequest]) (*connect.Response[outboxv1.DetachProvisionedResourceResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.DetachProvisionedResourceResponse{
				Connector: &outboxv1.Connector{Name: req.Msg.Name},
			}), nil
		},
	})
	if _, err := svc.Detach(context.Background(), "c1"); err != nil {
		t.Fatal(err)
	}
	if gotName != "connectors/c1" {
		t.Errorf("Name = %q, want connectors/c1", gotName)
	}
}

func TestConnectorsService_Detach_EmptyResponse(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		detachProvisionedResource: func(_ context.Context, _ *connect.Request[outboxv1.DetachProvisionedResourceRequest]) (*connect.Response[outboxv1.DetachProvisionedResourceResponse], error) {
			return connect.NewResponse(&outboxv1.DetachProvisionedResourceResponse{}), nil
		},
	})
	if _, err := svc.Detach(context.Background(), "c1"); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestMapConnector_NewFields(t *testing.T) {
	p := &outboxv1.Connector{
		Name:                 "connectors/c1",
		Kind:                 outboxv1.ConnectorKind_CONNECTOR_KIND_MANAGED,
		Readiness:            outboxv1.ConnectorReadiness_CONNECTOR_READINESS_READY,
		WebhookUrl:           "https://example.com/webhook",
		DisplayName:          "Alice",
		ProvisionedResources: []string{"provisioned_resources/pr1"},
	}
	c := mapConnector(p)
	if c.Kind != ConnectorKindManaged {
		t.Errorf("Kind = %v, want Managed", c.Kind)
	}
	if c.Readiness != ConnectorReadinessReady {
		t.Errorf("Readiness = %v, want Ready", c.Readiness)
	}
	if c.WebhookURL != "https://example.com/webhook" {
		t.Errorf("WebhookURL = %q", c.WebhookURL)
	}
	if c.DisplayName != "Alice" {
		t.Errorf("DisplayName = %q", c.DisplayName)
	}
	if len(c.ProvisionedResources) != 1 || c.ProvisionedResources[0] != "pr1" {
		t.Errorf("ProvisionedResources = %v, want [pr1]", c.ProvisionedResources)
	}
}

func TestConnectorsService_Create_ConsentAcknowledged(t *testing.T) {
	var got *outboxv1.CreateConnectorRequest
	svc := newConnectorSvc(&mockConnectorClient{
		createConnector: func(_ context.Context, req *connect.Request[outboxv1.CreateConnectorRequest]) (*connect.Response[outboxv1.CreateConnectorResponse], error) {
			got = req.Msg
			return okConnectorResp("c1"), nil
		},
	})
	if _, err := svc.Create(context.Background(), CreateConnectorInput{ConsentAcknowledged: true}); err != nil {
		t.Fatal(err)
	}
	if !got.ConsentAcknowledged {
		t.Error("expected ConsentAcknowledged=true, got false")
	}
}

func TestConnectorsService_Reauthorize_EmptyResponse(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		reauthorizeConnector: func(_ context.Context, _ *connect.Request[outboxv1.ReauthorizeConnectorRequest]) (*connect.Response[outboxv1.ReauthorizeConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.ReauthorizeConnectorResponse{}), nil
		},
	})
	if _, err := svc.Reauthorize(context.Background(), "c1"); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestConnectorsService_CreateManaged_Fields(t *testing.T) {
	connProto := &outboxv1.Connector{Name: "connectors/c1"}
	anyConn, _ := anypb.New(connProto)
	var got *outboxv1.CreateManagedConnectorRequest
	svc := &ConnectorsService{
		client: &mockConnectorClient{
			createManagedConnector: func(_ context.Context, req *connect.Request[outboxv1.CreateManagedConnectorRequest]) (*connect.Response[longrunningpb.Operation], error) {
				got = req.Msg
				return connect.NewResponse(&longrunningpb.Operation{
					Done:   true,
					Result: &longrunningpb.Operation_Response{Response: anyConn},
				}), nil
			},
		},
	}
	_, err := svc.CreateManaged(context.Background(), CreateManagedConnectorInput{
		Channel:    "sms",
		Filters:    map[string]string{"country": "US"},
		WebhookURL: "https://example.com/hook",
		Tags:       []string{"tag1", "tag2"},
		RequestID:  "req-123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Channel != "sms" {
		t.Errorf("Channel = %q, want sms", got.Channel)
	}
	if got.Filters["country"] != "US" {
		t.Errorf("Filters[country] = %q, want US", got.Filters["country"])
	}
	if got.WebhookUrl != "https://example.com/hook" {
		t.Errorf("WebhookUrl = %q, want https://example.com/hook", got.WebhookUrl)
	}
	if !slices.Equal(got.Tags, []string{"tag1", "tag2"}) {
		t.Errorf("Tags = %v, want [tag1 tag2]", got.Tags)
	}
	if got.RequestId != "req-123" {
		t.Errorf("RequestId = %q, want req-123", got.RequestId)
	}
}

func TestConnectorsService_CreateManaged_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel so ctx.Err() returns non-nil immediately
	svc := &ConnectorsService{
		pollInterval: 0,
		client: &mockConnectorClient{
			createManagedConnector: func(_ context.Context, _ *connect.Request[outboxv1.CreateManagedConnectorRequest]) (*connect.Response[longrunningpb.Operation], error) {
				return connect.NewResponse(&longrunningpb.Operation{Done: false, Name: "operations/op1"}), nil
			},
		},
		getOperation: func(_ context.Context, _ *connect.Request[longrunningpb.GetOperationRequest]) (*connect.Response[longrunningpb.Operation], error) {
			return connect.NewResponse(&longrunningpb.Operation{Done: false}), nil
		},
	}
	_, err := svc.CreateManaged(ctx, CreateManagedConnectorInput{Channel: "sms"})
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}

func TestConnectorsService_CreateManaged_ContextCancellationWithInterval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel
	svc := &ConnectorsService{
		pollInterval: time.Nanosecond, // non-zero: exercises the select { ctx.Done(), time.After } branch
		client: &mockConnectorClient{
			createManagedConnector: func(_ context.Context, _ *connect.Request[outboxv1.CreateManagedConnectorRequest]) (*connect.Response[longrunningpb.Operation], error) {
				return connect.NewResponse(&longrunningpb.Operation{Done: false, Name: "operations/op1"}), nil
			},
		},
		getOperation: func(_ context.Context, _ *connect.Request[longrunningpb.GetOperationRequest]) (*connect.Response[longrunningpb.Operation], error) {
			return connect.NewResponse(&longrunningpb.Operation{Done: false}), nil
		},
	}
	_, err := svc.CreateManaged(ctx, CreateManagedConnectorInput{Channel: "sms"})
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}

func TestMapConnector_EmptyProvisionedResources(t *testing.T) {
	c := mapConnector(&outboxv1.Connector{Name: "connectors/c1"})
	if c.ProvisionedResources == nil {
		t.Error("expected non-nil empty slice, got nil")
	}
}

func TestConnectorsService_List_MapsResult(t *testing.T) {
	svc := newConnectorSvc(&mockConnectorClient{
		listConnectors: func(_ context.Context, _ *connect.Request[outboxv1.ListConnectorsRequest]) (*connect.Response[outboxv1.ListConnectorsResponse], error) {
			return connect.NewResponse(&outboxv1.ListConnectorsResponse{
				Connectors:    []*outboxv1.Connector{{Name: "connectors/c1"}},
				NextPageToken: "tok1",
				TotalSize:     42,
			}), nil
		},
	})
	res, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Items) != 1 || res.Items[0].ID != "c1" {
		t.Errorf("Items = %v, want [{ID: c1}]", res.Items)
	}
	if res.NextPageToken != "tok1" {
		t.Errorf("NextPageToken = %q, want tok1", res.NextPageToken)
	}
	if res.TotalSize != 42 {
		t.Errorf("TotalSize = %d, want 42", res.TotalSize)
	}
}
