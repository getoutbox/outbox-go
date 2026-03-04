package outbox

import (
	"context"
	"slices"
	"testing"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
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
		ChannelConfig: &outboxv1.Connector_Slack{Slack: &outboxv1.SlackConfig{BotToken: "xoxb"}},
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if len(got.UpdateMask.Paths) != 1 || got.UpdateMask.Paths[0] != "slack" {
		t.Errorf("paths = %v, want [slack]", got.UpdateMask.Paths)
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
		ChannelConfig: &outboxv1.Connector_Slack{Slack: &outboxv1.SlackConfig{}},
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if !slices.Contains(got.UpdateMask.Paths, "tags") {
		t.Errorf("paths %v missing 'tags'", got.UpdateMask.Paths)
	}
	if !slices.Contains(got.UpdateMask.Paths, "slack") {
		t.Errorf("paths %v missing 'slack'", got.UpdateMask.Paths)
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
		reauthorizeConnector: func(_ context.Context, req *connect.Request[outboxv1.ReauthorizeConnectorRequest]) (*connect.Response[outboxv1.ReauthorizeConnectorResponse], error) {
			return connect.NewResponse(&outboxv1.ReauthorizeConnectorResponse{
				AuthorizationUrl: "https://oauth.example.com/reauth",
			}), nil
		},
	})
	url, err := svc.Reauthorize(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://oauth.example.com/reauth" {
		t.Errorf("URL = %q, want https://oauth.example.com/reauth", url)
	}
}

func TestConnectorsService_Reauthorize_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newConnectorSvc(&mockConnectorClient{
		reauthorizeConnector: func(_ context.Context, req *connect.Request[outboxv1.ReauthorizeConnectorRequest]) (*connect.Response[outboxv1.ReauthorizeConnectorResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.ReauthorizeConnectorResponse{}), nil
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
					State: outboxv1.Connector_STATE_ACTIVE,
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
	if c.State != outboxv1.Connector_STATE_ACTIVE {
		t.Errorf("State = %v, want STATE_ACTIVE", c.State)
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
					State: outboxv1.Connector_STATE_INACTIVE,
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
