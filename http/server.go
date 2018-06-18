package http

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hauxe/gom/environment"

	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"net/http"

	lib "github.com/hauxe/gom/library"
	sdklog "github.com/hauxe/gom/log"
	"github.com/hauxe/gom/pool"
	"github.com/hauxe/gom/trace"
)

// StartServerOptions type indicates start server options
type StartServerOptions func() error

var decoder = schema.NewDecoder()

const (
	// default http server config
	serverHost   = "localhost"
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
	Logger      sdklog.Factory
	WorkerPools []*pool.Worker
	URL         string
	Routes      map[string]map[string]http.HandlerFunc
	routesMux   sync.RWMutex
}

// CreateServer creates HTTP server
func CreateServer(options ...environment.CreateENVOptions) (server *Server, err error) {
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
	logger, err := sdklog.NewFactory()
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "get logger"))
	}
	return &Server{Config: &config, Logger: logger}, nil
}

// Start starts running http server
func (s *Server) Start(options ...StartServerOptions) (err error) {
	if s.Config == nil {
		return errors.New(lib.StringTags("start server", "config not found"))
	}
	if err = s.InitHandler(); err != nil {
		return err
	}

	for _, op := range options {
		if err = op(); err != nil {
			return errors.Wrap(err, lib.StringTags("start server", "option error"))
		}
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
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		if s.Config.ServeTLS {
			s.URL = fmt.Sprintf("%s://%s", "https", s.S.Addr)
			wg.Done()
			if err = s.S.ListenAndServeTLS(s.Config.CertFile, s.Config.KeyFile); err != nil {
				if err != http.ErrServerClosed {
					s.Logger.Bg().Info(fmt.Sprintf("Error Start TLS HTTP server at: %s:%d",
						s.Config.Host, s.Config.Port))
					s.Logger.Bg().Error(err.Error())
				}
			}
		} else {
			s.URL = fmt.Sprintf("%s://%s", "http", s.S.Addr)
			wg.Done()
			if err = s.S.ListenAndServe(); err != nil {
				if err != http.ErrServerClosed {
					s.Logger.Bg().Info(fmt.Sprintf("Error Start HTTP server at: %s:%d",
						s.Config.Host, s.Config.Port))
					s.Logger.Bg().Error(err.Error())
				}
			}
		}
	}()
	wg.Wait()
	return nil
}

// Stop stops http server
func (s *Server) Stop() error {
	if s.S != nil {
		s.Logger.Bg().Info("shutting down")
		return s.S.Shutdown(context.Background())
	}
	for _, workerPool := range s.WorkerPools {
		workerPool.StopServer()
	}
	return nil
}

//InitHandler initializes route handler
func (s *Server) InitHandler() error {
	s.Mux = http.NewServeMux()
	s.Handler = http.HandlerFunc(s.Mux.ServeHTTP)
	return nil
}

// SetHostPortOption set http server timeout
func (s *Server) SetHostPortOption(host string, port int) StartServerOptions {
	return func() (err error) {
		s.Config.Host = host
		s.Config.Port = port
		return nil
	}
}

// SetTimeoutOption set http server timeout
func (s *Server) SetTimeoutOption(read, write int) StartServerOptions {
	return func() (err error) {
		s.Config.ReadTimeout = read
		s.Config.WriteTimeout = write
		return nil
	}
}

// SetTLSOption set http server tls info
func (s *Server) SetTLSOption(serveTLS bool, certFile, keyFile string) StartServerOptions {
	return func() (err error) {
		s.Config.ServeTLS = serveTLS
		s.Config.CertFile = certFile
		s.Config.KeyFile = keyFile
		return nil
	}
}

// SetHandlerOption set http server route handler
func (s *Server) SetHandlerOption(routes ...ServerRoute) StartServerOptions {
	return func() (err error) {
		s.Logger.Bg().Info("setting up handler")
		for _, route := range routes {
			handler := s.BuildHandler(&route)
			if handler != nil {
				s.Mux.HandleFunc(route.Path, handler)
			}
			s.Logger.Bg().Info("Registered route", zap.String("name", route.Name),
				zap.String("method", route.Method), zap.String("path", route.Path))
		}
		return nil
	}
}

// BuildHandler build http handler
func (s *Server) BuildHandler(route *ServerRoute) (handler http.HandlerFunc) {
	s.routesMux.Lock()
	defer s.routesMux.Unlock()
	if s.Routes == nil {
		s.Routes = make(map[string]map[string]http.HandlerFunc)
	}
	if _, existed := s.Routes[route.Path]; !existed {
		s.Routes[route.Path] = make(map[string]http.HandlerFunc)
		handler = func(w http.ResponseWriter, r *http.Request) {
			// set response headers
			w.Header().Set(HeaderAllowOrigin, allowOrigins)
			w.Header().Set(HeaderAllowCredentials, allowCredentials)
			w.Header().Set(HeaderExposeHeaders, exposeHeaders)

			// preflight request
			if r.Method == http.MethodOptions {
				w.Header().Set(HeaderAllowHeaders, allowHeaders)
				w.Header().Set(HeaderAllowMethods, allowMethods)
				SendResponse(w, http.StatusOK, ErrorCodeSuccess, "ok", nil)
				return
			}
			s.routesMux.RLock()
			handler, existed := s.Routes[route.Path][route.Method]
			s.routesMux.RUnlock()
			if !existed {
				SendResponse(w, http.StatusMethodNotAllowed, ErrorCodeMalformedMethod,
					"method is not correct for the requested route", nil)
				return
			}
			handler(w, r)
		}
	}
	s.Routes[route.Path][route.Method] = buildRouteHandler(route.Method, route.Validators, route.Handler)
	return
}

// SetMiddlewareTracerOption set http server middleware type tracer
func (s *Server) SetMiddlewareTracerOption(tracer *trace.Client) StartServerOptions {
	return func() (err error) {
		if tracer == nil {
			return errors.New("option SetTracerOption must be set first")
		}
		s.Handler = &TracerMiddleWare{
			Handler: s.Handler,
			Client:  tracer,
			Logger:  s.Logger,
		}
		return nil
	}
}

// SetMiddlewareWorkerPoolOption set http server uses worker pool
func (s *Server) SetMiddlewareWorkerPoolOption(maxWorkers int) StartServerOptions {
	return func() (err error) {
		workerPool, err := pool.CreateWorker()
		if err != nil {
			return err
		}
		err = workerPool.StartServer(workerPool.SetMaxWorkersOption(maxWorkers))
		if err != nil {
			return err
		}

		s.WorkerPools = append(s.WorkerPools, workerPool)
		s.Handler = &WorkerPoolMiddleware{
			Handler: s.Handler,
			Pool:    workerPool,
			Timeout: time.Duration(s.Config.ReadTimeout) * time.Second,
		}
		return nil
	}
}

// SetupWorkerPoolHandler set up fast return and do hard job on worker
func (s *Server) SetupWorkerPoolHandler(maxWorkers int, handler http.HandlerFunc) (http.HandlerFunc, error) {
	workerPool, err := pool.CreateWorker()
	if err != nil {
		return nil, err
	}
	err = workerPool.StartServer(workerPool.SetMaxWorkersOption(maxWorkers))
	if err != nil {
		return nil, err
	}
	s.WorkerPools = append(s.WorkerPools, workerPool)
	return func(w http.ResponseWriter, r *http.Request) {
		job := JobHandler{
			name:    "handle http request async",
			r:       r,
			handler: handler,
		}
		err := workerPool.QueueJob(&job, time.Duration(s.Config.ReadTimeout)*time.Second)
		if err != nil {
			err = SendError(w, err)
			if err != nil {
				s.Logger.For(r.Context()).Error("queue job error", zap.Error(err))
			}
			return
		}
		SendResponse(w, http.StatusOK, ErrorCodeSuccess, "successfully queued job", nil)
		return
	}, nil
}
