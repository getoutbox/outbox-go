package outbox

import (
	"context"

	"connectrpc.com/connect"
	longrunningpb "cloud.google.com/go/longrunning/autogen/longrunningpb"
	outboxv1 "github.com/getoutbox/outbox-go/gen/outbox/v1"
)

// ---- mockConnectorClient ----

type mockConnectorClient struct {
	createConnector           func(context.Context, *connect.Request[outboxv1.CreateConnectorRequest]) (*connect.Response[outboxv1.CreateConnectorResponse], error)
	createManagedConnector    func(context.Context, *connect.Request[outboxv1.CreateManagedConnectorRequest]) (*connect.Response[longrunningpb.Operation], error)
	getConnector              func(context.Context, *connect.Request[outboxv1.GetConnectorRequest]) (*connect.Response[outboxv1.GetConnectorResponse], error)
	listConnectors            func(context.Context, *connect.Request[outboxv1.ListConnectorsRequest]) (*connect.Response[outboxv1.ListConnectorsResponse], error)
	updateConnector           func(context.Context, *connect.Request[outboxv1.UpdateConnectorRequest]) (*connect.Response[outboxv1.UpdateConnectorResponse], error)
	deleteConnector           func(context.Context, *connect.Request[outboxv1.DeleteConnectorRequest]) (*connect.Response[outboxv1.DeleteConnectorResponse], error)
	detachProvisionedResource func(context.Context, *connect.Request[outboxv1.DetachProvisionedResourceRequest]) (*connect.Response[outboxv1.DetachProvisionedResourceResponse], error)
	reauthorizeConnector      func(context.Context, *connect.Request[outboxv1.ReauthorizeConnectorRequest]) (*connect.Response[outboxv1.ReauthorizeConnectorResponse], error)
	activateConnector         func(context.Context, *connect.Request[outboxv1.ActivateConnectorRequest]) (*connect.Response[outboxv1.ActivateConnectorResponse], error)
	deactivateConnector       func(context.Context, *connect.Request[outboxv1.DeactivateConnectorRequest]) (*connect.Response[outboxv1.DeactivateConnectorResponse], error)
	verifyConnector           func(context.Context, *connect.Request[outboxv1.VerifyConnectorRequest]) (*connect.Response[outboxv1.VerifyConnectorResponse], error)
}

func (m *mockConnectorClient) CreateConnector(ctx context.Context, req *connect.Request[outboxv1.CreateConnectorRequest]) (*connect.Response[outboxv1.CreateConnectorResponse], error) {
	if m.createConnector != nil {
		return m.createConnector(ctx, req)
	}
	panic("mockConnectorClient.CreateConnector not set")
}

func (m *mockConnectorClient) CreateManagedConnector(ctx context.Context, req *connect.Request[outboxv1.CreateManagedConnectorRequest]) (*connect.Response[longrunningpb.Operation], error) {
	if m.createManagedConnector != nil {
		return m.createManagedConnector(ctx, req)
	}
	panic("mockConnectorClient.CreateManagedConnector not set")
}

func (m *mockConnectorClient) GetConnector(ctx context.Context, req *connect.Request[outboxv1.GetConnectorRequest]) (*connect.Response[outboxv1.GetConnectorResponse], error) {
	if m.getConnector != nil {
		return m.getConnector(ctx, req)
	}
	panic("mockConnectorClient.GetConnector not set")
}

func (m *mockConnectorClient) ListConnectors(ctx context.Context, req *connect.Request[outboxv1.ListConnectorsRequest]) (*connect.Response[outboxv1.ListConnectorsResponse], error) {
	if m.listConnectors != nil {
		return m.listConnectors(ctx, req)
	}
	panic("mockConnectorClient.ListConnectors not set")
}

func (m *mockConnectorClient) UpdateConnector(ctx context.Context, req *connect.Request[outboxv1.UpdateConnectorRequest]) (*connect.Response[outboxv1.UpdateConnectorResponse], error) {
	if m.updateConnector != nil {
		return m.updateConnector(ctx, req)
	}
	panic("mockConnectorClient.UpdateConnector not set")
}

func (m *mockConnectorClient) DeleteConnector(ctx context.Context, req *connect.Request[outboxv1.DeleteConnectorRequest]) (*connect.Response[outboxv1.DeleteConnectorResponse], error) {
	if m.deleteConnector != nil {
		return m.deleteConnector(ctx, req)
	}
	panic("mockConnectorClient.DeleteConnector not set")
}

func (m *mockConnectorClient) DetachProvisionedResource(ctx context.Context, req *connect.Request[outboxv1.DetachProvisionedResourceRequest]) (*connect.Response[outboxv1.DetachProvisionedResourceResponse], error) {
	if m.detachProvisionedResource != nil {
		return m.detachProvisionedResource(ctx, req)
	}
	panic("mockConnectorClient.DetachProvisionedResource not set")
}

func (m *mockConnectorClient) ReauthorizeConnector(ctx context.Context, req *connect.Request[outboxv1.ReauthorizeConnectorRequest]) (*connect.Response[outboxv1.ReauthorizeConnectorResponse], error) {
	if m.reauthorizeConnector != nil {
		return m.reauthorizeConnector(ctx, req)
	}
	panic("mockConnectorClient.ReauthorizeConnector not set")
}

func (m *mockConnectorClient) ActivateConnector(ctx context.Context, req *connect.Request[outboxv1.ActivateConnectorRequest]) (*connect.Response[outboxv1.ActivateConnectorResponse], error) {
	if m.activateConnector != nil {
		return m.activateConnector(ctx, req)
	}
	panic("mockConnectorClient.ActivateConnector not set")
}

func (m *mockConnectorClient) DeactivateConnector(ctx context.Context, req *connect.Request[outboxv1.DeactivateConnectorRequest]) (*connect.Response[outboxv1.DeactivateConnectorResponse], error) {
	if m.deactivateConnector != nil {
		return m.deactivateConnector(ctx, req)
	}
	panic("mockConnectorClient.DeactivateConnector not set")
}

func (m *mockConnectorClient) VerifyConnector(ctx context.Context, req *connect.Request[outboxv1.VerifyConnectorRequest]) (*connect.Response[outboxv1.VerifyConnectorResponse], error) {
	if m.verifyConnector != nil {
		return m.verifyConnector(ctx, req)
	}
	panic("mockConnectorClient.VerifyConnector not set")
}

// ---- mockTemplateClient ----

type mockTemplateClient struct {
	createTemplate func(context.Context, *connect.Request[outboxv1.CreateTemplateRequest]) (*connect.Response[outboxv1.CreateTemplateResponse], error)
	getTemplate    func(context.Context, *connect.Request[outboxv1.GetTemplateRequest]) (*connect.Response[outboxv1.GetTemplateResponse], error)
	listTemplates  func(context.Context, *connect.Request[outboxv1.ListTemplatesRequest]) (*connect.Response[outboxv1.ListTemplatesResponse], error)
	deleteTemplate func(context.Context, *connect.Request[outboxv1.DeleteTemplateRequest]) (*connect.Response[outboxv1.DeleteTemplateResponse], error)
}

func (m *mockTemplateClient) CreateTemplate(ctx context.Context, req *connect.Request[outboxv1.CreateTemplateRequest]) (*connect.Response[outboxv1.CreateTemplateResponse], error) {
	if m.createTemplate != nil {
		return m.createTemplate(ctx, req)
	}
	panic("mockTemplateClient.CreateTemplate not set")
}

func (m *mockTemplateClient) GetTemplate(ctx context.Context, req *connect.Request[outboxv1.GetTemplateRequest]) (*connect.Response[outboxv1.GetTemplateResponse], error) {
	if m.getTemplate != nil {
		return m.getTemplate(ctx, req)
	}
	panic("mockTemplateClient.GetTemplate not set")
}

func (m *mockTemplateClient) ListTemplates(ctx context.Context, req *connect.Request[outboxv1.ListTemplatesRequest]) (*connect.Response[outboxv1.ListTemplatesResponse], error) {
	if m.listTemplates != nil {
		return m.listTemplates(ctx, req)
	}
	panic("mockTemplateClient.ListTemplates not set")
}

func (m *mockTemplateClient) DeleteTemplate(ctx context.Context, req *connect.Request[outboxv1.DeleteTemplateRequest]) (*connect.Response[outboxv1.DeleteTemplateResponse], error) {
	if m.deleteTemplate != nil {
		return m.deleteTemplate(ctx, req)
	}
	panic("mockTemplateClient.DeleteTemplate not set")
}

// ---- mockAccountClient ----

type mockAccountClient struct {
	createAccount  func(context.Context, *connect.Request[outboxv1.CreateAccountRequest]) (*connect.Response[outboxv1.CreateAccountResponse], error)
	getAccount     func(context.Context, *connect.Request[outboxv1.GetAccountRequest]) (*connect.Response[outboxv1.GetAccountResponse], error)
	listAccounts   func(context.Context, *connect.Request[outboxv1.ListAccountsRequest]) (*connect.Response[outboxv1.ListAccountsResponse], error)
	updateAccount  func(context.Context, *connect.Request[outboxv1.UpdateAccountRequest]) (*connect.Response[outboxv1.UpdateAccountResponse], error)
	claimAccount   func(context.Context, *connect.Request[outboxv1.ClaimAccountRequest]) (*connect.Response[outboxv1.ClaimAccountResponse], error)
	deleteAccount  func(context.Context, *connect.Request[outboxv1.DeleteAccountRequest]) (*connect.Response[outboxv1.DeleteAccountResponse], error)
	resolveAccount func(context.Context, *connect.Request[outboxv1.ResolveAccountRequest]) (*connect.Response[outboxv1.ResolveAccountResponse], error)
}

func (m *mockAccountClient) CreateAccount(ctx context.Context, req *connect.Request[outboxv1.CreateAccountRequest]) (*connect.Response[outboxv1.CreateAccountResponse], error) {
	if m.createAccount != nil {
		return m.createAccount(ctx, req)
	}
	panic("mockAccountClient.CreateAccount not set")
}

func (m *mockAccountClient) GetAccount(ctx context.Context, req *connect.Request[outboxv1.GetAccountRequest]) (*connect.Response[outboxv1.GetAccountResponse], error) {
	if m.getAccount != nil {
		return m.getAccount(ctx, req)
	}
	panic("mockAccountClient.GetAccount not set")
}

func (m *mockAccountClient) ListAccounts(ctx context.Context, req *connect.Request[outboxv1.ListAccountsRequest]) (*connect.Response[outboxv1.ListAccountsResponse], error) {
	if m.listAccounts != nil {
		return m.listAccounts(ctx, req)
	}
	panic("mockAccountClient.ListAccounts not set")
}

func (m *mockAccountClient) UpdateAccount(ctx context.Context, req *connect.Request[outboxv1.UpdateAccountRequest]) (*connect.Response[outboxv1.UpdateAccountResponse], error) {
	if m.updateAccount != nil {
		return m.updateAccount(ctx, req)
	}
	panic("mockAccountClient.UpdateAccount not set")
}

func (m *mockAccountClient) ClaimAccount(ctx context.Context, req *connect.Request[outboxv1.ClaimAccountRequest]) (*connect.Response[outboxv1.ClaimAccountResponse], error) {
	if m.claimAccount != nil {
		return m.claimAccount(ctx, req)
	}
	panic("mockAccountClient.ClaimAccount not set")
}

func (m *mockAccountClient) DeleteAccount(ctx context.Context, req *connect.Request[outboxv1.DeleteAccountRequest]) (*connect.Response[outboxv1.DeleteAccountResponse], error) {
	if m.deleteAccount != nil {
		return m.deleteAccount(ctx, req)
	}
	panic("mockAccountClient.DeleteAccount not set")
}

func (m *mockAccountClient) ResolveAccount(ctx context.Context, req *connect.Request[outboxv1.ResolveAccountRequest]) (*connect.Response[outboxv1.ResolveAccountResponse], error) {
	if m.resolveAccount != nil {
		return m.resolveAccount(ctx, req)
	}
	panic("mockAccountClient.ResolveAccount not set")
}

// ---- mockMessageClient ----

type mockMessageClient struct {
	createMessage       func(context.Context, *connect.Request[outboxv1.CreateMessageRequest]) (*connect.Response[outboxv1.CreateMessageResponse], error)
	listMessages        func(context.Context, *connect.Request[outboxv1.ListMessagesRequest]) (*connect.Response[outboxv1.ListMessagesResponse], error)
	updateMessage       func(context.Context, *connect.Request[outboxv1.UpdateMessageRequest]) (*connect.Response[outboxv1.UpdateMessageResponse], error)
	deleteMessage       func(context.Context, *connect.Request[outboxv1.DeleteMessageRequest]) (*connect.Response[outboxv1.DeleteMessageResponse], error)
	sendReadReceipt     func(context.Context, *connect.Request[outboxv1.SendReadReceiptRequest]) (*connect.Response[outboxv1.SendReadReceiptResponse], error)
	sendTypingIndicator func(context.Context, *connect.Request[outboxv1.SendTypingIndicatorRequest]) (*connect.Response[outboxv1.SendTypingIndicatorResponse], error)
}

func (m *mockMessageClient) CreateMessage(ctx context.Context, req *connect.Request[outboxv1.CreateMessageRequest]) (*connect.Response[outboxv1.CreateMessageResponse], error) {
	if m.createMessage != nil {
		return m.createMessage(ctx, req)
	}
	panic("mockMessageClient.CreateMessage not set")
}

func (m *mockMessageClient) ListMessages(ctx context.Context, req *connect.Request[outboxv1.ListMessagesRequest]) (*connect.Response[outboxv1.ListMessagesResponse], error) {
	if m.listMessages != nil {
		return m.listMessages(ctx, req)
	}
	panic("mockMessageClient.ListMessages not set")
}

func (m *mockMessageClient) UpdateMessage(ctx context.Context, req *connect.Request[outboxv1.UpdateMessageRequest]) (*connect.Response[outboxv1.UpdateMessageResponse], error) {
	if m.updateMessage != nil {
		return m.updateMessage(ctx, req)
	}
	panic("mockMessageClient.UpdateMessage not set")
}

func (m *mockMessageClient) DeleteMessage(ctx context.Context, req *connect.Request[outboxv1.DeleteMessageRequest]) (*connect.Response[outboxv1.DeleteMessageResponse], error) {
	if m.deleteMessage != nil {
		return m.deleteMessage(ctx, req)
	}
	panic("mockMessageClient.DeleteMessage not set")
}

func (m *mockMessageClient) SendReadReceipt(ctx context.Context, req *connect.Request[outboxv1.SendReadReceiptRequest]) (*connect.Response[outboxv1.SendReadReceiptResponse], error) {
	if m.sendReadReceipt != nil {
		return m.sendReadReceipt(ctx, req)
	}
	panic("mockMessageClient.SendReadReceipt not set")
}

func (m *mockMessageClient) SendTypingIndicator(ctx context.Context, req *connect.Request[outboxv1.SendTypingIndicatorRequest]) (*connect.Response[outboxv1.SendTypingIndicatorResponse], error) {
	if m.sendTypingIndicator != nil {
		return m.sendTypingIndicator(ctx, req)
	}
	panic("mockMessageClient.SendTypingIndicator not set")
}

// ---- mockDestinationClient ----

type mockDestinationClient struct {
	createDestination          func(context.Context, *connect.Request[outboxv1.CreateDestinationRequest]) (*connect.Response[outboxv1.CreateDestinationResponse], error)
	getDestination             func(context.Context, *connect.Request[outboxv1.GetDestinationRequest]) (*connect.Response[outboxv1.GetDestinationResponse], error)
	listDestinations           func(context.Context, *connect.Request[outboxv1.ListDestinationsRequest]) (*connect.Response[outboxv1.ListDestinationsResponse], error)
	updateDestination          func(context.Context, *connect.Request[outboxv1.UpdateDestinationRequest]) (*connect.Response[outboxv1.UpdateDestinationResponse], error)
	deleteDestination          func(context.Context, *connect.Request[outboxv1.DeleteDestinationRequest]) (*connect.Response[outboxv1.DeleteDestinationResponse], error)
	testDestination            func(context.Context, *connect.Request[outboxv1.TestDestinationRequest]) (*connect.Response[outboxv1.TestDestinationResponse], error)
	listDestinationTestResults func(context.Context, *connect.Request[outboxv1.ListDestinationTestResultsRequest]) (*connect.Response[outboxv1.ListDestinationTestResultsResponse], error)
	validateDestinationFilter  func(context.Context, *connect.Request[outboxv1.ValidateDestinationFilterRequest]) (*connect.Response[outboxv1.ValidateDestinationFilterResponse], error)
	pollEvents                 func(context.Context, *connect.Request[outboxv1.PollEventsRequest]) (*connect.Response[outboxv1.PollEventsResponse], error)
}

func (m *mockDestinationClient) CreateDestination(ctx context.Context, req *connect.Request[outboxv1.CreateDestinationRequest]) (*connect.Response[outboxv1.CreateDestinationResponse], error) {
	if m.createDestination != nil {
		return m.createDestination(ctx, req)
	}
	panic("mockDestinationClient.CreateDestination not set")
}

func (m *mockDestinationClient) GetDestination(ctx context.Context, req *connect.Request[outboxv1.GetDestinationRequest]) (*connect.Response[outboxv1.GetDestinationResponse], error) {
	if m.getDestination != nil {
		return m.getDestination(ctx, req)
	}
	panic("mockDestinationClient.GetDestination not set")
}

func (m *mockDestinationClient) ListDestinations(ctx context.Context, req *connect.Request[outboxv1.ListDestinationsRequest]) (*connect.Response[outboxv1.ListDestinationsResponse], error) {
	if m.listDestinations != nil {
		return m.listDestinations(ctx, req)
	}
	panic("mockDestinationClient.ListDestinations not set")
}

func (m *mockDestinationClient) UpdateDestination(ctx context.Context, req *connect.Request[outboxv1.UpdateDestinationRequest]) (*connect.Response[outboxv1.UpdateDestinationResponse], error) {
	if m.updateDestination != nil {
		return m.updateDestination(ctx, req)
	}
	panic("mockDestinationClient.UpdateDestination not set")
}

func (m *mockDestinationClient) DeleteDestination(ctx context.Context, req *connect.Request[outboxv1.DeleteDestinationRequest]) (*connect.Response[outboxv1.DeleteDestinationResponse], error) {
	if m.deleteDestination != nil {
		return m.deleteDestination(ctx, req)
	}
	panic("mockDestinationClient.DeleteDestination not set")
}

func (m *mockDestinationClient) TestDestination(ctx context.Context, req *connect.Request[outboxv1.TestDestinationRequest]) (*connect.Response[outboxv1.TestDestinationResponse], error) {
	if m.testDestination != nil {
		return m.testDestination(ctx, req)
	}
	panic("mockDestinationClient.TestDestination not set")
}

func (m *mockDestinationClient) ListDestinationTestResults(ctx context.Context, req *connect.Request[outboxv1.ListDestinationTestResultsRequest]) (*connect.Response[outboxv1.ListDestinationTestResultsResponse], error) {
	if m.listDestinationTestResults != nil {
		return m.listDestinationTestResults(ctx, req)
	}
	panic("mockDestinationClient.ListDestinationTestResults not set")
}

func (m *mockDestinationClient) ValidateDestinationFilter(ctx context.Context, req *connect.Request[outboxv1.ValidateDestinationFilterRequest]) (*connect.Response[outboxv1.ValidateDestinationFilterResponse], error) {
	if m.validateDestinationFilter != nil {
		return m.validateDestinationFilter(ctx, req)
	}
	panic("mockDestinationClient.ValidateDestinationFilter not set")
}

func (m *mockDestinationClient) PollEvents(ctx context.Context, req *connect.Request[outboxv1.PollEventsRequest]) (*connect.Response[outboxv1.PollEventsResponse], error) {
	if m.pollEvents != nil {
		return m.pollEvents(ctx, req)
	}
	panic("mockDestinationClient.PollEvents not set")
}
