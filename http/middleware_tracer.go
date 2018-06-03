package http

import (
	"net/http"

	sdklog "github.com/hauxe/gom/log"
	"github.com/hauxe/gom/trace"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/zap"
)

// TracerMiddleWare http tracer middleware
type TracerMiddleWare struct {
	Handler http.Handler
	Client  *trace.Client
	Logger  sdklog.Factory
}

func (tracer *TracerMiddleWare) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Try to join to a trace propagated in request.
	wireContext, err := tracer.Client.Tracer.Extract(
		opentracing.TextMap,
		opentracing.HTTPHeadersCarrier(r.Header),
	)
	if err != nil &&
		err != opentracing.ErrSpanContextNotFound &&
		err != opentracing.ErrUnsupportedFormat {
		tracer.Logger.For(r.Context()).Fatal("error encountered while trying to extract span",
			zap.Error(err))
	}
	if wireContext != nil {
		// create span
		span := tracer.Client.Tracer.StartSpan("middleware tracer", ext.RPCServerOption(wireContext))

		// store span in context
		ctx := opentracing.ContextWithSpan(r.Context(), span)

		// update request context to include our new span
		r = r.WithContext(ctx)
		span.Finish()
	}
	tracer.Handler.ServeHTTP(w, r)
}
