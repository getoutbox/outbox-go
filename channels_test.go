package outbox

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
)

func newChannelSvc(m *mockChannelClient) *ChannelsService {
	return &ChannelsService{client: m}
}

func TestChannelsService_Get_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newChannelSvc(&mockChannelClient{
		getChannel: func(_ context.Context, req *connect.Request[outboxv1.GetChannelRequest]) (*connect.Response[outboxv1.GetChannelResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.GetChannelResponse{
				Channel: &outboxv1.Channel{Name: req.Msg.Name},
			}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "whatsapp"); err != nil {
		t.Fatal(err)
	}
	if gotName != "channels/whatsapp" {
		t.Errorf("Name = %q, want channels/whatsapp", gotName)
	}
}

func TestChannelsService_Get_ReturnsMappedChannel(t *testing.T) {
	svc := newChannelSvc(&mockChannelClient{
		getChannel: func(_ context.Context, _ *connect.Request[outboxv1.GetChannelRequest]) (*connect.Response[outboxv1.GetChannelResponse], error) {
			return connect.NewResponse(&outboxv1.GetChannelResponse{
				Channel: &outboxv1.Channel{
					Name: "channels/slack",
					Capabilities: &outboxv1.Channel_Capabilities{
						Groups:    true,
						Reactions: true,
					},
				},
			}), nil
		},
	})
	ch, err := svc.Get(context.Background(), "slack")
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != "slack" {
		t.Errorf("ID = %q, want slack", ch.ID)
	}
	if ch.Capabilities == nil {
		t.Fatal("Capabilities is nil")
	}
	if !ch.Capabilities.Groups {
		t.Error("Groups = false, want true")
	}
}

func TestChannelsService_Get_EmptyResponse(t *testing.T) {
	svc := newChannelSvc(&mockChannelClient{
		getChannel: func(_ context.Context, _ *connect.Request[outboxv1.GetChannelRequest]) (*connect.Response[outboxv1.GetChannelResponse], error) {
			return connect.NewResponse(&outboxv1.GetChannelResponse{}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "whatsapp"); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestChannelsService_List_PropagatesOptions(t *testing.T) {
	var got *outboxv1.ListChannelsRequest
	svc := newChannelSvc(&mockChannelClient{
		listChannels: func(_ context.Context, req *connect.Request[outboxv1.ListChannelsRequest]) (*connect.Response[outboxv1.ListChannelsResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.ListChannelsResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), &ListChannelsOptions{
		PageSize:  20,
		PageToken: "cursor",
	}); err != nil {
		t.Fatal(err)
	}
	if got.PageSize != 20 {
		t.Errorf("PageSize = %d, want 20", got.PageSize)
	}
	if got.PageToken != "cursor" {
		t.Errorf("PageToken = %q, want cursor", got.PageToken)
	}
}

func TestChannelsService_List_NilOptions(t *testing.T) {
	svc := newChannelSvc(&mockChannelClient{
		listChannels: func(_ context.Context, _ *connect.Request[outboxv1.ListChannelsRequest]) (*connect.Response[outboxv1.ListChannelsResponse], error) {
			return connect.NewResponse(&outboxv1.ListChannelsResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestChannelsService_List_ResultStructure(t *testing.T) {
	svc := newChannelSvc(&mockChannelClient{
		listChannels: func(_ context.Context, _ *connect.Request[outboxv1.ListChannelsRequest]) (*connect.Response[outboxv1.ListChannelsResponse], error) {
			return connect.NewResponse(&outboxv1.ListChannelsResponse{
				Channels: []*outboxv1.Channel{
					{Name: "channels/whatsapp"},
					{Name: "channels/slack"},
				},
				NextPageToken: "next-page",
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
	if res.Items[0].ID != "whatsapp" {
		t.Errorf("Items[0].ID = %q, want whatsapp", res.Items[0].ID)
	}
	if res.Items[1].ID != "slack" {
		t.Errorf("Items[1].ID = %q, want slack", res.Items[1].ID)
	}
	if res.NextPageToken != "next-page" {
		t.Errorf("NextPageToken = %q, want next-page", res.NextPageToken)
	}
	if res.TotalSize != 2 {
		t.Errorf("TotalSize = %d, want 2", res.TotalSize)
	}
}

func TestChannelsService_List_MapsCapabilities(t *testing.T) {
	svc := newChannelSvc(&mockChannelClient{
		listChannels: func(_ context.Context, _ *connect.Request[outboxv1.ListChannelsRequest]) (*connect.Response[outboxv1.ListChannelsResponse], error) {
			return connect.NewResponse(&outboxv1.ListChannelsResponse{
				Channels: []*outboxv1.Channel{
					{
						Name: "channels/whatsapp",
						Capabilities: &outboxv1.Channel_Capabilities{
							Groups:           true,
							ReadReceipts:     true,
							TypingIndicators: false,
						},
					},
				},
			}), nil
		},
	})
	res, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(res.Items))
	}
	caps := res.Items[0].Capabilities
	if caps == nil {
		t.Fatal("Capabilities is nil")
	}
	if !caps.Groups {
		t.Error("Groups = false, want true")
	}
	if !caps.ReadReceipts {
		t.Error("ReadReceipts = false, want true")
	}
	if caps.TypingIndicators {
		t.Error("TypingIndicators = true, want false")
	}
}
