package outbox

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	outboxv1connect "github.com/getoutbox/outbox-go/outboxv1/outboxv1connect"
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
	// Channels provides access to the available messaging Channels (e.g. WhatsApp, Slack).
	Channels *ChannelsService
	// Connectors provides operations on Connectors, which bridge an Account to a Channel.
	Connectors *ConnectorsService
	// Destinations provides operations on Destinations, which receive webhook delivery events.
	Destinations *DestinationsService
	// Messages provides operations for sending, updating, and receiving Messages.
	Messages *MessagesService
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
	return &Client{
		Accounts:     &AccountsService{client: outboxv1connect.NewAccountServiceClient(cfg.httpClient, cfg.baseURL, interceptor)},
		Channels:     &ChannelsService{client: outboxv1connect.NewChannelServiceClient(cfg.httpClient, cfg.baseURL, interceptor)},
		Connectors:   &ConnectorsService{client: outboxv1connect.NewConnectorServiceClient(cfg.httpClient, cfg.baseURL, interceptor)},
		Destinations: &DestinationsService{client: outboxv1connect.NewDestinationServiceClient(cfg.httpClient, cfg.baseURL, interceptor)},
		Messages:     &MessagesService{client: outboxv1connect.NewMessageServiceClient(cfg.httpClient, cfg.baseURL, interceptor)},
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
