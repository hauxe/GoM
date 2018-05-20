package trace

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"go.uber.org/zap"

	"github.com/stretchr/testify/require"
)

func f1(ctx context.Context, tracer *Client) (err error) {
	ctx, err = tracer.StartTracing(ctx, Tag("test1", "value1"))
	if err != nil {
		return err
	}
	defer tracer.StopTracing(ctx, err)
	err = f2(ctx, tracer)
	return err
}
func f2(ctx context.Context, tracer *Client) (err error) {
	ctx, err = tracer.StartTracing(ctx, Tag("test2", "value2"))
	if err != nil {
		return err
	}
	defer tracer.StopTracing(ctx, err)
	err = f3(ctx, tracer)
	return err
}
func f3(ctx context.Context, tracer *Client) (err error) {
	ctx, err = tracer.StartTracing(ctx, Tag("test1", "value1"))
	if err != nil {
		return err
	}
	defer tracer.StopTracing(ctx, err)
	err = errors.New("test propagated f3 error")
	tracer.Logger.For(ctx).Error("log error", zap.Error(err))
	return err
}

func TestTrace(t *testing.T) {
	t.Parallel()
	receive := make(chan bool, 1)
	routeReportTrace := ServerRoute{
		Path: "/api/v1/spans",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			receive <- true
		},
	}

	server := CreateSampleServer(routeReportTrace)
	tracer, err := CreateClient()
	require.Nil(t, err)
	require.NotNil(t, tracer)
	require.Nil(t, tracer.Connect(tracer.SetServiceNameOption("dkjadasdlkasjdkjlasjdaks"),
		tracer.SetHostPortOption("0.0.0.0", 0),
		tracer.SetHTTPCollectorOption(server.URL+routeReportTrace.Path)))
	ctx := context.Background()
	f1(ctx, tracer)
	require.Nil(t, tracer.Disconnect())
	require.True(t, <-receive)
}
