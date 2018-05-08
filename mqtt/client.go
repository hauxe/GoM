package mqtt

import (
	"context"

	"github.com/hauxe/gom/environment"

	"github.com/hauxe/gom/trace"

	mq "github.com/eclipse/paho.mqtt.golang"
	lib "github.com/hauxe/gom/library"
	"github.com/pkg/errors"
)

const (
	// default grpc client config
	clientHost     = "0.0.0.0"
	clientPort     = 1883
	clientUserName = "username"
	clientPassword = "password"
)

// ClientConfig defines mqtt client config properties
type ClientConfig struct {
	Host     string `env:"-"`
	Port     int    `env:"-"`
	UserName string `env:"MQTT_CLIENT_USERNAME"`
	Password string `env:"MQTT_CLIENT_PASSWORD"`
}

// Client defines mqtt client properties
type Client struct {
	Config      *ClientConfig
	C           mq.Client
	TraceClient *trace.Client
}

// CreateClient create mqtt client
func CreateClient(options ...func(*environment.ENVConfig) error) (client *Client, err error) {
	env, err := environment.CreateENV(options...)
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "create env"))
	}
	config := ClientConfig{clientHost, clientPort, clientUserName, clientPassword}
	if err = env.Parse(&config); err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "parse env"))
	}
	return &Client{Config: &config}, nil
}

// Connect connect to a mqtt server
func (c *Client) Connect(options ...func() error) (err error) {
	if c.Config == nil {
		return errors.New(lib.StringTags("connect client", "config not found"))
	}
	opts := mq.NewClientOptions()
	opts.AddBroker(lib.GetURL(c.Config.Host, c.Config.Port))
	if err = lib.RunOptionalFunc(options...); err != nil {
		return errors.Wrap(err, lib.StringTags("connect client", "option error"))
	}
	opts.SetUsername(c.Config.UserName)
	opts.SetPassword(c.Config.Password)

	c.C = mq.NewClient(opts)
	if token := c.C.Connect(); token.Wait() && token.Error() != nil {
		err := token.Error()
		return errors.Wrap(err, lib.StringTags("connect client"))
	}
	return nil
}

// Disconnect close all mqtt connections
func (c *Client) Disconnect() error {
	if c.C != nil {
		c.C.Disconnect(300)
	}
	return nil
}

// SetAuthOption set mqtt auth
func (c *Client) SetAuthOption(username, password string) func() error {
	return func() (err error) {
		c.Config.UserName = username
		c.Config.Password = password
		return nil
	}
}

// SetHostPortOption set client host port
func (c *Client) SetHostPortOption(host string, port int) func() error {
	return func() (err error) {
		c.Config.Host = host
		c.Config.Port = port
		return nil
	}
}

// SetTracerOption set tracer
func (c *Client) SetTracerOption(tracer *trace.Client) func() error {
	return func() (err error) {
		c.TraceClient = tracer
		return nil
	}
}

// Send send a message to channel
func (c *Client) Send(ctx context.Context, msg []byte, to string) (err error) {
	if c.TraceClient != nil {
		ctx, err = c.TraceClient.StartTracing(ctx,
			trace.Tag("msg", string(msg)),
			trace.Tag("to", to))
		if err != nil {
			return errors.Wrap(err, lib.StringTags("client send", "trace error"))
		}
		defer c.TraceClient.StopTracing(ctx, err)
	}
	if token := c.C.Publish(to, 0, false, msg); token.Wait() && token.Error() != nil {
		err = token.Error()
		return errors.Wrap(err, lib.StringTags("send message"))
	}
	return nil
}
