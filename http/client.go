package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/hauxe/gom/environment"

	"github.com/opentracing/opentracing-go"

	lib "github.com/hauxe/gom/library"
	sdklog "github.com/hauxe/gom/log"
	"github.com/hauxe/gom/trace"
	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
)

// StartClientOptions type indicates start client options
type StartClientOptions func() error

// SendClientOptions type indicates send client options
type SendClientOptions func(*RequestOption) error

const (
	timeout         = 32
	tlsVerification = false
)

// ClientConfig contains default config for http client
type ClientConfig struct {
	Timeout         int  `env:"HTTP_CLIENT_TIMEOUT"`
	TLSVerification bool `env:"HTTP_CLIENT_TLS_VERIFICATION"`
}

// RequestOption contains optional header, query, body, timeout of the request
type RequestOption struct {
	Header    map[string]interface{}
	Query     map[string]interface{}
	Body      io.Reader
	Timeout   time.Duration
	Transport *http.Transport
	headerMux sync.Mutex
}

// Client defines GRPC client properties
type Client struct {
	Config      *ClientConfig
	TraceClient *trace.Client
	Logger      sdklog.Factory
}

// CreateClient creates GRPC client
func CreateClient(options ...environment.CreateENVOptions) (client *Client, err error) {
	env, err := environment.CreateENV(options...)
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create server", "create env"))
	}
	config := ClientConfig{timeout, tlsVerification}
	if err = env.Parse(&config); err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "parse env"))
	}
	logger, err := sdklog.NewFactory()
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "get logger"))
	}
	return &Client{Config: &config, Logger: logger}, nil
}

// Connect create client connection
func (c *Client) Connect(options ...StartClientOptions) (err error) {
	if c.Config == nil {
		return errors.New(lib.StringTags("connect client", "config not found"))
	}
	for _, op := range options {
		if err = op(); err != nil {
			return errors.Wrap(err, lib.StringTags("connect client", "option error"))
		}
	}
	return err
}

// Disconnect disconnect client
func (c *Client) Disconnect() error {
	return nil
}

// SetTracerOption set tracer
func (c *Client) SetTracerOption(tracer *trace.Client) StartClientOptions {
	return func() (err error) {
		c.TraceClient = tracer
		return nil
	}
}

// Send sends general request to a URL and returns HTTP response
func (c *Client) Send(ctx context.Context, method string, url string,
	options ...SendClientOptions) (res *http.Response, err error) {
	var request *http.Request
	if c.TraceClient != nil {
		ctx, err = c.TraceClient.StartTracing(ctx,
			trace.Tag(string(ext.HTTPMethod), method),
			trace.Tag(string(ext.HTTPUrl), url))
		if err != nil {
			return nil, errors.Wrap(err, lib.StringTags("client send", "trace error"))
		}
		// Inject the Span context into the outgoing HTTP Request.
		if err := c.TraceClient.Tracer.Inject(
			opentracing.SpanFromContext(ctx).Context(),
			opentracing.TextMap,
			opentracing.HTTPHeadersCarrier(request.Header),
		); err != nil {
			return nil, errors.Wrap(err, lib.StringTags("client send", "error encountered while trying to inject span"))
		}
		defer func(res *http.Response) {
			if res != nil {
				tags := []opentracing.StartSpanOption{trace.Tag(string(ext.HTTPStatusCode), res.StatusCode)}
				if res.StatusCode != http.StatusOK {
					body, err := ReadBodyString(res)
					if err != nil {
						c.Logger.For(ctx).Fatal(err.Error())
						return
					}
					tags = append(tags, trace.Tag("http.body", body))
				}
				c.TraceClient.StopTracing(ctx, err, tags...)
			}
		}(res)
	}
	requestOption := &RequestOption{}
	for _, op := range options {
		if err = op(requestOption); err != nil {
			return nil, errors.Wrap(err, lib.StringTags("client send", "option error"))
		}
	}
	request, err = http.NewRequest(method, url, requestOption.Body)

	if err != nil {
		return nil, err
	}

	if requestOption.Query != nil {
		q := request.URL.Query()
		for key, val := range requestOption.Query {
			q.Add(key, lib.ToString(val))
		}
		request.URL.RawQuery = q.Encode()
	}

	if requestOption.Header != nil {
		for key, val := range requestOption.Header {
			request.Header.Set(key, lib.ToString(val))
		}
	}
	timeout := requestOption.Timeout
	if timeout <= 0 {
		timeout = time.Duration(c.Config.Timeout) * time.Second
	}
	transport := requestOption.Transport
	if transport == nil {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.Config.TLSVerification},
		}
	}
	request.WithContext(ctx)
	client := &http.Client{Timeout: timeout, Transport: transport}
	res, err = client.Do(request)
	return res, err
}

// SetRequestOptionJSON set request json
func (c *Client) SetRequestOptionJSON(body interface{}) SendClientOptions {
	return func(ro *RequestOption) error {
		// set header json
		err := c.SetRequestOptionHeader(map[string]interface{}{
			HeaderContentType: ContentTypeJSON,
		})(ro)
		if err != nil {
			return errors.Wrap(err, lib.StringTags("set request option",
				"set json header"))
		}
		data, err := json.Marshal(body)
		if err != nil {
			return errors.Wrap(err, lib.StringTags("set request option",
				fmt.Sprintf("error marshaling body: %v", body)))
		}
		ro.Body = bytes.NewReader(data)
		return nil
	}
}

// SetRequestOptionQuery set request query
func (c *Client) SetRequestOptionQuery(query map[string]interface{}) SendClientOptions {
	return func(ro *RequestOption) error {
		ro.Query = query
		return nil
	}
}

// SetRequestOptionHeader set request header
func (c *Client) SetRequestOptionHeader(headers map[string]interface{}) SendClientOptions {
	return func(ro *RequestOption) error {
		ro.headerMux.Lock()
		defer ro.headerMux.Unlock()
		if ro.Header == nil {
			ro.Header = make(map[string]interface{}, len(headers))
		}
		for key, header := range headers {
			ro.Header[key] = header
		}
		return nil
	}
}

// SetRequestOptionTimeout set request timeout in seconds
func (c *Client) SetRequestOptionTimeout(timeout int) SendClientOptions {
	return func(ro *RequestOption) error {
		ro.Timeout = time.Duration(timeout) * time.Second
		return nil
	}
}

// SetRequestOptionTransport set request transport type
func (c *Client) SetRequestOptionTransport(transport *http.Transport) SendClientOptions {
	return func(ro *RequestOption) error {
		ro.Transport = transport
		return nil
	}
}

// ParseJSON parses response body to json type
func (c *Client) ParseJSON(resp *http.Response, dest interface{}) error {
	if resp == nil {
		return errors.New("response is nil")
	}
	if resp.Body == nil {
		return errors.New("body is nil")
	}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	d.UseNumber()
	return d.Decode(dest)
}

// ReadBodyString read response body as string
func ReadBodyString(r *http.Response) (string, error) {
	if r == nil {
		return "", errors.New("response is nil")
	}
	if r.Body == nil {
		return "", nil
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
