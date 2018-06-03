package http

import (
	"net/http"
	"time"

	sdklog "github.com/hauxe/gom/log"
	"github.com/hauxe/gom/pool"
)

// WorkerPoolMiddleware http worker pool middleware
type WorkerPoolMiddleware struct {
	Handler http.Handler
	Pool    *pool.Worker
	Logger  sdklog.Factory
	Timeout time.Duration
}

func (worker *WorkerPoolMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	job := JobHandler{
		name:    "handle http request",
		r:       r,
		w:       w,
		handler: worker.Handler.ServeHTTP,
		c:       make(chan struct{}, 1),
	}
	worker.Pool.QueueJob(&job, worker.Timeout)
	<-job.c
	worker.Handler.ServeHTTP(w, r)
}
