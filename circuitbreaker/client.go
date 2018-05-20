package circuitbreaker

import (
	"context"
	"net/http"
	"time"

	"github.com/cep21/circuit/closers/hystrix"
	"github.com/cep21/circuit/metriceventstream"
	"go.uber.org/zap"

	"github.com/cep21/circuit"
	lib "github.com/hauxe/gom/library"
	sdklog "github.com/hauxe/gom/log"
	"github.com/pkg/errors"
)

// CreateClientOptions type indicates create client options
type CreateClientOptions func(*circuit.Manager) error

// StartClientOptions type indicates start client options
type StartClientOptions func(*circuit.Config, *hystrix.ConfigureCloser, *hystrix.ConfigureOpener) error

// Client defines circuit breaker client properties
type Client struct {
	manager *circuit.Manager
	Logger  sdklog.Factory
}

// CreateClient create new circuit client
func CreateClient(options ...CreateClientOptions) (*Client, error) {
	manager := circuit.Manager{}
	for _, op := range options {
		if err := op(&manager); err != nil {
			return nil, errors.Wrap(err, lib.StringTags("create env", "option error"))
		}
	}
	logger, err := sdklog.NewFactory()
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "get logger"))
	}
	return &Client{manager: &manager, Logger: logger}, nil
}

// SetDefaultCircuitProperty set default circuit constructor
func SetDefaultCircuitProperty(cons circuit.CommandPropertiesConstructor) CreateClientOptions {
	return func(manager *circuit.Manager) error {
		manager.DefaultCircuitProperties = append(manager.DefaultCircuitProperties, cons)
		return nil
	}
}

// Start with a circuit name
func (c *Client) Start(name string, options ...StartClientOptions) error {
	// default is no timeout no max concurrent request
	config := circuit.Config{
		Execution: circuit.ExecutionConfig{
			Timeout:               time.Duration(-1),
			MaxConcurrentRequests: -1,
		},
		Fallback: circuit.FallbackConfig{
			MaxConcurrentRequests: -1,
		},
		General: circuit.GeneralConfig{},
	}
	closerConfig := hystrix.ConfigureCloser{}
	openerConfig := hystrix.ConfigureOpener{}
	for _, op := range options {
		if err := op(&config, &closerConfig, &openerConfig); err != nil {
			return errors.Wrap(err, lib.StringTags("start client", "option error"))
		}
	}
	config.General = circuit.GeneralConfig{
		OpenToClosedFactory: hystrix.CloserFactory(closerConfig),
		ClosedToOpenFactory: hystrix.OpenerFactory(openerConfig),
	}
	_, err := c.manager.CreateCircuit(name, config)
	if err != nil {
		errors.Wrap(err, lib.StringTags("start client", "create circuit"))
	}
	return nil
}

// SetExecuteTimeoutOption set execution timeout config
func (c *Client) SetExecuteTimeoutOption(timeout time.Duration) StartClientOptions {
	return func(config *circuit.Config, _ *hystrix.ConfigureCloser, _ *hystrix.ConfigureOpener) error {
		config.Execution.Timeout = timeout
		return nil
	}
}

// SetExecuteMaxConcurrentRequestsOption set execution max concurrent config
func (c *Client) SetExecuteMaxConcurrentRequestsOption(maxConn int64) StartClientOptions {
	return func(config *circuit.Config, _ *hystrix.ConfigureCloser, _ *hystrix.ConfigureOpener) error {
		config.Execution.MaxConcurrentRequests = maxConn
		return nil
	}
}

// SetExecuteIgnoreInteruptsOption set execution ignore interupts
func (c *Client) SetExecuteIgnoreInteruptsOption(ignored bool) StartClientOptions {
	return func(config *circuit.Config, _ *hystrix.ConfigureCloser, _ *hystrix.ConfigureOpener) error {
		config.Execution.IgnoreInterrputs = ignored
		return nil
	}
}

// SetFallbackDisable set fallback disable/enable
func (c *Client) SetFallbackDisable(disable bool) StartClientOptions {
	return func(config *circuit.Config, _ *hystrix.ConfigureCloser, _ *hystrix.ConfigureOpener) error {
		config.Fallback.Disabled = disable
		return nil
	}
}

// SetFallbackMaxConcurrentRequestsOption set fallback max concurrent request
func (c *Client) SetFallbackMaxConcurrentRequestsOption(maxConn int64) StartClientOptions {
	return func(config *circuit.Config, _ *hystrix.ConfigureCloser, _ *hystrix.ConfigureOpener) error {
		config.Fallback.MaxConcurrentRequests = maxConn
		return nil
	}
}

// SetCloserSleepWindow set closer sleep windown
func (c *Client) SetCloserSleepWindow(sleepWindow time.Duration) StartClientOptions {
	return func(_ *circuit.Config, closer *hystrix.ConfigureCloser, _ *hystrix.ConfigureOpener) error {
		closer.SleepWindow = sleepWindow
		return nil
	}
}

// SetCloserAttempts set closer attempts
func (c *Client) SetCloserAttempts(attempts int64) StartClientOptions {
	return func(_ *circuit.Config, closer *hystrix.ConfigureCloser, _ *hystrix.ConfigureOpener) error {
		closer.HalfOpenAttempts = attempts
		return nil
	}
}

// SetCloserRequiredSuccessful set closer required successful request before close
func (c *Client) SetCloserRequiredSuccessful(require int64) StartClientOptions {
	return func(_ *circuit.Config, closer *hystrix.ConfigureCloser, _ *hystrix.ConfigureOpener) error {
		closer.RequiredConcurrentSuccessful = require
		return nil
	}
}

// SetOpenerErrorThreshold set opener error threshold
func (c *Client) SetOpenerErrorThreshold(threshold int64) StartClientOptions {
	return func(_ *circuit.Config, _ *hystrix.ConfigureCloser, opener *hystrix.ConfigureOpener) error {
		opener.ErrorThresholdPercentage = threshold
		return nil
	}
}

// SetOpenerRequestThreshold set opener request threshold
func (c *Client) SetOpenerRequestThreshold(threshold int64) StartClientOptions {
	return func(_ *circuit.Config, _ *hystrix.ConfigureCloser, opener *hystrix.ConfigureOpener) error {
		opener.RequestVolumeThreshold = threshold
		return nil
	}
}

// SetOpenerRollingDuration set opener rolling duration
func (c *Client) SetOpenerRollingDuration(rollingDuration time.Duration) StartClientOptions {
	return func(_ *circuit.Config, _ *hystrix.ConfigureCloser, opener *hystrix.ConfigureOpener) error {
		opener.RollingDuration = rollingDuration
		return nil
	}
}

// SetOpenerNumBuckets set opener number of buckets
func (c *Client) SetOpenerNumBuckets(buckets int) StartClientOptions {
	return func(_ *circuit.Config, _ *hystrix.ConfigureCloser, opener *hystrix.ConfigureOpener) error {
		opener.NumBuckets = buckets
		return nil
	}
}

// Execute execute based on circuit
func (c *Client) Execute(ctx context.Context, name string,
	runFunc func(context.Context) error, fallbackFunc func(context.Context, error) error) error {
	circuitbreaker := c.manager.GetCircuit(name)
	if circuitbreaker == nil {
		return errors.Errorf("no circuit match name %s", name)
	}
	err := circuitbreaker.Execute(ctx, runFunc, fallbackFunc)
	if err != nil {
		return errors.Wrap(err, "execute")
	}
	return nil
}

// GetMetricsHandler enable metrics and return metrics handler
func (c *Client) GetMetricsHandler(ctx context.Context) http.Handler {
	es := metriceventstream.MetricEventStream{
		Manager: c.manager,
	}
	go func() {
		if err := es.Start(); err != nil {
			c.Logger.For(ctx).Error("circuit breaker metrics error", zap.Error(err))
		}
	}()
	return &es
}

// BadRequest construct a bad request error
func BadRequest(err error) circuit.SimpleBadRequest {
	return circuit.SimpleBadRequest{Err: err}
}
