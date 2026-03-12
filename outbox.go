package outbox

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	longrunningpb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	outboxv1connect "github.com/getoutbox/outbox-go/gen/outbox/v1/outboxv1connect"
)

const defaultBaseURL = "https://api.outbox.chat"

type clientConfig struct {
	baseURL    string
	httpClient *http.Client
}

// Option configures the Outbox client.
type Option func(*clientConfig)

// WithBaseURL overrides the Outbox API base URL. Defaults to https://api.outbox.chat.
func WithBaseURL(url string) Option {
	return func(c *clientConfig) { c.baseURL = url }
}

// WithHTTPClient sets a custom *http.Client for all requests.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *clientConfig) { c.httpClient = hc }
}

// Client is the Outbox API client. Create one with New.
type Client struct {
	// Accounts provides operations on Accounts (end-user identities on a Channel).
	Accounts *AccountsService
	// Connectors provides operations on Connectors, which bridge an Account to a Channel.
	Connectors *ConnectorsService
	// Destinations provides operations on Destinations, which receive webhook delivery events.
	Destinations *DestinationsService
	// Messages provides operations for sending, updating, and receiving Messages.
	Messages *MessagesService
	// Templates provides operations on message Templates (e.g. WhatsApp templates).
	Templates *TemplatesService
}

// New creates a new Outbox client authenticated with the given API key.
// API keys look like ob_live_... or ob_test_...
func New(apiKey string, opts ...Option) *Client {
	cfg := &clientConfig{
		baseURL:    defaultBaseURL,
		httpClient: http.DefaultClient,
	}
	for _, o := range opts {
		o(cfg)
	}
	interceptor := connect.WithInterceptors(newBearerAuthInterceptor(apiKey))
	opsClient := connect.NewClient[longrunningpb.GetOperationRequest, longrunningpb.Operation](
		cfg.httpClient,
		cfg.baseURL+"/google.longrunning.Operations/GetOperation",
		interceptor,
	)
	return &Client{
		Accounts: &AccountsService{client: outboxv1connect.NewAccountServiceClient(cfg.httpClient, cfg.baseURL, interceptor)},
		Connectors: &ConnectorsService{
			client: outboxv1connect.NewConnectorServiceClient(cfg.httpClient, cfg.baseURL, interceptor),
			getOperation: func(ctx context.Context, req *connect.Request[longrunningpb.GetOperationRequest]) (*connect.Response[longrunningpb.Operation], error) {
				return opsClient.CallUnary(ctx, req)
			},
			pollInterval: 2 * time.Second,
		},
		Destinations: &DestinationsService{client: outboxv1connect.NewDestinationServiceClient(cfg.httpClient, cfg.baseURL, interceptor)},
		Messages:     &MessagesService{client: outboxv1connect.NewMessageServiceClient(cfg.httpClient, cfg.baseURL, interceptor)},
		Templates:    &TemplatesService{client: outboxv1connect.NewTemplateServiceClient(cfg.httpClient, cfg.baseURL, interceptor)},
	}
}

func newBearerAuthInterceptor(apiKey string) connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+apiKey)
			return next(ctx, req)
		})
	})
}
