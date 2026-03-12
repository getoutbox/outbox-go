package outbox

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"
)

func newAccountSvc(m *mockAccountClient) *AccountsService {
	return &AccountsService{client: m}
}

func okAccount(id string) *outboxv1.Account {
	return &outboxv1.Account{Name: "accounts/" + id}
}

func TestAccountsService_Create_RequestID(t *testing.T) {
	var got *outboxv1.CreateAccountRequest
	svc := newAccountSvc(&mockAccountClient{
		createAccount: func(_ context.Context, req *connect.Request[outboxv1.CreateAccountRequest]) (*connect.Response[outboxv1.CreateAccountResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.CreateAccountResponse{Account: okAccount("a1")}), nil
		},
	})
	if _, err := svc.Create(context.Background(), CreateAccountInput{RequestID: "idem-1"}); err != nil {
		t.Fatal(err)
	}
	if got.RequestId != "idem-1" {
		t.Errorf("RequestId = %q, want idem-1", got.RequestId)
	}
}

func TestAccountsService_Create_Fields(t *testing.T) {
	var got *outboxv1.CreateAccountRequest
	svc := newAccountSvc(&mockAccountClient{
		createAccount: func(_ context.Context, req *connect.Request[outboxv1.CreateAccountRequest]) (*connect.Response[outboxv1.CreateAccountResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.CreateAccountResponse{Account: okAccount("a1")}), nil
		},
	})
	if _, err := svc.Create(context.Background(), CreateAccountInput{
		ExternalID: "ext-1",
		ContactID:  "contact-1",
		Metadata:   map[string]string{"k": "v"},
	}); err != nil {
		t.Fatal(err)
	}
	if got.Account.ExternalId != "ext-1" {
		t.Errorf("ExternalId = %q, want ext-1", got.Account.ExternalId)
	}
	if got.Account.ContactId != "contact-1" {
		t.Errorf("ContactId = %q, want contact-1", got.Account.ContactId)
	}
	if got.Account.Metadata["k"] != "v" {
		t.Errorf("Metadata[k] = %q, want v", got.Account.Metadata["k"])
	}
}

func TestAccountsService_Create_EmptyResponse(t *testing.T) {
	svc := newAccountSvc(&mockAccountClient{
		createAccount: func(_ context.Context, _ *connect.Request[outboxv1.CreateAccountRequest]) (*connect.Response[outboxv1.CreateAccountResponse], error) {
			return connect.NewResponse(&outboxv1.CreateAccountResponse{}), nil
		},
	})
	if _, err := svc.Create(context.Background(), CreateAccountInput{}); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestAccountsService_List_PropagatesOptions(t *testing.T) {
	var got *outboxv1.ListAccountsRequest
	svc := newAccountSvc(&mockAccountClient{
		listAccounts: func(_ context.Context, req *connect.Request[outboxv1.ListAccountsRequest]) (*connect.Response[outboxv1.ListAccountsResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.ListAccountsResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), &ListAccountsOptions{
		PageSize:  10,
		PageToken: "tok",
		Filter:    "external_id=='x'",
		OrderBy:   "create_time desc",
	}); err != nil {
		t.Fatal(err)
	}
	if got.PageSize != 10 {
		t.Errorf("PageSize = %d, want 10", got.PageSize)
	}
	if got.PageToken != "tok" {
		t.Errorf("PageToken = %q, want tok", got.PageToken)
	}
	if got.Filter != "external_id=='x'" {
		t.Errorf("Filter = %q, want external_id=='x'", got.Filter)
	}
	if got.OrderBy != "create_time desc" {
		t.Errorf("OrderBy = %q, want create_time desc", got.OrderBy)
	}
}

func TestAccountsService_List_NilOptions(t *testing.T) {
	svc := newAccountSvc(&mockAccountClient{
		listAccounts: func(_ context.Context, _ *connect.Request[outboxv1.ListAccountsRequest]) (*connect.Response[outboxv1.ListAccountsResponse], error) {
			return connect.NewResponse(&outboxv1.ListAccountsResponse{}), nil
		},
	})
	if _, err := svc.List(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestAccountsService_List_ResultStructure(t *testing.T) {
	svc := newAccountSvc(&mockAccountClient{
		listAccounts: func(_ context.Context, _ *connect.Request[outboxv1.ListAccountsRequest]) (*connect.Response[outboxv1.ListAccountsResponse], error) {
			return connect.NewResponse(&outboxv1.ListAccountsResponse{
				Accounts:      []*outboxv1.Account{okAccount("a1"), okAccount("a2")},
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
	if res.Items[0].ID != "a1" {
		t.Errorf("Items[0].ID = %q, want a1", res.Items[0].ID)
	}
	if res.Items[1].ID != "a2" {
		t.Errorf("Items[1].ID = %q, want a2", res.Items[1].ID)
	}
	if res.NextPageToken != "next" {
		t.Errorf("NextPageToken = %q, want next", res.NextPageToken)
	}
	if res.TotalSize != 2 {
		t.Errorf("TotalSize = %d, want 2", res.TotalSize)
	}
}

func TestAccountsService_Get_EmptyResponse(t *testing.T) {
	svc := newAccountSvc(&mockAccountClient{
		getAccount: func(_ context.Context, _ *connect.Request[outboxv1.GetAccountRequest]) (*connect.Response[outboxv1.GetAccountResponse], error) {
			return connect.NewResponse(&outboxv1.GetAccountResponse{}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "a1"); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestAccountsService_Get_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newAccountSvc(&mockAccountClient{
		getAccount: func(_ context.Context, req *connect.Request[outboxv1.GetAccountRequest]) (*connect.Response[outboxv1.GetAccountResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.GetAccountResponse{Account: okAccount("a1")}), nil
		},
	})
	if _, err := svc.Get(context.Background(), "a1"); err != nil {
		t.Fatal(err)
	}
	if gotName != "accounts/a1" {
		t.Errorf("Name = %q, want accounts/a1", gotName)
	}
}

func TestAccountsService_Update_MetadataNil_NilMask(t *testing.T) {
	var got *outboxv1.UpdateAccountRequest
	svc := newAccountSvc(&mockAccountClient{
		updateAccount: func(_ context.Context, req *connect.Request[outboxv1.UpdateAccountRequest]) (*connect.Response[outboxv1.UpdateAccountResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateAccountResponse{Account: okAccount("a1")}), nil
		},
	})
	// nil Metadata → nothing to update → no mask
	if _, err := svc.Update(context.Background(), UpdateAccountInput{ID: "a1"}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask != nil {
		t.Errorf("expected nil update mask, got %v", got.UpdateMask.Paths)
	}
}

func TestAccountsService_Update_MetadataNonNil_HasMask(t *testing.T) {
	var got *outboxv1.UpdateAccountRequest
	svc := newAccountSvc(&mockAccountClient{
		updateAccount: func(_ context.Context, req *connect.Request[outboxv1.UpdateAccountRequest]) (*connect.Response[outboxv1.UpdateAccountResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateAccountResponse{Account: okAccount("a1")}), nil
		},
	})
	if _, err := svc.Update(context.Background(), UpdateAccountInput{
		ID:       "a1",
		Metadata: map[string]string{"key": "val"},
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if len(got.UpdateMask.Paths) != 1 || got.UpdateMask.Paths[0] != "metadata" {
		t.Errorf("paths = %v, want [metadata]", got.UpdateMask.Paths)
	}
	if got.Account.Metadata["key"] != "val" {
		t.Errorf("Metadata[key] = %q, want val", got.Account.Metadata["key"])
	}
}

func TestAccountsService_Update_MetadataEmpty_ClearsMetadata(t *testing.T) {
	var got *outboxv1.UpdateAccountRequest
	svc := newAccountSvc(&mockAccountClient{
		updateAccount: func(_ context.Context, req *connect.Request[outboxv1.UpdateAccountRequest]) (*connect.Response[outboxv1.UpdateAccountResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.UpdateAccountResponse{Account: okAccount("a1")}), nil
		},
	})
	// Empty (non-nil) map → update metadata to empty (clear all)
	if _, err := svc.Update(context.Background(), UpdateAccountInput{
		ID:       "a1",
		Metadata: map[string]string{},
	}); err != nil {
		t.Fatal(err)
	}
	if got.UpdateMask == nil {
		t.Fatal("expected update mask, got nil")
	}
	if len(got.UpdateMask.Paths) != 1 || got.UpdateMask.Paths[0] != "metadata" {
		t.Errorf("paths = %v, want [metadata]", got.UpdateMask.Paths)
	}
}

func TestAccountsService_Delete_BuildsResourceName(t *testing.T) {
	var gotName string
	svc := newAccountSvc(&mockAccountClient{
		deleteAccount: func(_ context.Context, req *connect.Request[outboxv1.DeleteAccountRequest]) (*connect.Response[outboxv1.DeleteAccountResponse], error) {
			gotName = req.Msg.Name
			return connect.NewResponse(&outboxv1.DeleteAccountResponse{}), nil
		},
	})
	if err := svc.Delete(context.Background(), "a1"); err != nil {
		t.Fatal(err)
	}
	if gotName != "accounts/a1" {
		t.Errorf("Name = %q, want accounts/a1", gotName)
	}
}

func TestAccountsService_Resolve_EmptyResponse(t *testing.T) {
	svc := newAccountSvc(&mockAccountClient{
		resolveAccount: func(_ context.Context, _ *connect.Request[outboxv1.ResolveAccountRequest]) (*connect.Response[outboxv1.ResolveAccountResponse], error) {
			return connect.NewResponse(&outboxv1.ResolveAccountResponse{}), nil
		},
	})
	if _, err := svc.Resolve(context.Background(), "ext-abc"); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestAccountsService_Resolve_PassesExternalID(t *testing.T) {
	var gotExtID string
	svc := newAccountSvc(&mockAccountClient{
		resolveAccount: func(_ context.Context, req *connect.Request[outboxv1.ResolveAccountRequest]) (*connect.Response[outboxv1.ResolveAccountResponse], error) {
			gotExtID = req.Msg.ExternalId
			return connect.NewResponse(&outboxv1.ResolveAccountResponse{Account: okAccount("a1")}), nil
		},
	})
	if _, err := svc.Resolve(context.Background(), "ext-abc"); err != nil {
		t.Fatal(err)
	}
	if gotExtID != "ext-abc" {
		t.Errorf("ExternalId = %q, want ext-abc", gotExtID)
	}
}

func TestAccountsService_Claim_EmptyResponse(t *testing.T) {
	svc := newAccountSvc(&mockAccountClient{
		claimAccount: func(_ context.Context, _ *connect.Request[outboxv1.ClaimAccountRequest]) (*connect.Response[outboxv1.ClaimAccountResponse], error) {
			return connect.NewResponse(&outboxv1.ClaimAccountResponse{}), nil
		},
	})
	if _, err := svc.Claim(context.Background(), ClaimAccountInput{ID: "a1"}); err == nil {
		t.Error("expected error for empty response, got nil")
	}
}

func TestAccountsService_Claim_BuildsResourceName(t *testing.T) {
	var got *outboxv1.ClaimAccountRequest
	svc := newAccountSvc(&mockAccountClient{
		claimAccount: func(_ context.Context, req *connect.Request[outboxv1.ClaimAccountRequest]) (*connect.Response[outboxv1.ClaimAccountResponse], error) {
			got = req.Msg
			return connect.NewResponse(&outboxv1.ClaimAccountResponse{Account: okAccount("a1")}), nil
		},
	})
	if _, err := svc.Claim(context.Background(), ClaimAccountInput{
		ID:        "a1",
		ContactID: "contact-xyz",
		RequestID: "req-claim",
	}); err != nil {
		t.Fatal(err)
	}
	if got.Name != "accounts/a1" {
		t.Errorf("Name = %q, want accounts/a1", got.Name)
	}
	if got.ContactId != "contact-xyz" {
		t.Errorf("ContactId = %q, want contact-xyz", got.ContactId)
	}
	if got.RequestId != "req-claim" {
		t.Errorf("RequestId = %q, want req-claim", got.RequestId)
	}
}
