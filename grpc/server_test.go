package grpc

import (
	"errors"
	"testing"

	"google.golang.org/grpc"

	"github.com/stretchr/testify/require"
	context "golang.org/x/net/context"
)

type test struct {
}

func (t *test) Test(_ context.Context, req *Request) (*Response, error) {
	if req.Name == "error" {
		return &Response{
			Code:    -1,
			Content: req.Content,
		}, errors.New("test new error")
	}
	return &Response{
		Code:    1,
		Content: req.Content,
	}, nil
}

func TestServer(t *testing.T) {
	t.Parallel()
	testSrv := test{}
	// create server
	server, err := CreateServer()
	require.Nil(t, err)
	require.NotNil(t, server)
	// start server
	require.Nil(t, server.Start([]RegisterService{
		func(s *grpc.Server) error {
			RegisterTestSrvServer(s, &testSrv)
			return nil
		},
	}, server.SetMiddlewarePoolWorkerOption(10)))
	defer server.Stop()
	// register test service

	client, err := CreateClient()
	require.Nil(t, err)
	require.NotNil(t, client)
	require.Nil(t, client.Connect(client.SetHostPortOption(server.Config.Host, server.Config.Port)))
	testClient := NewTestSrvClient(client.C)
	// test success response
	req := Request{
		Name:    "success",
		Content: "test success content",
	}
	resp, err := testClient.Test(context.Background(), &req)
	require.Nil(t, err)
	require.Equal(t, resp.Code, int32(1))
	require.Equal(t, resp.Content, req.Content)
	// test error response
	req = Request{
		Name:    "error",
		Content: "test success content",
	}
	resp, err = testClient.Test(context.Background(), &req)
	require.Error(t, err)
	require.Nil(t, resp)
}
