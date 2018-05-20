package grpc

import (
	"github.com/pkg/errors"

	lib "github.com/hauxe/gom/library"
	sdklog "github.com/hauxe/gom/log"
	"github.com/hauxe/gom/trace"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	g "google.golang.org/grpc"
)

const (
	// default grpc client config
	clientHost = "0.0.0.0"
	clientPort = 10000
)

// StartClientOptions type indicates start client options
type StartClientOptions func() error

// ClientConfig defines GRPC client config properties
type ClientConfig struct {
	Host string
	Port int
}

// Client defines GRPC client properties
type Client struct {
	Config      *ClientConfig
	C           *g.ClientConn
	Logger      sdklog.Factory
	TraceClient *trace.Client
	DialOptions []g.DialOption
}

// CreateClient creates GRPC client
func CreateClient() (client *Client, err error) {
	config := ClientConfig{clientHost, clientPort}
	logger, err := sdklog.NewFactory()
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "get logger"))
	}
	return &Client{
		Config:      &config,
		Logger:      logger,
		DialOptions: []g.DialOption{g.WithInsecure()},
	}, nil
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
	url := lib.GetURL(c.Config.Host, c.Config.Port)
	c.C, err = g.Dial(url, c.DialOptions...)

	return err
}

// Disconnect disconnect client
func (c *Client) Disconnect() error {
	if c.C != nil {
		c.C.Close()
	}
	return nil
}

// SetHostPortOption set client host port
func (c *Client) SetHostPortOption(host string, port int) StartClientOptions {
	return func() (err error) {
		c.Config.Host = host
		c.Config.Port = port
		return nil
	}
}

// SetTracerOption set tracer
func (c *Client) SetTracerOption(tracer *trace.Client) StartClientOptions {
	return func() (err error) {
		c.TraceClient = tracer
		return nil
	}
}

// SetMiddlewareTracerOption set grpc tracer middleware
func (c *Client) SetMiddlewareTracerOption() StartClientOptions {
	return func() error {
		if c.TraceClient == nil {
			return errors.New("option SetTracerOption must be set first")
		}
		dialOption := g.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(c.TraceClient.Tracer))
		c.DialOptions = append(c.DialOptions, dialOption)
		return nil
	}
}
