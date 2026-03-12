package outbox

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"
	outboxv1connect "github.com/getoutbox/outbox-go/gen/outbox/v1/outboxv1connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newTemplateSvc(m outboxv1connect.TemplateServiceClient) *TemplatesService {
	return &TemplatesService{client: m}
}

func TestTemplatesService_Create_BuildsParent(t *testing.T) {
	var gotParent string
	svc := newTemplateSvc(&mockTemplateClient{
		createTemplate: func(_ context.Context, req *connect.Request[outboxv1.CreateTemplateRequest]) (*connect.Response[outboxv1.CreateTemplateResponse], error) {
			gotParent = req.Msg.Parent
			return connect.NewResponse(&outboxv1.CreateTemplateResponse{
				Template: &outboxv1.Template{Name: "connectors/c1/templates/t1"},
			}), nil
		},
	})
	if _, err := svc.Create(context.Background(), "c1", CreateTemplateInput{
		TemplateName:   "hello_world",
		Language:       "en",
		Category:       TemplateCategoryUtility,
		ComponentsJSON: "[]",
	}); err != nil {
		t.Fatal(err)
	}
	if gotParent != "connectors/c1" {
		t.Errorf("Parent = %q, want connectors/c1", gotParent)
	}
}

func TestTemplatesService_Create_PropagatesFields(t *testing.T) {
	var got *outboxv1.CreateTemplateRequest
	svc := newTemplateSvc(&mockTemplateClient{
		createTemplate: func(_ context.Context, req *connect.Request[outboxv1.CreateTemplateRequest]) (*connect.Response[outboxv1.CreateTemplateResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.CreateTemplateResponse{
				Template: &outboxv1.Template{Name: "connectors/c1/templates/t1"},
			}), nil
		},
	})
	if _, err := svc.Create(context.Background(), "c1", CreateTemplateInput{
		TemplateName:   "hello_world",
		Language:       "en_US",
		Category:       TemplateCategoryMarketing,
		ComponentsJSON: `[{"type":"BODY"}]`,
	}); err != nil {
		t.Fatal(err)
	}
	if got.Template.TemplateName != "hello_world" {
		t.Errorf("TemplateName = %q, want hello_world", got.Template.TemplateName)
	}
	if got.Template.Language != "en_US" {
		t.Errorf("Language = %q, want en_US", got.Template.Language)
	}
	if got.Template.Category != outboxv1.Template_CATEGORY_MARKETING {
		t.Errorf("Category = %v, want CATEGORY_MARKETING", got.Template.Category)
	}
	if got.Template.ComponentsJson != `[{"type":"BODY"}]` {
		t.Errorf("ComponentsJson = %q, want [{\"type\":\"BODY\"}]", got.Template.ComponentsJson)
	}
}

func TestTemplatesService_Create_EmptyResponse(t *testing.T) {
	svc := newTemplateSvc(&mockTemplateClient{
		createTemplate: func(_ context.Context, _ *connect.Request[outboxv1.CreateTemplateRequest]) (*connect.Response[outboxv1.CreateTemplateResponse], error) {
			return connect.NewResponse(&outboxv1.CreateTemplateResponse{}), nil
		},
	})
	if _, err := svc.Create(context.Background(), "c1", CreateTemplateInput{}); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestTemplatesService_Get_BuildsName(t *testing.T) {
	var gotName string
	svc := newTemplateSvc(&mockTemplateClient{
		getTemplate: func(_ context.Context, req *connect.Request[outboxv1.GetTemplateRequest]) (*connect.Response[outboxv1.GetTemplateResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.GetTemplateResponse{
				Template: &outboxv1.Template{Name: req.Msg.Name},
			}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "c1", "t1"); err != nil {
		t.Fatal(err)
	}
	if gotName != "connectors/c1/templates/t1" {
		t.Errorf("Name = %q, want connectors/c1/templates/t1", gotName)
	}
}

func TestTemplatesService_List_BuildsParent(t *testing.T) {
	var gotParent string
	svc := newTemplateSvc(&mockTemplateClient{
		listTemplates: func(_ context.Context, req *connect.Request[outboxv1.ListTemplatesRequest]) (*connect.Response[outboxv1.ListTemplatesResponse], error) {
			gotParent = req.Msg.Parent
			return connect.NewResponse(&outboxv1.ListTemplatesResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), "c1", nil); err != nil {
		t.Fatal(err)
	}
	if gotParent != "connectors/c1" {
		t.Errorf("Parent = %q, want connectors/c1", gotParent)
	}
}

func TestTemplatesService_Delete_BuildsName(t *testing.T) {
	var gotName string
	svc := newTemplateSvc(&mockTemplateClient{
		deleteTemplate: func(_ context.Context, req *connect.Request[outboxv1.DeleteTemplateRequest]) (*connect.Response[outboxv1.DeleteTemplateResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.DeleteTemplateResponse{}), nil
		},
	})
	if err := svc.Delete(context.Background(), "c1", "t1"); err != nil {
		t.Fatal(err)
	}
	if gotName != "connectors/c1/templates/t1" {
		t.Errorf("Name = %q, want connectors/c1/templates/t1", gotName)
	}
}

func TestMapTemplate_ParsesIDs(t *testing.T) {
	p := &outboxv1.Template{
		Name:            "connectors/c1/templates/t1",
		TemplateName:    "hello",
		Language:        "en",
		Category:        outboxv1.Template_CATEGORY_UTILITY,
		ComponentsJson:  "[]",
		Status:          outboxv1.TemplateStatus_TEMPLATE_STATUS_APPROVED,
		RejectionReason: "",
		ExternalId:      "ext-1",
	}
	tmpl := mapTemplate(p)
	if tmpl.ID != "t1" {
		t.Errorf("ID = %q, want t1", tmpl.ID)
	}
	if tmpl.ConnectorID != "c1" {
		t.Errorf("ConnectorID = %q, want c1", tmpl.ConnectorID)
	}
	if tmpl.Status != TemplateStatusApproved {
		t.Errorf("Status = %v, want Approved", tmpl.Status)
	}
}

func TestMapTemplate_AllFields(t *testing.T) {
	ts := time.Unix(1700000000, 0).UTC()
	p := &outboxv1.Template{
		Name:            "connectors/c1/templates/t1",
		TemplateName:    "hello_world",
		Language:        "en_US",
		Category:        outboxv1.Template_CATEGORY_MARKETING,
		ComponentsJson:  `[{"type":"BODY"}]`,
		Status:          outboxv1.TemplateStatus_TEMPLATE_STATUS_PENDING,
		RejectionReason: "content policy violation",
		ExternalId:      "whatsapp-ext-123",
		CreateTime:      timestamppb.New(ts),
		UpdateTime:      timestamppb.New(ts.Add(time.Hour)),
	}
	tmpl := mapTemplate(p)
	if tmpl.TemplateName != "hello_world" {
		t.Errorf("TemplateName = %q, want hello_world", tmpl.TemplateName)
	}
	if tmpl.Language != "en_US" {
		t.Errorf("Language = %q, want en_US", tmpl.Language)
	}
	if tmpl.Category != TemplateCategoryMarketing {
		t.Errorf("Category = %v, want Marketing", tmpl.Category)
	}
	if tmpl.ComponentsJSON != `[{"type":"BODY"}]` {
		t.Errorf("ComponentsJSON = %q, want [{\"type\":\"BODY\"}]", tmpl.ComponentsJSON)
	}
	if tmpl.RejectionReason != "content policy violation" {
		t.Errorf("RejectionReason = %q, want 'content policy violation'", tmpl.RejectionReason)
	}
	if tmpl.ExternalID != "whatsapp-ext-123" {
		t.Errorf("ExternalID = %q, want whatsapp-ext-123", tmpl.ExternalID)
	}
	if !tmpl.CreateTime.Equal(ts) {
		t.Errorf("CreateTime = %v, want %v", tmpl.CreateTime, ts)
	}
	if !tmpl.UpdateTime.Equal(ts.Add(time.Hour)) {
		t.Errorf("UpdateTime = %v, want %v", tmpl.UpdateTime, ts.Add(time.Hour))
	}
}

func TestTemplatesService_Get_EmptyResponse(t *testing.T) {
	svc := newTemplateSvc(&mockTemplateClient{
		getTemplate: func(_ context.Context, _ *connect.Request[outboxv1.GetTemplateRequest]) (*connect.Response[outboxv1.GetTemplateResponse], error) {
			return connect.NewResponse(&outboxv1.GetTemplateResponse{}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "c1", "t1"); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestTemplatesService_List_PropagatesOptions(t *testing.T) {
	var got *outboxv1.ListTemplatesRequest
	svc := newTemplateSvc(&mockTemplateClient{
		listTemplates: func(_ context.Context, req *connect.Request[outboxv1.ListTemplatesRequest]) (*connect.Response[outboxv1.ListTemplatesResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.ListTemplatesResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), "c1", &ListTemplatesOptions{
		PageSize:  5,
		PageToken: "tok",
	}); err != nil {
		t.Fatal(err)
	}
	if got.PageSize != 5 {
		t.Errorf("PageSize = %d, want 5", got.PageSize)
	}
	if got.PageToken != "tok" {
		t.Errorf("PageToken = %q, want tok", got.PageToken)
	}
}

func TestTemplatesService_List_NilOptions(t *testing.T) {
	svc := newTemplateSvc(&mockTemplateClient{
		listTemplates: func(_ context.Context, _ *connect.Request[outboxv1.ListTemplatesRequest]) (*connect.Response[outboxv1.ListTemplatesResponse], error) {
			return connect.NewResponse(&outboxv1.ListTemplatesResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), "c1", nil); err != nil {
		t.Fatal(err)
	}
}

func TestMapTemplate_MalformedName(t *testing.T) {
	p := &outboxv1.Template{Name: "bad-name"}
	tmpl := mapTemplate(p)
	if tmpl.ConnectorID != "" {
		t.Errorf("ConnectorID = %q, want empty for malformed name", tmpl.ConnectorID)
	}
	if tmpl.ID != "bad-name" {
		t.Errorf("ID = %q, want bad-name (last segment fallback)", tmpl.ID)
	}
}

func TestTemplatesService_List_MapsResult(t *testing.T) {
	svc := newTemplateSvc(&mockTemplateClient{
		listTemplates: func(_ context.Context, _ *connect.Request[outboxv1.ListTemplatesRequest]) (*connect.Response[outboxv1.ListTemplatesResponse], error) {
			return connect.NewResponse(&outboxv1.ListTemplatesResponse{
				Templates:     []*outboxv1.Template{{Name: "connectors/c1/templates/t1"}},
				NextPageToken: "tok1",
				TotalSize:     7,
			}), nil
		},
	})
	res, err := svc.List(context.Background(), "c1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Items) != 1 || res.Items[0].ID != "t1" {
		t.Errorf("Items = %v, want [{ID: t1}]", res.Items)
	}
	if res.NextPageToken != "tok1" {
		t.Errorf("NextPageToken = %q, want tok1", res.NextPageToken)
	}
	if res.TotalSize != 7 {
		t.Errorf("TotalSize = %d, want 7", res.TotalSize)
	}
}
