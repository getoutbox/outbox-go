package main

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// ProcessMessageInput holds the structured data passed to the MessageWorkflow.
type ProcessMessageInput struct {
	ConnectorID string
	AccountID   string
	MessageID   string
	UserText    string
}

// MessageWorkflow is the Temporal workflow that processes an inbound message.
func MessageWorkflow(ctx workflow.Context, input ProcessMessageInput) error {
	opts := workflow.ActivityOptions{StartToCloseTimeout: time.Minute}
	ctx = workflow.WithActivityOptions(ctx, opts)
	return workflow.ExecuteActivity(ctx, "ProcessMessage", input).Get(ctx, nil)
}
