package trace

import (
	"context"
	"errors"
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
	tracer, err := CreateClient()
	require.Nil(t, err)
	require.NotNil(t, tracer)
	require.Nil(t, tracer.Connect(tracer.SetServiceNameOption("dkjadasdlkasjdkjlasjdaks"),
		tracer.SetHostPortOption("0.0.0.0", 0),
		tracer.SetHTTPCollectorOption("http://localhost:9411/api/v1/spans")))
	ctx := context.Background()
	f1(ctx, tracer)
	require.Nil(t, tracer.Disconnect())
}
