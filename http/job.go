package http

import (
	"context"
	"net/http"

	"github.com/pkg/errors"
)

// JobHandler define http job properties
type JobHandler struct {
	name    string
	w       http.ResponseWriter
	r       *http.Request
	handler http.HandlerFunc
	c       chan struct{}
}

// Name get job name
func (job *JobHandler) Name() string {
	return job.name
}

// GetContext get http context
func (job *JobHandler) GetContext() context.Context {
	return job.r.Context()
}

// Execute handle job
func (job *JobHandler) Execute() error {
	if job.handler == nil {
		return errors.New("empty handler")
	}
	job.handler(job.w, job.r)
	job.c <- struct{}{}
	return nil
}
