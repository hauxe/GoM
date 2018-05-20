package grpc

import (
	"context"

	"github.com/pkg/errors"

	"google.golang.org/grpc"
)

// JobResult defines job result fields
type JobResult struct {
	resp interface{}
	err  error
}

// JobHandler define grpc job properties
type JobHandler struct {
	name    string
	ctx     context.Context
	req     interface{}
	handler grpc.UnaryHandler
	c       chan JobResult
}

// Name get job name
func (job *JobHandler) Name() string {
	return job.name
}

// GetContext get grpc context
func (job *JobHandler) GetContext() context.Context {
	return job.ctx
}

// Execute handle job
func (job *JobHandler) Execute() error {
	if job.handler == nil {
		return errors.New("empty handler")
	}
	if job.c == nil {
		return errors.New("empty result channel")
	}
	resp, err := job.handler(job.ctx, job.req)
	job.c <- JobResult{resp, err}
	return nil
}
