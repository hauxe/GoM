package http

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	server, err := CreateServer()
	require.Nil(t, err)
	require.NotNil(t, server)

	routes := []ServerRoute{
		ServerRoute{
			Name:   "test1",
			Method: http.MethodGet,
			Path:   "/test1",
			Validators: []ParamValidator{
				func(_ context.Context, dst interface{}) error {
					d, ok := dst.(data)
					if !ok {
						return errors.New("validator failed")
					}
					if d.Field1 == "error" {
						return errors.New("validation failed")
					}
					return nil
				},
			},
			Handler: func(w http.ResponseWriter, r *http.Request) {
				dest := data{}
				err := ParseParameters(r, &dest)
				if err != nil {
					err = SendResponse(w, http.StatusBadRequest, ErrorCodeValidationFailed, "failed", map[string]interface{}{
						"error": dest,
					})
				} else {
					err = SendResponse(w, http.StatusOK, ErrorCodeSuccess, "success", map[string]interface{}{
						"success": dest,
					})
				}
				require.Nil(t, err)
			},
		},
	}
	server.Start(server.SetHandlerOption(routes...),
		server.SetMiddlewareWorkerPoolOption(10))
	// server.Start(server.SetHandlerOption(routes...))
	defer server.Stop()
	client, err := CreateClient()
	require.Nil(t, err)
	require.NotNil(t, client)
	require.Nil(t, client.Connect())
	defer client.Disconnect()
	url := "http://" + server.S.Addr + "/test1"
	// send request
	field1 := "error"
	field2 := int64(12345)
	field3 := true
	fieldRequire := true
	resp, err := client.Send(context.Background(), http.MethodGet,
		url, client.SetRequestOptionQuery(map[string]interface{}{
			"field1":        field1,
			"field2":        field2,
			"field3":        field3,
			"field_require": fieldRequire,
		}))
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	dest := response{}
	decoder := json.NewDecoder(resp.Body)
	// numbers are represented as string instead of float64
	decoder.UseNumber()
	err = decoder.Decode(&dest)
	require.Nil(t, err)
	require.Equal(t, ErrorCodeValidationFailed, dest.ErrorCode)
	d := dest.Data.Error
	require.Equal(t, field1, d.Field1)
	require.Equal(t, field2, d.Field2)
	require.Equal(t, field3, d.Field3)
	require.Equal(t, fieldRequire, d.FieldRequire)
}
