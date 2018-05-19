package redis

import (
	"strings"

	"github.com/hauxe/gom/environment"

	"github.com/go-redis/redis"
	lib "github.com/hauxe/gom/library"
	"github.com/pkg/errors"
)

// ConnectClientOptions type indicates connect client options
type ConnectClientOptions func() error

const (
	separator          = "|"
	standAloneServer   = ""
	sentinelServers    = "0.0.0.0:26379|0.0.0.0:26380|0.0.0.0:26381"
	sentinelMasterName = "master"
	password           = "password"
	db                 = 0
)

// ClientConfig contains configuration to connect to redis
type ClientConfig struct {
	Separator          string `env:"REDIS_SEPARATOR"`
	StandAloneServer   string `env:"REDIS_STAND_ALONE_SERVER"`
	SentinelServers    string `env:"REDIS_SENTINEL_SERVERS"`
	SentinelMasterName string `env:"REDIS_SENTINAL_MASTER_NAME"`
	Password           string `env:"REDIS_PASSWORD"`
	DB                 int    `env:"REDIS_DB"`
}

// Client  stores a client to connect to redis
type Client struct {
	Config *ClientConfig
	C      *redis.Client
}

// CreateClient create mqtt client
func CreateClient(options ...func(*environment.ENVConfig) error) (client *Client, err error) {
	env, err := environment.CreateENV(options...)
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "create env"))
	}
	config := ClientConfig{separator, standAloneServer, sentinelServers,
		sentinelMasterName, password, db}
	if err = env.Parse(&config); err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "parse env"))
	}
	return &Client{Config: &config}, nil
}

// Connect connect to connect to redis
func (c *Client) Connect(options ...ConnectClientOptions) (err error) {
	for _, op := range options {
		if err = op(); err != nil {
			return errors.Wrap(err, "connect redis client")
		}
	}
	if _, err = c.C.Ping().Result(); err != nil {
		return errors.Wrap(err, lib.StringTags("connect redis client", "ping client"))
	}
	return nil
}

// Disconnect closes redis client and releases any open resources.
func (c *Client) Disconnect() (err error) {
	if err = c.C.Close(); err != nil {
		return errors.Wrap(err, lib.StringTags("disconect redis client"))
	}
	return nil
}

// SetAuthOption set redis auth
func (c *Client) SetAuthOption(password string) ConnectClientOptions {
	return func() (err error) {
		c.Config.Password = password
		return nil
	}
}

// SetDBOption set redis db
func (c *Client) SetDBOption(db int) ConnectClientOptions {
	return func() (err error) {
		c.Config.DB = db
		return nil
	}
}

// ConnectFailoverClientOption connect to failover client
func (c *Client) ConnectFailoverClientOption() ConnectClientOptions {
	return func() (err error) {
		if c.Config.SentinelMasterName == "" ||
			c.Config.SentinelServers == "" {
			return errors.New("invalid failover client config")
		}
		c.C = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    c.Config.SentinelMasterName,
			SentinelAddrs: strings.Split(c.Config.SentinelServers, c.Config.Separator),
			Password:      c.Config.Password,
			DB:            c.Config.DB,
		})
		return nil
	}
}

// ConnectStandaloneClientOption connect to standalone client
func (c *Client) ConnectStandaloneClientOption() ConnectClientOptions {
	return func() (err error) {
		if c.Config.StandAloneServer == "" {
			return errors.New("invalid standalone client config")
		}
		c.C = redis.NewClient(&redis.Options{
			Addr:     c.Config.StandAloneServer,
			Password: c.Config.Password,
			DB:       c.Config.DB,
		})
		return nil
	}
}
