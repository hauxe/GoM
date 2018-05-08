package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	lib "github.com/hauxe/gom/library"
	"github.com/stretchr/testify/require"
)

type data struct {
	Field1       string `json:"field1,omitempty" schema:"field1"`
	Field2       int64  `json:"field2,omitempty" schema:"field2"`
	Field3       bool   `json:"field3,omitempty" schema:"field3"`
	FieldRequire bool   `json:"field_require" schema:"field_require,required"`
}

// ServerResponseData defines server response data type, we have 3 type
type responseData struct {
	Success data `json:"success,omitempty"`
	Error   data `json:"error,omitempty"`
	Others  data `json:"others,omitempty"`
}

type response struct {
	ErrorCode    errorCode        `json:"error_code"`
	ErrorMessage string           `json:"error_message"`
	Data         responseData     `json:"data"`
	Time         *lib.TimeRFC3339 `json:"time"`
}

func TestParseParameters(t *testing.T) {
	t.Parallel()
	decoder.IgnoreUnknownKeys(true)
	decoder.ZeroEmpty(false)
	routeErrorForm := ServerRoute{
		Path: "/error_form",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			dest := data{}
			err := ParseParameters(r, &dest)
			require.Error(t, err)
			err = SendResponse(w, http.StatusBadRequest, ErrorCodeFailed, err.Error(), map[string]interface{}{
				"error": dest,
			})
			require.Nil(t, err)
		},
	}
	routeErrorBody := ServerRoute{
		Path:    "/error_body",
		Handler: routeErrorForm.Handler,
	}
	routeErrorQuery := ServerRoute{
		Path:    "/error_query",
		Handler: routeErrorForm.Handler,
	}
	routeSuccessJSONEmptyValidator := ServerRoute{
		Path: "/success_empty_validator",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			dest := data{}
			err := ParseParameters(r, &dest)
			require.Nil(t, err)
			err = SendResponse(w, http.StatusOK, ErrorCodeSuccess, "success", map[string]interface{}{
				"success": dest,
			})
			require.Nil(t, err)
		},
	}
	routeErrorValidator := ServerRoute{
		Path: "/error_validator",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			validators := []ParamValidator{
				func(_ context.Context, _ interface{}) error {
					return nil
				},
				func(_ context.Context, _ interface{}) error {
					return nil
				},
				func(_ context.Context, _ interface{}) error {
					return errors.New("validator failed")
				},
			}
			ctx = context.WithValue(ctx, ContextValidatorKey, validators)
			dest := data{}
			err := ParseParameters(r.WithContext(ctx), &dest)
			require.Error(t, err)
			err = SendResponse(w, http.StatusBadRequest, ErrorCodeValidationFailed, err.Error(), map[string]interface{}{
				"error": dest,
			})
			require.Nil(t, err)
		},
	}
	routeSuccessValidator := ServerRoute{
		Path: "/success_validator",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			validators := []ParamValidator{
				func(_ context.Context, _ interface{}) error {
					return nil
				},
				func(_ context.Context, _ interface{}) error {
					return nil
				},
				func(_ context.Context, _ interface{}) error {
					return nil
				},
			}
			ctx = context.WithValue(ctx, ContextValidatorKey, validators)
			dest := data{}
			err := ParseParameters(r.WithContext(ctx), &dest)
			require.Nil(t, err)
			err = SendResponse(w, http.StatusOK, ErrorCodeSuccess, "success", map[string]interface{}{
				"success": dest,
			})
			require.Nil(t, err)
		},
	}
	server := CreateSampleServer(routeErrorForm, routeErrorBody, routeErrorQuery,
		routeSuccessJSONEmptyValidator, routeErrorValidator, routeSuccessValidator)
	t.Run("error form", func(t *testing.T) {
		t.Parallel()
		// send request
		field1 := "value1"
		field2 := int64(12345)
		field3 := true
		hc := http.Client{}
		form := url.Values{}
		form.Add("field1", "value1")
		form.Add("field2", lib.ToString(field2))
		form.Add("field3", lib.ToString(field3))
		req, err := http.NewRequest(http.MethodPost, server.URL+routeErrorForm.Path, strings.NewReader(form.Encode()))
		req.Header.Add(HeaderContentType, ContentTypeForm)
		resp, err := hc.Do(req)
		require.Nil(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		dest := response{}
		decoder := json.NewDecoder(resp.Body)
		// numbers are represented as string instead of float64
		decoder.UseNumber()
		err = decoder.Decode(&dest)
		require.Nil(t, err)
		require.Equal(t, ErrorCodeFailed, dest.ErrorCode)
		d := dest.Data.Error
		require.Equal(t, field1, d.Field1)
		require.Equal(t, field2, d.Field2)
		require.Equal(t, field3, d.Field3)
		require.False(t, d.FieldRequire)
	})

	t.Run("error body", func(t *testing.T) {
		t.Parallel()
		// send request
		hc := http.Client{}
		bodyReader := strings.NewReader("invalid json content")
		req, err := http.NewRequest(http.MethodPost, server.URL+routeErrorBody.Path, bodyReader)
		req.Header.Add(HeaderContentType, ContentTypeJSON)
		resp, err := hc.Do(req)
		require.Nil(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		dest := response{}
		decoder := json.NewDecoder(resp.Body)
		// numbers are represented as string instead of float64
		decoder.UseNumber()
		err = decoder.Decode(&dest)
		require.Nil(t, err)
		require.Equal(t, ErrorCodeFailed, dest.ErrorCode)
		d := dest.Data.Error
		require.Empty(t, d.Field1)
		require.Zero(t, d.Field2)
		require.False(t, d.Field3)
		require.False(t, d.FieldRequire)
	})

	t.Run("error query", func(t *testing.T) {
		t.Parallel()
		// send request
		field1 := "value1"
		field2 := int64(12345)
		field3 := true
		hc := http.Client{}
		req, err := http.NewRequest(http.MethodGet, server.URL+routeErrorQuery.Path, nil)
		q := req.URL.Query()
		q.Add("field1", "value1")
		q.Add("field2", lib.ToString(field2))
		q.Add("field3", lib.ToString(field3))
		req.URL.RawQuery = q.Encode()
		req.Header.Add(HeaderContentType, ContentTypeText)
		resp, err := hc.Do(req)
		require.Nil(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		dest := response{}
		decoder := json.NewDecoder(resp.Body)
		// numbers are represented as string instead of float64
		decoder.UseNumber()
		err = decoder.Decode(&dest)
		require.Nil(t, err)
		require.Equal(t, ErrorCodeFailed, dest.ErrorCode)
		d := dest.Data.Error
		require.Equal(t, field1, d.Field1)
		require.Equal(t, field2, d.Field2)
		require.Equal(t, field3, d.Field3)
		require.False(t, d.FieldRequire)
	})

	t.Run("success body empty validator", func(t *testing.T) {
		t.Parallel()
		// send request
		field1 := "value1"
		field2 := int64(12345)
		field3 := true
		fieldRequire := true
		hc := http.Client{}
		body := data{
			Field1:       field1,
			Field2:       field2,
			Field3:       field3,
			FieldRequire: fieldRequire,
		}
		b, err := json.Marshal(body)
		require.Nil(t, err)
		bodyReader := strings.NewReader(string(b))
		req, err := http.NewRequest(http.MethodPost, server.URL+routeSuccessJSONEmptyValidator.Path, bodyReader)
		req.Header.Add(HeaderContentType, ContentTypeJSON)
		resp, err := hc.Do(req)
		require.Nil(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		dest := response{}
		decoder := json.NewDecoder(resp.Body)
		// numbers are represented as string instead of float64
		decoder.UseNumber()
		err = decoder.Decode(&dest)
		require.Nil(t, err)
		require.Equal(t, ErrorCodeSuccess, dest.ErrorCode)
		d := dest.Data.Success
		require.Equal(t, field1, d.Field1)
		require.Equal(t, field2, d.Field2)
		require.Equal(t, field3, d.Field3)
		require.Equal(t, fieldRequire, d.FieldRequire)
	})

	t.Run("error validator", func(t *testing.T) {
		t.Parallel()
		// send request
		field1 := "value1"
		field2 := int64(12345)
		field3 := true
		fieldRequire := true
		hc := http.Client{}
		req, err := http.NewRequest(http.MethodGet, server.URL+routeErrorValidator.Path, nil)
		q := req.URL.Query()
		q.Add("field1", "value1")
		q.Add("field2", lib.ToString(field2))
		q.Add("field3", lib.ToString(field3))
		q.Add("field_require", lib.ToString(fieldRequire))
		req.URL.RawQuery = q.Encode()
		req.Header.Add(HeaderContentType, ContentTypeText)
		resp, err := hc.Do(req)
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
	})

	t.Run("success validator", func(t *testing.T) {
		t.Parallel()
		// send request
		field1 := "value1"
		field2 := int64(12345)
		field3 := true
		fieldRequire := true
		hc := http.Client{}
		req, err := http.NewRequest(http.MethodGet, server.URL+routeSuccessValidator.Path, nil)
		q := req.URL.Query()
		q.Add("field1", "value1")
		q.Add("field2", lib.ToString(field2))
		q.Add("field3", lib.ToString(field3))
		q.Add("field_require", lib.ToString(fieldRequire))
		req.URL.RawQuery = q.Encode()
		req.Header.Add(HeaderContentType, ContentTypeText)
		resp, err := hc.Do(req)
		require.Nil(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		dest := response{}
		decoder := json.NewDecoder(resp.Body)
		// numbers are represented as string instead of float64
		decoder.UseNumber()
		err = decoder.Decode(&dest)
		require.Nil(t, err)
		require.Equal(t, ErrorCodeSuccess, dest.ErrorCode)
		d := dest.Data.Success
		require.Equal(t, field1, d.Field1)
		require.Equal(t, field2, d.Field2)
		require.Equal(t, field3, d.Field3)
		require.Equal(t, fieldRequire, d.FieldRequire)
	})
}

func TestSendResponse(t *testing.T) {
	t.Parallel()
	routeSuccess := ServerRoute{
		Path: "/success",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			dest := data{}
			err := ParseParameters(r, &dest)
			require.Nil(t, err)
			err = SendResponse(w, http.StatusOK, ErrorCodeSuccess, "success", map[string]interface{}{
				"success": dest,
			})
			require.Nil(t, err)
		},
	}
	server := CreateSampleServer(routeSuccess)
	// send request
	field1 := "value1"
	field2 := int64(12345)
	field3 := true
	fieldRequire := true
	hc := http.Client{}
	form := url.Values{}
	form.Add("field1", "value1")
	form.Add("field2", lib.ToString(field2))
	form.Add("field3", lib.ToString(field3))
	form.Add("field_require", lib.ToString(fieldRequire))
	req, err := http.NewRequest(http.MethodPost, server.URL+routeSuccess.Path, strings.NewReader(form.Encode()))
	req.Header.Add(HeaderContentType, ContentTypeForm)
	resp, err := hc.Do(req)
	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	dest := response{}
	decoder := json.NewDecoder(resp.Body)
	// numbers are represented as string instead of float64
	decoder.UseNumber()
	err = decoder.Decode(&dest)
	require.Nil(t, err)
	require.Equal(t, ErrorCodeSuccess, dest.ErrorCode)
	require.Equal(t, "success", dest.ErrorMessage)
	d := dest.Data.Success
	require.Equal(t, field1, d.Field1)
	require.Equal(t, field2, d.Field2)
	require.Equal(t, field3, d.Field3)
	require.Equal(t, fieldRequire, d.FieldRequire)
}

func TestBuildRouteHandler(t *testing.T) {
	t.Parallel()
	routeNoValidator := ServerRoute{
		Path: "/no_validator",
		Handler: buildRouteHandler(http.MethodGet, nil, func(w http.ResponseWriter, r *http.Request) {
			dest := data{}
			err := ParseParameters(r, &dest)
			if err != nil {
				err = SendResponse(w, http.StatusBadRequest, ErrorCodeFailed, err.Error(), map[string]interface{}{
					"error": dest,
				})
			} else {
				err = SendResponse(w, http.StatusOK, ErrorCodeSuccess, "success", map[string]interface{}{
					"success": dest,
				})
			}
			require.Nil(t, err)
		}),
	}
	routeErrorValidator := ServerRoute{
		Path: "/error_validator",
		Handler: buildRouteHandler(http.MethodGet, []ParamValidator{
			func(_ context.Context, _ interface{}) error {
				return nil
			},
			func(_ context.Context, _ interface{}) error {
				return nil
			},
			func(_ context.Context, _ interface{}) error {
				return errors.New("validator failed")
			},
		}, func(w http.ResponseWriter, r *http.Request) {
			dest := data{}
			err := ParseParameters(r, &dest)
			if err != nil {
				err = SendResponse(w, http.StatusBadRequest, ErrorCodeValidationFailed, err.Error(), map[string]interface{}{
					"error": dest,
				})
			} else {
				err = SendResponse(w, http.StatusOK, ErrorCodeSuccess, "success", map[string]interface{}{
					"success": dest,
				})
			}
			require.Nil(t, err)
		}),
	}
	routeSuccessValidator := ServerRoute{
		Path: "/success_validator",
		Handler: buildRouteHandler(http.MethodGet, []ParamValidator{
			func(_ context.Context, _ interface{}) error {
				return nil
			},
			func(_ context.Context, _ interface{}) error {
				return nil
			},
			func(_ context.Context, _ interface{}) error {
				return nil
			},
		}, func(w http.ResponseWriter, r *http.Request) {
			dest := data{}
			err := ParseParameters(r, &dest)
			if err != nil {
				err = SendResponse(w, http.StatusBadRequest, ErrorCodeValidationFailed, err.Error(), map[string]interface{}{
					"error": dest,
				})
			} else {
				err = SendResponse(w, http.StatusOK, ErrorCodeSuccess, "success", map[string]interface{}{
					"success": dest,
				})
			}
			require.Nil(t, err)
		}),
	}
	server := CreateSampleServer(routeNoValidator, routeErrorValidator, routeSuccessValidator)
	t.Run("success options method", func(t *testing.T) {
		t.Parallel()
		hc := http.Client{}
		req, err := http.NewRequest(http.MethodOptions, server.URL+routeNoValidator.Path, nil)
		require.Nil(t, err)
		resp, err := hc.Do(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		dest := response{}
		decoder := json.NewDecoder(resp.Body)
		// numbers are represented as string instead of float64
		decoder.UseNumber()
		err = decoder.Decode(&dest)
		require.Nil(t, err)
		require.Equal(t, ErrorCodeSuccess, dest.ErrorCode)
		require.Equal(t, "ok", dest.ErrorMessage)
	})
	t.Run("error mismatch method", func(t *testing.T) {
		t.Parallel()
		hc := http.Client{}
		req, err := http.NewRequest(http.MethodPost, server.URL+routeNoValidator.Path, nil)
		require.Nil(t, err)
		resp, err := hc.Do(req)
		require.Nil(t, err)
		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
		dest := response{}
		decoder := json.NewDecoder(resp.Body)
		// numbers are represented as string instead of float64
		decoder.UseNumber()
		err = decoder.Decode(&dest)
		require.Nil(t, err)
		require.Equal(t, ErrorCodeMalformedMethod, dest.ErrorCode)
		require.Equal(t, "method is not correct for the requested route", dest.ErrorMessage)
	})

	t.Run("success no validator", func(t *testing.T) {
		t.Parallel()
		// send request
		field1 := "value1"
		field2 := int64(12345)
		field3 := true
		fieldRequire := true
		hc := http.Client{}
		req, err := http.NewRequest(http.MethodGet, server.URL+routeNoValidator.Path, nil)
		q := req.URL.Query()
		q.Add("field1", "value1")
		q.Add("field2", lib.ToString(field2))
		q.Add("field3", lib.ToString(field3))
		q.Add("field_require", lib.ToString(fieldRequire))
		req.URL.RawQuery = q.Encode()
		req.Header.Add(HeaderContentType, ContentTypeText)
		resp, err := hc.Do(req)
		require.Nil(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		dest := response{}
		decoder := json.NewDecoder(resp.Body)
		// numbers are represented as string instead of float64
		decoder.UseNumber()
		err = decoder.Decode(&dest)
		require.Nil(t, err)
		require.Equal(t, ErrorCodeSuccess, dest.ErrorCode)
		d := dest.Data.Success
		require.Equal(t, field1, d.Field1)
		require.Equal(t, field2, d.Field2)
		require.Equal(t, field3, d.Field3)
		require.Equal(t, fieldRequire, d.FieldRequire)
	})

	t.Run("error validator", func(t *testing.T) {
		t.Parallel()
		// send request
		field1 := "value1"
		field2 := int64(12345)
		field3 := true
		fieldRequire := true
		hc := http.Client{}
		req, err := http.NewRequest(http.MethodGet, server.URL+routeErrorValidator.Path, nil)
		q := req.URL.Query()
		q.Add("field1", "value1")
		q.Add("field2", lib.ToString(field2))
		q.Add("field3", lib.ToString(field3))
		q.Add("field_require", lib.ToString(fieldRequire))
		req.URL.RawQuery = q.Encode()
		req.Header.Add(HeaderContentType, ContentTypeText)
		resp, err := hc.Do(req)
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
	})

	t.Run("success validator", func(t *testing.T) {
		t.Parallel()
		// send request
		field1 := "value1"
		field2 := int64(12345)
		field3 := true
		fieldRequire := true
		hc := http.Client{}
		req, err := http.NewRequest(http.MethodGet, server.URL+routeSuccessValidator.Path, nil)
		q := req.URL.Query()
		q.Add("field1", "value1")
		q.Add("field2", lib.ToString(field2))
		q.Add("field3", lib.ToString(field3))
		q.Add("field_require", lib.ToString(fieldRequire))
		req.URL.RawQuery = q.Encode()
		req.Header.Add(HeaderContentType, ContentTypeText)
		resp, err := hc.Do(req)
		require.Nil(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		dest := response{}
		decoder := json.NewDecoder(resp.Body)
		// numbers are represented as string instead of float64
		decoder.UseNumber()
		err = decoder.Decode(&dest)
		require.Nil(t, err)
		require.Equal(t, ErrorCodeSuccess, dest.ErrorCode)
		d := dest.Data.Success
		require.Equal(t, field1, d.Field1)
		require.Equal(t, field2, d.Field2)
		require.Equal(t, field3, d.Field3)
		require.Equal(t, fieldRequire, d.FieldRequire)
	})
}
