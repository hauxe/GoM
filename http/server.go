package http

import (
	"fmt"
	"time"

	"github.com/hauxe/gom/environment"

	"github.com/gorilla/schema"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"net/http"

	lib "github.com/hauxe/gom/library"
	sdklog "github.com/hauxe/gom/log"
	"github.com/hauxe/gom/pool"
	"github.com/hauxe/gom/trace"
)

var decoder = schema.NewDecoder()

const (
	// default http server config
	serverHost   = "0.0.0.0"
	serverPort   = 8000
	serveTLS     = false
	certFile     = ""
	keyFile      = ""
	readTimeout  = 32
	writeTimeout = 64
)

// ServerConfig defines HTTP sever config value
type ServerConfig struct {
	Host         string `env:"HTTP_SERVER_HOST"`
	Port         int    `env:"HTTP_SERVER_PORT"`
	ServeTLS     bool   `env:"HTTP_SERVER_TLS"`
	CertFile     string `env:"HTTP_SERVER_CERT"`
	KeyFile      string `env:"HTTP_SERVER_KEY"`
	ReadTimeout  int    `env:"HTTP_SERVER_READ_TIMEOUT"`
	WriteTimeout int    `env:"HTTP_SERVER_WRITE_TIMEOUT"`
}

// Server defines HTTP server properties
type Server struct {
	Config      *ServerConfig
	S           *http.Server
	Handler     http.Handler
	Mux         *http.ServeMux
	TraceClient *trace.Client
	Logger      sdklog.Factory
}

// CreateServer creates HTTP server
func CreateServer(options ...func(*environment.ENVConfig) error) (server *Server, err error) {
	env, err := environment.CreateENV(options...)
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create server", "create env"))
	}
	config := ServerConfig{
		serverHost,
		serverPort,
		serveTLS,
		certFile,
		keyFile,
		readTimeout,
		writeTimeout}
	if err = env.Parse(&config); err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create server", "parse env"))
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "get logger"))
	}
	return &Server{Config: &config, Logger: sdklog.Factory{Logger: logger}}, nil
}

// Start starts running http server
func (s *Server) Start(options ...func() error) (err error) {
	if s.Config == nil {
		return errors.New(lib.StringTags("start server", "config not found"))
	}
	if err = s.InitHandler(); err != nil {
		return err
	}
	if err = lib.RunOptionalFunc(options...); err != nil {
		return errors.Wrap(err, lib.StringTags("start server", "option error"))
	}
	decoder.IgnoreUnknownKeys(true)
	decoder.ZeroEmpty(false)
	address := lib.GetURL(s.Config.Host, s.Config.Port)
	if s.Config.ReadTimeout <= 0 {
		s.Config.ReadTimeout = readTimeout
	}

	if s.Config.WriteTimeout <= 0 {
		s.Config.WriteTimeout = writeTimeout
	}
	readTimeout := time.Duration(s.Config.ReadTimeout) * time.Second
	writeTimeout := time.Duration(s.Config.WriteTimeout) * time.Second

	s.S = &http.Server{
		Addr:         address,
		Handler:      s.Handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
	s.Logger.Bg().Info(fmt.Sprintf("Starting HTTP server at: %s:%d",
		s.Config.Host, s.Config.Port))
	if s.Config.ServeTLS {
		if err = s.S.ListenAndServeTLS(s.Config.CertFile, s.Config.KeyFile); err != nil {
			s.Logger.Bg().Info(fmt.Sprintf("Error start TLS HTTP server at: %s:%d",
				s.Config.Host, s.Config.Port))
		}
	} else {
		if err = s.S.ListenAndServe(); err != nil {
			s.Logger.Bg().Info(fmt.Sprintf("Error start HTTP server at: %s:%d",
				s.Config.Host, s.Config.Port))
		}
	}

	return nil
}

// Stop stops http server
func (s *Server) Stop() error {
	if s.S != nil {
		return s.S.Close()
	}
	return nil
}

//InitHandler initializes route handler
func (s *Server) InitHandler() error {
	s.Mux = http.NewServeMux()
	s.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		SendResponse(w, http.StatusOK, ErrorCodeSuccess, "reached default handler", nil)
	})
	s.Handler = http.HandlerFunc(s.Mux.ServeHTTP)
	return nil
}

// SetTimeoutOption set http server timeout
func (s *Server) SetTimeoutOption(read, write int) func() error {
	return func() (err error) {
		s.Config.ReadTimeout = read
		s.Config.WriteTimeout = write
		return nil
	}
}

// SetTLSOption set http server tls info
func (s *Server) SetTLSOption(serveTLS bool, certFile, keyFile string) func() error {
	return func() (err error) {
		s.Config.ServeTLS = serveTLS
		s.Config.CertFile = certFile
		s.Config.KeyFile = keyFile
		return nil
	}
}

// SetTracerOption set tracer
func (s *Server) SetTracerOption(tracer *trace.Client) func() error {
	return func() (err error) {
		s.TraceClient = tracer
		return nil
	}
}

// SetHandlerOption set http server route handler
func (s *Server) SetHandlerOption(routes ...ServerRoute) func() error {
	return func() (err error) {
		if s.TraceClient == nil {
			return errors.New("option SetTracerOption must be called first")
		}
		for _, route := range routes {
			s.Mux.HandleFunc(route.Path, buildRouteHandler(route.Method, route.Validators, route.Handler))
			s.Logger.Bg().Info("Registered route", zap.String("name", route.Name),
				zap.String("method", route.Method), zap.String("path", route.Path))
		}
		return nil
	}
}

// SetMiddlewareTracerOption set http server middleware type tracer
func (s *Server) SetMiddlewareTracerOption() func() error {
	return func() (err error) {
		if s.TraceClient == nil {
			return errors.New("option SetTracerOption must be set first")
		}
		s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to join to a trace propagated in request.
			wireContext, err := s.TraceClient.Tracer.Extract(
				opentracing.TextMap,
				opentracing.HTTPHeadersCarrier(r.Header),
			)
			if err != nil &&
				err != opentracing.ErrSpanContextNotFound &&
				err != opentracing.ErrUnsupportedFormat {
				s.Logger.For(r.Context()).Fatal("error encountered while trying to extract span",
					zap.Error(err))
			}
			if wireContext != nil {
				// create span
				span := s.TraceClient.Tracer.StartSpan("middleware tracer", ext.RPCServerOption(wireContext))

				// store span in context
				ctx := opentracing.ContextWithSpan(r.Context(), span)

				// update request context to include our new span
				r = r.WithContext(ctx)
				span.Finish()
			}
			s.Handler.ServeHTTP(w, r)
		})
		return nil
	}
}

// SetMiddlewareWorkerPoolOption set http server uses worker pool
func (s *Server) SetMiddlewareWorkerPoolOption(worker *pool.Worker) func() error {
	return func() (err error) {
		s.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			job := JobHandler{
				name:    "handle http request",
				r:       r,
				w:       w,
				handler: s.Handler.ServeHTTP,
			}
			worker.QueueJob(&job, time.Duration(s.Config.ReadTimeout)*time.Second)
		})
		return nil
	}
}
