package outbox

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"
	outboxv1connect "github.com/getoutbox/outbox-go/gen/outbox/v1/outboxv1connect"
)

// TemplatesService provides operations on Templates.
// Templates are scoped to a Connector and immutable after creation.
type TemplatesService struct {
	client outboxv1connect.TemplateServiceClient
}

// CreateTemplateInput holds parameters for Templates.Create.
type CreateTemplateInput struct {
	TemplateName   string
	Language       string
	Category       TemplateCategory
	ComponentsJSON string
}

// ListTemplatesOptions configures a List request.
type ListTemplatesOptions struct {
	PageSize  int32
	PageToken string
}

// Create creates a new Template scoped to the given Connector.
func (s *TemplatesService) Create(ctx context.Context, connectorID string, input CreateTemplateInput) (*Template, error) {
	res, err := s.client.CreateTemplate(ctx, connect.NewRequest(&outboxv1.CreateTemplateRequest{
		Parent: "connectors/" + connectorID,
		Template: &outboxv1.Template{
			TemplateName:   input.TemplateName,
			Language:       input.Language,
			Category:       input.Category,
			ComponentsJson: input.ComponentsJSON,
		},
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Template == nil {
		return nil, errEmpty("CreateTemplate")
	}
	tmpl := mapTemplate(res.Msg.Template)
	return &tmpl, nil
}

// Get retrieves a Template by Connector ID and Template ID.
func (s *TemplatesService) Get(ctx context.Context, connectorID, id string) (*Template, error) {
	res, err := s.client.GetTemplate(ctx, connect.NewRequest(&outboxv1.GetTemplateRequest{
		Name: "connectors/" + connectorID + "/templates/" + id,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Template == nil {
		return nil, errEmpty("GetTemplate")
	}
	tmpl := mapTemplate(res.Msg.Template)
	return &tmpl, nil
}

// List returns a paginated list of Templates for the given Connector.
func (s *TemplatesService) List(ctx context.Context, connectorID string, opts *ListTemplatesOptions) (*ListTemplatesResult, error) {
	r := &outboxv1.ListTemplatesRequest{
		Parent: "connectors/" + connectorID,
	}
	if opts != nil {
		r.PageSize = opts.PageSize
		r.PageToken = opts.PageToken
	}
	res, err := s.client.ListTemplates(ctx, connect.NewRequest(r))
	if err != nil {
		return nil, err
	}
	items := make([]Template, len(res.Msg.Templates))
	for i, t := range res.Msg.Templates {
		items[i] = mapTemplate(t)
	}
	return &ListTemplatesResult{
		Items:         items,
		NextPageToken: res.Msg.NextPageToken,
		TotalSize:     int64(res.Msg.TotalSize),
	}, nil
}

// Delete deletes a Template by Connector ID and Template ID.
func (s *TemplatesService) Delete(ctx context.Context, connectorID, id string) error {
	_, err := s.client.DeleteTemplate(ctx, connect.NewRequest(&outboxv1.DeleteTemplateRequest{
		Name: "connectors/" + connectorID + "/templates/" + id,
	}))
	return err
}

func mapTemplate(p *outboxv1.Template) Template {
	name := p.GetName()
	templateID := ParseID(name)
	connectorID := ""
	if suffix := "/templates/" + templateID; strings.HasSuffix(name, suffix) {
		connectorID = ParseID(name[:len(name)-len(suffix)])
	}
	return Template{
		ID:              templateID,
		ConnectorID:     connectorID,
		TemplateName:    p.GetTemplateName(),
		Language:        p.GetLanguage(),
		Category:        p.GetCategory(),
		ComponentsJSON:  p.GetComponentsJson(),
		Status:          p.GetStatus(),
		RejectionReason: p.GetRejectionReason(),
		ExternalID:      p.GetExternalId(),
		CreateTime:      protoTime(p.GetCreateTime()),
		UpdateTime:      protoTime(p.GetUpdateTime()),
	}
}
