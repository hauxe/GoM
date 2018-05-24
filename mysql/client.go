package mysql

import (
	"fmt"
	"time"

	// import mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/hauxe/gom/environment"
	lib "github.com/hauxe/gom/library"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// ConnectClientOptions type indicates start client options
type ConnectClientOptions func() error

const mysqlDriver = "mysql"

const (
	host                      = "0.0.0.0"
	port                      = 3306
	schema                    = "default_db"
	username                  = "username"
	password                  = "password"
	option                    = "charset=utf8&parseTime=True&loc=Local&multiStatements=True"
	connectionLifetimeSeconds = 300
	maxIdleConnections        = 1
	maxOpenConnections        = 1
)

// Config contains config data to connect to MySQL database
type Config struct {
	Host                      string `env:"HOST"`
	Port                      int    `env:"PORT"`
	Schema                    string `env:"SCHEMA"`
	UserName                  string `env:"USERNAME"`
	Password                  string `env:"PASSWORD"`
	Option                    string `env:"OPTION"`
	ConnectionLifetimeSeconds int    `env:"CONNECTION_LIFE_TIME_SECONDS"`
	MaxIdleConnections        int    `env:"MAX_IDLE_CONNECTIONS"`
	MaxOpenConnections        int    `env:"MAX_OPEN_CONNECTIONS"`
}

// Client  stores a client to connect to redis
type Client struct {
	Config *Config
	C      *sqlx.DB
}

// CreateClient create mqtt client
func CreateClient(options ...environment.CreateENVOptions) (client *Client, err error) {
	env, err := environment.CreateENV(options...)
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "create env"))
	}
	config := Config{host, port, schema, username, password, option,
		connectionLifetimeSeconds, maxIdleConnections, maxOpenConnections}
	if err = env.Parse(&config); err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "parse env"))
	}
	return &Client{Config: &config}, nil
}

// Connect connect client
func (c *Client) Connect(options ...ConnectClientOptions) error {
	if c.Config == nil {
		return errors.New(lib.StringTags("connect client", "config not found"))
	}
	for _, op := range options {
		if err := op(); err != nil {
			return errors.Wrap(err, lib.StringTags("connect client", "option error"))
		}
	}
	source := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", c.Config.UserName, c.Config.Password,
		c.Config.Host, c.Config.Port, c.Config.Schema, c.Config.Option)
	db, err := sqlx.Connect(mysqlDriver, source)
	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(time.Duration(c.Config.ConnectionLifetimeSeconds) * time.Second)
	db.SetMaxIdleConns(c.Config.MaxIdleConnections)
	db.SetMaxOpenConns(c.Config.MaxOpenConnections)
	c.C = db
	return nil
}

// Disconnect disconnect client
func (c *Client) Disconnect() error {
	if c.C != nil {
		return c.C.Close()
	}
	return nil
}
