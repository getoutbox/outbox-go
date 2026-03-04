package outbox

import (
	"context"
	"slices"
	"strings"
	"testing"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
)

func newMessageSvc(m *mockMessageClient) *MessagesService {
	return &MessagesService{client: m}
}

func okMessage(id string) *outboxv1.Message {
	return &outboxv1.Message{Name: "messages/" + id}
}

// ---- Send ----

func TestMessagesService_Send_BasicFields(t *testing.T) {
	var got *outboxv1.CreateMessageRequest
	svc := newMessageSvc(&mockMessageClient{
		createMessage: func(_ context.Context, req *connect.Request[outboxv1.CreateMessageRequest]) (*connect.Response[outboxv1.CreateMessageResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.CreateMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	if _, err := svc.Send(context.Background(), SendMessageInput{
		ConnectorID: "conn1",
		RecipientID: "acc1",
		RequestID:   "req-1",
	}); err != nil {
		t.Fatal(err)
	}
	if got.Connector != "connectors/conn1" {
		t.Errorf("Connector = %q, want connectors/conn1", got.Connector)
	}
	if got.Message.Recipient != "accounts/acc1" {
		t.Errorf("Recipient = %q, want accounts/acc1", got.Message.Recipient)
	}
	if got.RequestId != "req-1" {
		t.Errorf("RequestId = %q, want req-1", got.RequestId)
	}
}

func TestMessagesService_Send_ReplyTo(t *testing.T) {
	var got *outboxv1.CreateMessageRequest
	svc := newMessageSvc(&mockMessageClient{
		createMessage: func(_ context.Context, req *connect.Request[outboxv1.CreateMessageRequest]) (*connect.Response[outboxv1.CreateMessageResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.CreateMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	if _, err := svc.Send(context.Background(), SendMessageInput{
		ConnectorID:      "c1",
		RecipientID:      "a1",
		ReplyToMessageID: "orig",
	}); err != nil {
		t.Fatal(err)
	}
	if got.Message.ReplyTo != "messages/orig" {
		t.Errorf("ReplyTo = %q, want messages/orig", got.Message.ReplyTo)
	}
}

func TestMessagesService_Send_NoReplyTo_EmptyField(t *testing.T) {
	var got *outboxv1.CreateMessageRequest
	svc := newMessageSvc(&mockMessageClient{
		createMessage: func(_ context.Context, req *connect.Request[outboxv1.CreateMessageRequest]) (*connect.Response[outboxv1.CreateMessageResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.CreateMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	if _, err := svc.Send(context.Background(), SendMessageInput{ConnectorID: "c1", RecipientID: "a1"}); err != nil {
		t.Fatal(err)
	}
	if got.Message.ReplyTo != "" {
		t.Errorf("ReplyTo = %q, want empty", got.Message.ReplyTo)
	}
}

func TestMessagesService_Send_GroupID(t *testing.T) {
	var got *outboxv1.CreateMessageRequest
	svc := newMessageSvc(&mockMessageClient{
		createMessage: func(_ context.Context, req *connect.Request[outboxv1.CreateMessageRequest]) (*connect.Response[outboxv1.CreateMessageResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.CreateMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	if _, err := svc.Send(context.Background(), SendMessageInput{
		ConnectorID: "c1",
		RecipientID: "a1",
		GroupID:     "grp1",
	}); err != nil {
		t.Fatal(err)
	}
	if got.Message.GroupId != "grp1" {
		t.Errorf("GroupId = %q, want grp1", got.Message.GroupId)
	}
}

func TestMessagesService_Send_EmptyResponse(t *testing.T) {
	svc := newMessageSvc(&mockMessageClient{
		createMessage: func(_ context.Context, _ *connect.Request[outboxv1.CreateMessageRequest]) (*connect.Response[outboxv1.CreateMessageResponse], error) {
			return connect.NewResponse(&outboxv1.CreateMessageResponse{}), nil
		},
	})
	if _, err := svc.Send(context.Background(), SendMessageInput{}); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestMessagesService_Send_WithDelivery(t *testing.T) {
	svc := newMessageSvc(&mockMessageClient{
		createMessage: func(_ context.Context, _ *connect.Request[outboxv1.CreateMessageRequest]) (*connect.Response[outboxv1.CreateMessageResponse], error) {
			return connect.NewResponse(&outboxv1.CreateMessageResponse{
				Message: okMessage("m1"),
				Delivery: &outboxv1.MessageDelivery{
					Message: "messages/m1",
					Status:  outboxv1.MessageDelivery_STATUS_PENDING,
				},
			}), nil
		},
	})
	res, err := svc.Send(context.Background(), SendMessageInput{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Delivery == nil {
		t.Fatal("expected non-nil Delivery")
	}
	if res.Delivery.Status != MessageDeliveryStatusPending {
		t.Errorf("Delivery.Status = %v, want Pending", res.Delivery.Status)
	}
}

func TestMessagesService_Send_NilDelivery(t *testing.T) {
	svc := newMessageSvc(&mockMessageClient{
		createMessage: func(_ context.Context, _ *connect.Request[outboxv1.CreateMessageRequest]) (*connect.Response[outboxv1.CreateMessageResponse], error) {
			return connect.NewResponse(&outboxv1.CreateMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	res, err := svc.Send(context.Background(), SendMessageInput{})
	if err != nil {
		t.Fatal(err)
	}
	if res.Delivery != nil {
		t.Error("expected nil Delivery when not returned by server")
	}
}

// ---- Get ----

func TestMessagesService_Get_BuildsFilter(t *testing.T) {
	var got *outboxv1.ListMessagesRequest
	svc := newMessageSvc(&mockMessageClient{
		listMessages: func(_ context.Context, req *connect.Request[outboxv1.ListMessagesRequest]) (*connect.Response[outboxv1.ListMessagesResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.ListMessagesResponse{
				Messages: []*outboxv1.Message{okMessage("m1")},
			}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "m1"); err != nil {
		t.Fatal(err)
	}
	wantFilter := `name == "messages/m1"`
	if got.Filter != wantFilter {
		t.Errorf("Filter = %q, want %q", got.Filter, wantFilter)
	}
	if got.PageSize != 1 {
		t.Errorf("PageSize = %d, want 1", got.PageSize)
	}
}

func TestMessagesService_Get_EmptyList_Error(t *testing.T) {
	svc := newMessageSvc(&mockMessageClient{
		listMessages: func(_ context.Context, _ *connect.Request[outboxv1.ListMessagesRequest]) (*connect.Response[outboxv1.ListMessagesResponse], error) {
			return connect.NewResponse(&outboxv1.ListMessagesResponse{}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "m1"); err == nil {
		t.Error("expected error for empty list, got nil")
	}
}

func TestMessagesGet_InvalidIDWithQuote(t *testing.T) {
	s := &MessagesService{client: nil}
	_, err := s.Get(context.Background(), `msg"id`)
	if err == nil {
		t.Fatal("expected error for ID containing quote, got nil")
	}
	if !strings.Contains(err.Error(), "invalid message ID") {
		t.Errorf("error = %q, want to contain 'invalid message ID'", err.Error())
	}
}

// ---- List ----

func TestMessagesService_List_PropagatesOptions(t *testing.T) {
	var got *outboxv1.ListMessagesRequest
	svc := newMessageSvc(&mockMessageClient{
		listMessages: func(_ context.Context, req *connect.Request[outboxv1.ListMessagesRequest]) (*connect.Response[outboxv1.ListMessagesResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.ListMessagesResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), &ListMessagesOptions{
		ConnectorID: "c1",
		PageSize:    5,
		PageToken:   "tok",
		Filter:      "direction==INBOUND",
		OrderBy:     "create_time desc",
	}); err != nil {
		t.Fatal(err)
	}
	if got.Parent != "connectors/c1" {
		t.Errorf("Parent = %q, want connectors/c1", got.Parent)
	}
	if got.PageSize != 5 {
		t.Errorf("PageSize = %d, want 5", got.PageSize)
	}
	if got.Filter != "direction==INBOUND" {
		t.Errorf("Filter = %q, want direction==INBOUND", got.Filter)
	}
	if got.OrderBy != "create_time desc" {
		t.Errorf("OrderBy = %q, want create_time desc", got.OrderBy)
	}
}

func TestMessagesService_List_NoConnectorID_NoParent(t *testing.T) {
	var got *outboxv1.ListMessagesRequest
	svc := newMessageSvc(&mockMessageClient{
		listMessages: func(_ context.Context, req *connect.Request[outboxv1.ListMessagesRequest]) (*connect.Response[outboxv1.ListMessagesResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.ListMessagesResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), &ListMessagesOptions{}); err != nil {
		t.Fatal(err)
	}
	if got.Parent != "" {
		t.Errorf("Parent = %q, want empty when ConnectorID not set", got.Parent)
	}
}

func TestMessagesService_List_NilOptions(t *testing.T) {
	svc := newMessageSvc(&mockMessageClient{
		listMessages: func(_ context.Context, _ *connect.Request[outboxv1.ListMessagesRequest]) (*connect.Response[outboxv1.ListMessagesResponse], error) {
			return connect.NewResponse(&outboxv1.ListMessagesResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestMessagesService_List_ResultStructure(t *testing.T) {
	svc := newMessageSvc(&mockMessageClient{
		listMessages: func(_ context.Context, _ *connect.Request[outboxv1.ListMessagesRequest]) (*connect.Response[outboxv1.ListMessagesResponse], error) {
			return connect.NewResponse(&outboxv1.ListMessagesResponse{
				Messages:      []*outboxv1.Message{okMessage("m1"), okMessage("m2")},
				NextPageToken: "next",
				TotalSize:     42,
			}), nil
		},
	})
	res, err := svc.List(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Items) != 2 {
		t.Errorf("len(Items) = %d, want 2", len(res.Items))
	}
	if res.NextPageToken != "next" {
		t.Errorf("NextPageToken = %q, want next", res.NextPageToken)
	}
	if res.TotalSize != 42 {
		t.Errorf("TotalSize = %d, want 42", res.TotalSize)
	}
}

// ---- Update ----

func TestMessagesService_Update_PartsOnly(t *testing.T) {
	var got *outboxv1.UpdateMessageRequest
	svc := newMessageSvc(&mockMessageClient{
		updateMessage: func(_ context.Context, req *connect.Request[outboxv1.UpdateMessageRequest]) (*connect.Response[outboxv1.UpdateMessageResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateMessageInput{
		ID:    "m1",
		Parts: []MessagePart{TextPart("hello")},
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if !slices.Equal(got.UpdateMask.Paths, []string{"parts"}) {
		t.Errorf("paths = %v, want [parts]", got.UpdateMask.Paths)
	}
}

func TestMessagesService_Update_MetadataOnly(t *testing.T) {
	var got *outboxv1.UpdateMessageRequest
	svc := newMessageSvc(&mockMessageClient{
		updateMessage: func(_ context.Context, req *connect.Request[outboxv1.UpdateMessageRequest]) (*connect.Response[outboxv1.UpdateMessageResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateMessageInput{
		ID:       "m1",
		Metadata: map[string]string{"k": "v"},
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if !slices.Equal(got.UpdateMask.Paths, []string{"metadata"}) {
		t.Errorf("paths = %v, want [metadata]", got.UpdateMask.Paths)
	}
}

func TestMessagesService_Update_BothPartsAndMetadata(t *testing.T) {
	var got *outboxv1.UpdateMessageRequest
	svc := newMessageSvc(&mockMessageClient{
		updateMessage: func(_ context.Context, req *connect.Request[outboxv1.UpdateMessageRequest]) (*connect.Response[outboxv1.UpdateMessageResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateMessageInput{
		ID:       "m1",
		Parts:    []MessagePart{TextPart("hi")},
		Metadata: map[string]string{"x": "y"},
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if !slices.Contains(got.UpdateMask.Paths, "parts") {
		t.Errorf("paths %v missing parts", got.UpdateMask.Paths)
	}
	if !slices.Contains(got.UpdateMask.Paths, "metadata") {
		t.Errorf("paths %v missing metadata", got.UpdateMask.Paths)
	}
}

func TestMessagesService_Update_NoFields_NilMask(t *testing.T) {
	var got *outboxv1.UpdateMessageRequest
	svc := newMessageSvc(&mockMessageClient{
		updateMessage: func(_ context.Context, req *connect.Request[outboxv1.UpdateMessageRequest]) (*connect.Response[outboxv1.UpdateMessageResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateMessageInput{ID: "m1"}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask != nil {
		t.Errorf("expected nil update mask, got %v", got.UpdateMask.Paths)
	}
}

func TestMessagesService_Update_RequestID(t *testing.T) {
	var got *outboxv1.UpdateMessageRequest
	svc := newMessageSvc(&mockMessageClient{
		updateMessage: func(_ context.Context, req *connect.Request[outboxv1.UpdateMessageRequest]) (*connect.Response[outboxv1.UpdateMessageResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateMessageInput{ID: "m1", RequestID: "idem-upd"}); err != nil {
		t.Fatal(err)
	}
	if got.RequestId != "idem-upd" {
		t.Errorf("RequestId = %q, want idem-upd", got.RequestId)
	}
}

func TestMessagesService_Update_EmptyResponse(t *testing.T) {
	svc := newMessageSvc(&mockMessageClient{
		updateMessage: func(_ context.Context, _ *connect.Request[outboxv1.UpdateMessageRequest]) (*connect.Response[outboxv1.UpdateMessageResponse], error) {
			return connect.NewResponse(&outboxv1.UpdateMessageResponse{}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateMessageInput{ID: "m1", Parts: []MessagePart{TextPart("hi")}}); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

// ---- Delete ----

func TestMessagesService_Delete_BuildsResourceName(t *testing.T) {
	var got *outboxv1.DeleteMessageRequest
	svc := newMessageSvc(&mockMessageClient{
		deleteMessage: func(_ context.Context, req *connect.Request[outboxv1.DeleteMessageRequest]) (*connect.Response[outboxv1.DeleteMessageResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.DeleteMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	if _, err := svc.Delete(context.Background(), DeleteMessageInput{ID: "m1"}); err != nil {
		t.Fatal(err)
	}
	if got.Name != "messages/m1" {
		t.Errorf("Name = %q, want messages/m1", got.Name)
	}
}

func TestMessagesService_Delete_PropagatesScopeAndRequestID(t *testing.T) {
	var got *outboxv1.DeleteMessageRequest
	svc := newMessageSvc(&mockMessageClient{
		deleteMessage: func(_ context.Context, req *connect.Request[outboxv1.DeleteMessageRequest]) (*connect.Response[outboxv1.DeleteMessageResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.DeleteMessageResponse{Message: okMessage("m1")}), nil
		},
	})
	if _, err := svc.Delete(context.Background(), DeleteMessageInput{
		ID:            "m1",
		DeletionScope: MessageDeletionScopeForEveryone,
		RequestID:     "del-1",
	}); err != nil {
		t.Fatal(err)
	}
	if got.Scope != MessageDeletionScopeForEveryone {
		t.Errorf("Scope = %v, want ForEveryone", got.Scope)
	}
	if got.RequestId != "del-1" {
		t.Errorf("RequestId = %q, want del-1", got.RequestId)
	}
}

func TestMessagesService_Delete_EmptyResponse(t *testing.T) {
	svc := newMessageSvc(&mockMessageClient{
		deleteMessage: func(_ context.Context, _ *connect.Request[outboxv1.DeleteMessageRequest]) (*connect.Response[outboxv1.DeleteMessageResponse], error) {
			return connect.NewResponse(&outboxv1.DeleteMessageResponse{}), nil
		},
	})
	if _, err := svc.Delete(context.Background(), DeleteMessageInput{ID: "m1"}); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

// ---- MarkRead ----

func TestMessagesService_MarkRead_BuildsResourceNames(t *testing.T) {
	var got *outboxv1.SendReadReceiptRequest
	svc := newMessageSvc(&mockMessageClient{
		sendReadReceipt: func(_ context.Context, req *connect.Request[outboxv1.SendReadReceiptRequest]) (*connect.Response[outboxv1.SendReadReceiptResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.SendReadReceiptResponse{}), nil
		},
	})
	if err := svc.MarkRead(context.Background(), MarkReadInput{
		ConnectorID: "c1",
		AccountID:   "a1",
		MessageIDs:  []string{"m1", "m2"},
	}); err != nil {
		t.Fatal(err)
	}
	if got.Connector != "connectors/c1" {
		t.Errorf("Connector = %q, want connectors/c1", got.Connector)
	}
	if got.Account != "accounts/a1" {
		t.Errorf("Account = %q, want accounts/a1", got.Account)
	}
	if !slices.Equal(got.Messages, []string{"messages/m1", "messages/m2"}) {
		t.Errorf("Messages = %v, want [messages/m1 messages/m2]", got.Messages)
	}
}

// ---- SendTypingIndicator ----

func TestMessagesService_SendTypingIndicator_True(t *testing.T) {
	var got *outboxv1.SendTypingIndicatorRequest
	svc := newMessageSvc(&mockMessageClient{
		sendTypingIndicator: func(_ context.Context, req *connect.Request[outboxv1.SendTypingIndicatorRequest]) (*connect.Response[outboxv1.SendTypingIndicatorResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.SendTypingIndicatorResponse{}), nil
		},
	})
	if err := svc.SendTypingIndicator(context.Background(), TypingInput{
		ConnectorID: "c1",
		AccountID:   "a1",
		Typing:      true,
	}); err != nil {
		t.Fatal(err)
	}
	if got.Connector != "connectors/c1" {
		t.Errorf("Connector = %q, want connectors/c1", got.Connector)
	}
	if got.Account != "accounts/a1" {
		t.Errorf("Account = %q, want accounts/a1", got.Account)
	}
	if got.Typing == nil || !*got.Typing {
		t.Errorf("Typing = %v, want true", got.Typing)
	}
}

func TestMessagesService_SendTypingIndicator_False(t *testing.T) {
	var got *outboxv1.SendTypingIndicatorRequest
	svc := newMessageSvc(&mockMessageClient{
		sendTypingIndicator: func(_ context.Context, req *connect.Request[outboxv1.SendTypingIndicatorRequest]) (*connect.Response[outboxv1.SendTypingIndicatorResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.SendTypingIndicatorResponse{}), nil
		},
	})
	if err := svc.SendTypingIndicator(context.Background(), TypingInput{Typing: false}); err != nil {
		t.Fatal(err)
	}
	if got.Typing == nil || *got.Typing {
		t.Errorf("Typing = %v, want false", got.Typing)
	}
}

// ---- History ----

func TestMessagesService_History_BuildsFilter(t *testing.T) {
	var got *outboxv1.ListMessagesRequest
	svc := newMessageSvc(&mockMessageClient{
		listMessages: func(_ context.Context, req *connect.Request[outboxv1.ListMessagesRequest]) (*connect.Response[outboxv1.ListMessagesResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.ListMessagesResponse{}), nil
		},
	})
	if _, err := svc.History(context.Background(), HistoryInput{
		ConnectorID: "c1",
		AccountID:   "a1",
	}); err != nil {
		t.Fatal(err)
	}
	wantFilter := `account.name == "accounts/a1"`
	if got.Filter != wantFilter {
		t.Errorf("Filter = %q, want %q", got.Filter, wantFilter)
	}
	if got.OrderBy != "create_time asc" {
		t.Errorf("OrderBy = %q, want create_time asc", got.OrderBy)
	}
	if got.Parent != "connectors/c1" {
		t.Errorf("Parent = %q, want connectors/c1", got.Parent)
	}
}

func TestMessagesHistory_InvalidAccountIDWithQuote(t *testing.T) {
	s := &MessagesService{client: nil}
	_, err := s.History(context.Background(), HistoryInput{AccountID: `acc"id`})
	if err == nil {
		t.Fatal("expected error for AccountID containing quote, got nil")
	}
	if !strings.Contains(err.Error(), "invalid account ID") {
		t.Errorf("error = %q, want to contain 'invalid account ID'", err.Error())
	}
}

func TestMessagesService_History_PropagatesPagination(t *testing.T) {
	var got *outboxv1.ListMessagesRequest
	svc := newMessageSvc(&mockMessageClient{
		listMessages: func(_ context.Context, req *connect.Request[outboxv1.ListMessagesRequest]) (*connect.Response[outboxv1.ListMessagesResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.ListMessagesResponse{}), nil
		},
	})
	if _, err := svc.History(context.Background(), HistoryInput{
		AccountID: "a1",
		PageSize:  10,
		PageToken: "page2",
	}); err != nil {
		t.Fatal(err)
	}
	if got.PageSize != 10 {
		t.Errorf("PageSize = %d, want 10", got.PageSize)
	}
	if got.PageToken != "page2" {
		t.Errorf("PageToken = %q, want page2", got.PageToken)
	}
}
