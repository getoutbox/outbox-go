package outbox

import (
	"context"

	"connectrpc.com/connect"
	outboxv1 "github.com/getoutbox/outbox-go/outboxv1"
	outboxv1connect "github.com/getoutbox/outbox-go/outboxv1/outboxv1connect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// AccountsService provides operations on Accounts.
type AccountsService struct {
	client outboxv1connect.AccountServiceClient
}

// CreateAccountInput holds parameters for creating an Account.
type CreateAccountInput struct {
	ExternalID string
	ContactID  string
	Metadata   map[string]string
	RequestID  string
}

// UpdateAccountInput holds parameters for updating an Account.
// nil fields are not sent to the server — pass a non-nil map (including empty)
// to update metadata; nil means don't update.
type UpdateAccountInput struct {
	ID       string
	Metadata map[string]string // nil = don't update; empty map = clear all metadata
}

// ClaimAccountInput holds parameters for claiming an auto-created Account.
type ClaimAccountInput struct {
	ID        string
	ContactID string
	RequestID string
}

// ListAccountsOptions configures a List request.
type ListAccountsOptions struct {
	PageSize  int32
	PageToken string
	Filter    string
	OrderBy   string
}

// ListAccountsResult is the paginated result of a List call.
type ListAccountsResult struct {
	Items         []Account
	NextPageToken string
	TotalSize     int64
}

// Create creates a new Account.
func (s *AccountsService) Create(ctx context.Context, input CreateAccountInput) (*Account, error) {
	res, err := s.client.CreateAccount(ctx, connect.NewRequest(&outboxv1.CreateAccountRequest{
		Account: &outboxv1.Account{
			ExternalId: input.ExternalID,
			ContactId:  input.ContactID,
			Metadata:   input.Metadata,
		},
		RequestId: input.RequestID,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Account == nil {
		return nil, errEmpty("CreateAccount")
	}
	a := mapAccount(res.Msg.Account)
	return &a, nil
}

// Get retrieves an Account by its ID.
func (s *AccountsService) Get(ctx context.Context, id string) (*Account, error) {
	res, err := s.client.GetAccount(ctx, connect.NewRequest(&outboxv1.GetAccountRequest{
		Name: "accounts/" + id,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Account == nil {
		return nil, errEmpty("GetAccount")
	}
	a := mapAccount(res.Msg.Account)
	return &a, nil
}

// List returns a paginated list of Accounts.
func (s *AccountsService) List(ctx context.Context, opts *ListAccountsOptions) (*ListAccountsResult, error) {
	r := &outboxv1.ListAccountsRequest{}
	if opts != nil {
		r.PageSize = opts.PageSize
		r.PageToken = opts.PageToken
		r.Filter = opts.Filter
		r.OrderBy = opts.OrderBy
	}
	res, err := s.client.ListAccounts(ctx, connect.NewRequest(r))
	if err != nil {
		return nil, err
	}
	items := make([]Account, len(res.Msg.Accounts))
	for i, a := range res.Msg.Accounts {
		items[i] = mapAccount(a)
	}
	return &ListAccountsResult{
		Items:         items,
		NextPageToken: res.Msg.NextPageToken,
		TotalSize:     int64(res.Msg.TotalSize),
	}, nil
}

// Update updates an Account. Only non-nil fields are sent to the server via field mask.
func (s *AccountsService) Update(ctx context.Context, input UpdateAccountInput) (*Account, error) {
	acc := &outboxv1.Account{Name: "accounts/" + input.ID}
	var paths []string
	if input.Metadata != nil {
		acc.Metadata = input.Metadata
		paths = append(paths, "metadata")
	}
	var mask *fieldmaskpb.FieldMask
	if len(paths) > 0 {
		mask = &fieldmaskpb.FieldMask{Paths: paths}
	}
	res, err := s.client.UpdateAccount(ctx, connect.NewRequest(&outboxv1.UpdateAccountRequest{
		Account:    acc,
		UpdateMask: mask,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Account == nil {
		return nil, errEmpty("UpdateAccount")
	}
	a := mapAccount(res.Msg.Account)
	return &a, nil
}

// Delete deletes an Account by its ID.
func (s *AccountsService) Delete(ctx context.Context, id string) error {
	_, err := s.client.DeleteAccount(ctx, connect.NewRequest(&outboxv1.DeleteAccountRequest{
		Name: "accounts/" + id,
	}))
	return err
}

// Resolve looks up an Account by its external ID.
func (s *AccountsService) Resolve(ctx context.Context, externalID string) (*Account, error) {
	res, err := s.client.ResolveAccount(ctx, connect.NewRequest(&outboxv1.ResolveAccountRequest{
		ExternalId: externalID,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Account == nil {
		return nil, errEmpty("ResolveAccount")
	}
	a := mapAccount(res.Msg.Account)
	return &a, nil
}

// Claim claims an auto-created Account, associating it with a contact.
func (s *AccountsService) Claim(ctx context.Context, input ClaimAccountInput) (*Account, error) {
	res, err := s.client.ClaimAccount(ctx, connect.NewRequest(&outboxv1.ClaimAccountRequest{
		Name:      "accounts/" + input.ID,
		ContactId: input.ContactID,
		RequestId: input.RequestID,
	}))
	if err != nil {
		return nil, err
	}
	if res.Msg.Account == nil {
		return nil, errEmpty("ClaimAccount")
	}
	a := mapAccount(res.Msg.Account)
	return &a, nil
}
