package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hauxe/gom/environment"

	"github.com/pkg/errors"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	lib "github.com/hauxe/gom/library"
	sdklog "github.com/hauxe/gom/log"
	"github.com/hauxe/gom/pool"
	"github.com/hauxe/gom/trace"
	"go.uber.org/zap"
	g "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// StartServerOptions type indicates start server options
type StartServerOptions func() error

const (
	// default grpc server config
	serverHost              = "0.0.0.0"
	serverPort              = 10000
	enableServiceDisconvery = false
	readTimeout             = 500 // milliseconds
)

// ServerConfig defines GRPC sever config value
type ServerConfig struct {
	Host                   string `env:"GRPC_SERVER_HOST"`
	Port                   int    `env:"GRPC_SERVER_PORT"`
	EnableServiceDiscovery bool   `env:"GRPC_SERVER_SERVICE_DISCOVERY"`
	ReadTimeout            int    `env:"GRPC_SERVER_READ_TIMEOUT"`
}

// Server defines GRPC server properties
type Server struct {
	Config        *ServerConfig
	Conn          net.Listener
	S             *g.Server
	Logger        sdklog.Factory
	TraceClient   *trace.Client
	ServerOptions []g.ServerOption
}

// CreateServer creates GRPC server
func CreateServer(options ...environment.CreateENVOptions) error) (server *Server, err error) {
	env, err := environment.CreateENV(options...)
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create server", "create env"))
	}
	config := ServerConfig{serverHost, serverPort, enableServiceDisconvery, readTimeout}
	if err = env.Parse(&config); err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create server", "parse env"))
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "get logger"))
	}
	return &Server{Config: &config, Logger: sdklog.Factory{Logger: logger}}, nil
}

// Start starts running grpc server
func (s *Server) Start(options ...StartServerOptions) (err error) {
	if s.Config == nil {
		return errors.New(lib.StringTags("start server", "config not found"))
	}
	for _, op := range options {
		if err = op(); err != nil {
			return errors.Wrap(err, lib.StringTags("start server", "option error"))
		}
	}
	url := lib.GetURL(s.Config.Host, s.Config.Port)
	s.Conn, err = net.Listen("tcp", url)
	if err != nil {
		return errors.Errorf("[%s]failed to listen grpc: %v", url, err)
	}
	s.S = g.NewServer()
	if s.Config.EnableServiceDiscovery {
		// Register reflection service on gRPC server.
		reflection.Register(s.S)
	}
	if err := s.S.Serve(s.Conn); err != nil {
		return errors.Errorf("Failed to serve: %v", err)
	}
	log.Println(fmt.Sprintf("Started gRPC server at: %s:%d, with service discovery enabling is %v",
		s.Config.Host, s.Config.Port, s.Config.EnableServiceDiscovery))
	return nil
}

// Stop stops grpc server
func (s *Server) Stop() error {
	if s.S != nil {
		s.S.Stop()
	}
	if s.Conn != nil {
		return s.Conn.Close()
	}
	return nil
}

// SetServiceDiscoveryOption set service discovery option
func (s *Server) SetServiceDiscoveryOption(val bool) StartServerOptions {
	return func() (err error) {
		s.Config.EnableServiceDiscovery = val
		return nil
	}
}

// SetTracerOption set tracer
func (s *Server) SetTracerOption(tracer *trace.Client) StartServerOptions {
	return func() (err error) {
		s.TraceClient = tracer
		return nil
	}
}

// SetMiddlewareTracerOption set grpc tracer middleware
func (s *Server) SetMiddlewareTracerOption() StartServerOptions {
	return StartServerOptions {
		if s.TraceClient == nil {
			return errors.New("option SetTracerOption must be set first")
		}
		serverOption := g.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(s.TraceClient.Tracer))
		s.ServerOptions = append(s.ServerOptions, serverOption)
		return nil
	}
}

// SetMiddlewarePoolWorkerOption set grpc pool worker middleware
func (s *Server) SetMiddlewarePoolWorkerOption(worker *pool.Worker) StartServerOptions {
	return func() error {
		if s.TraceClient == nil {
			return errors.New("option SetTracerOption must be set first")
		}
		interceptor := func(ctx context.Context, req interface{}, info *g.UnaryServerInfo, handler g.UnaryHandler) (resp interface{}, err error) {
			job := JobHandler{
				name:    info.FullMethod,
				ctx:     ctx,
				req:     req,
				handler: handler,
			}
			err = worker.QueueJob(&job, time.Duration(s.Config.ReadTimeout)*time.Millisecond)
			if err != nil {
				return nil, errors.Wrap(err, lib.StringTags("pool worker middleware"))
			}
			result := <-job.c
			return result.resp, result.err
		}
		serverOption := g.UnaryInterceptor(interceptor)
		s.ServerOptions = append(s.ServerOptions, serverOption)
		return nil
	}
}
