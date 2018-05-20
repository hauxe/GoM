package trace

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/hauxe/gom/environment"

	zipkin "github.com/openzipkin/zipkin-go-opentracing"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"

	lib "github.com/hauxe/gom/library"
	sdklog "github.com/hauxe/gom/log"
)

const (
	// default tracer client config
	separator   = "|"
	servers     = "0.0.0.0:2181"
	serviceName = "sdk"
	hostPort    = "0.0.0.0:0"
	debug       = false
)

// defines tracer tag names
const (
	TracerTimeTag = "time"
	TracerFuncTag = "function"
	TracerFileTag = "file"
	TracerLineTag = "line"
)

// ClientConfig defines GRPC client config properties
type ClientConfig struct {
	Separator   string `env:"TRACER_SEPARATOR"`
	Servers     string `env:"TRACER_KAFKA_SERVERS"`
	ServiceName string `env:"TRACER_SERVICE_NAME"`
	HostPort    string `env:"TRACER_HOST_PORT"`
	Debug       bool   `env:"TRACER_DEBUG"`
}

// Client defines GRPC client properties
type Client struct {
	Config    *ClientConfig
	Collector zipkin.Collector
	Tracer    opentracing.Tracer
	Logger    sdklog.Factory
}

// CreateClient creates GRPC client
func CreateClient(options ...environment.CreateENVOptions) (client *Client, err error) {
	env, err := environment.CreateENV(options...)
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create server", "create env"))
	}
	config := ClientConfig{separator, servers, serviceName, hostPort, debug}
	if err = env.Parse(&config); err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create server", "parse env"))
	}
	logger, err := sdklog.NewFactory()
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create server", "get logger"))
	}
	return &Client{Config: &config, Logger: logger}, nil
}

// Connect create client connection
func (c *Client) Connect(options ...func() error) (err error) {
	if c.Config == nil {
		return errors.New(lib.StringTags("connect client", "config not found"))
	}
	for _, op := range options {
		if err = op(); err != nil {
			return errors.Wrap(err, lib.StringTags("connect client", "option error"))
		}
	}

	// create recorder.
	recorder := zipkin.NewRecorder(c.Collector, c.Config.Debug, c.Config.HostPort, c.Config.ServiceName)

	// create tracer.
	c.Tracer, err = zipkin.NewTracer(
		recorder,
		zipkin.ClientServerSameSpan(false),
		zipkin.TraceID128Bit(true),
	)
	if err != nil {
		return errors.Wrap(err, lib.StringTags("connect client", "unable to create Zipkin tracer"))
	}

	// explicitly set our tracer to be the default tracer.
	opentracing.InitGlobalTracer(c.Tracer)
	return nil
}

// Disconnect disconnect client
func (c *Client) Disconnect() error {
	if c.Collector != nil {
		return c.Collector.Close()
	}
	return nil
}

// SetHostPortOption set tracing host and port option
func (c *Client) SetHostPortOption(host string, port int) func() error {
	return func() (err error) {
		c.Config.HostPort = lib.GetURL(host, port)
		return nil
	}
}

// SetDebugOption set tracing debug option
func (c *Client) SetDebugOption(val bool) func() error {
	return func() (err error) {
		c.Config.Debug = val
		return nil
	}
}

// SetServiceNameOption set tracing debug option
func (c *Client) SetServiceNameOption(serviceName string) func() error {
	return func() (err error) {
		c.Config.ServiceName = serviceName
		return nil
	}
}

// SetKafkaCollectorOption set tracing kafka collector
func (c *Client) SetKafkaCollectorOption(servers string) func() error {
	return func() (err error) {
		if servers == "" {
			servers = c.Config.Servers
		}
		// create collector.
		c.Collector, err = zipkin.NewKafkaCollector(strings.Split(servers, c.Config.Separator))
		if err != nil {
			return errors.Wrap(err, lib.StringTags("connect client", "unable to create Zipkin Kafka collector"))
		}
		return nil
	}
}

// SetHTTPCollectorOption set tracing http collector
func (c *Client) SetHTTPCollectorOption(server string) func() error {
	return func() (err error) {
		if server == "" {
			return errors.New("must specify http server")
		}
		// create collector.
		c.Collector, err = zipkin.NewHTTPCollector(server)
		if err != nil {
			return errors.Wrap(err, lib.StringTags("connect client", "unable to create Zipkin HTTP collector"))
		}
		return nil
	}
}

// StartTracing starts tracing
func (c *Client) StartTracing(ctx context.Context, tags ...opentracing.StartSpanOption) (context.Context, error) {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		err := errors.New("get caller error")
		c.Logger.For(ctx).Error(err.Error())
		return nil, err
	}
	tags = append(tags, Tag(TracerTimeTag, time.Now().Format(time.RFC3339)))
	funcName := "start_tracing_func_name"
	fc := runtime.FuncForPC(pc)
	if fc != nil {
		funcName = fc.Name()
		parts := strings.Split(funcName, ".")
		funcName = parts[len(parts)-1]
		tags = append(tags, Tag(TracerFuncTag, funcName))
	}
	tags = append(tags, Tag(TracerFileTag, file),
		Tag(TracerLineTag, line))
	_, ctx = opentracing.StartSpanFromContext(ctx, funcName, tags...)
	return ctx, nil
}

// StopTracing stops tracing
func (c *Client) StopTracing(ctx context.Context, err error, tags ...opentracing.StartSpanOption) {
	if err != nil {
		c.Logger.For(ctx).Fatal(fmt.Sprintf("%+v", err))
	}
	span := opentracing.SpanFromContext(ctx)
	for _, tag := range tags {
		t, ok := tag.(opentracing.Tag)
		if ok {
			t.Set(span)
		}
	}
	if span == nil {
		c.Logger.For(ctx).Error("span not found")
		return
	}
	span.Finish()
}

// Tag build a trace tag
func Tag(key string, value interface{}) opentracing.StartSpanOption {
	return opentracing.Tag{Key: key, Value: value}
}
