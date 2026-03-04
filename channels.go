package outbox

import (
	"context"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	outboxv1connect "github.com/getoutbox/outbox-go/outboxv1/outboxv1connect"
)

// ChannelsService provides read-only access to Channels.
type ChannelsService struct {
	client outboxv1connect.ChannelServiceClient
}

// ListChannelsOptions configures a List request.
type ListChannelsOptions struct {
	PageSize  int32
	PageToken string
}

// ListChannelsResult is the paginated result of a List call.
type ListChannelsResult struct {
	Items         []Channel
	NextPageToken string
	TotalSize     int64
}

// Get retrieves a Channel by its ID.
func (s *ChannelsService) Get(ctx context.Context, id string) (*Channel, error) {
	res, err := s.client.GetChannel(ctx, connect.NewRequest(&outboxv1.GetChannelRequest{
		Name: "channels/" + id,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Channel == nil {
		return nil, errEmpty("GetChannel")
	}
	ch := mapChannel(res.Msg.Channel)
	return &ch, nil
}

// List returns a paginated list of Channels.
func (s *ChannelsService) List(ctx context.Context, opts *ListChannelsOptions) (*ListChannelsResult, error) {
	r := &outboxv1.ListChannelsRequest{}
	if opts != nil {
		r.PageSize = opts.PageSize
		r.PageToken = opts.PageToken
	}
	res, err := s.client.ListChannels(ctx, connect.NewRequest(r))
	if err != nil {
		return nil, err
	}
	items := make([]Channel, len(res.Msg.Channels))
	for i, c := range res.Msg.Channels {
		items[i] = mapChannel(c)
	}
	return &ListChannelsResult{
		Items:         items,
		NextPageToken: res.Msg.NextPageToken,
		TotalSize:     int64(res.Msg.TotalSize),
	}, nil
}

func mapChannel(p *outboxv1.Channel) Channel {
	ch := Channel{
		ID:         ParseID(p.GetName()),
		CreateTime: protoTime(p.GetCreateTime()),
	}
	if caps := p.GetCapabilities(); caps != nil {
		ch.Capabilities = &ChannelCapabilities{
			Groups:                caps.GetGroups(),
			Reactions:             caps.GetReactions(),
			Edits:                 caps.GetEdits(),
			Deletions:             caps.GetDeletions(),
			ReadReceipts:          caps.GetReadReceipts(),
			TypingIndicators:      caps.GetTypingIndicators(),
			SupportedContentTypes: caps.GetSupportedContentTypes(),
		}
	}
	return ch
}
