package retry

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/cenkalti/backoff"
	lib "github.com/hauxe/gom/library"
)

// Func retry functions type
type Func func(*Client) error

// Client defines retry client properties
type Client struct {
	C backoff.BackOff
}

// CreateClient create retry client by type
func CreateClient(t Func) (*Client, error) {
	client := Client{}
	err := t(&client)
	if err != nil {
		return nil, errors.Wrap(err, lib.StringTags("create client", "run retry func"))
	}
	return &client, nil
}

// UseConstantRetry setup constant retry
func UseConstantRetry(d time.Duration) Func {
	return func(c *Client) error {
		c.C = backoff.NewConstantBackOff(d)
		return nil
	}
}

// UseExponentialRetry setup exponential retry
func UseExponentialRetry() Func {
	return func(c *Client) error {
		c.C = backoff.NewExponentialBackOff()
		return nil
	}
}

// Init init the retry client
func (c *Client) Init(options ...func() error) error {
	if c.C == nil {
		return errors.New("retry client doesnt be created")
	}
	if err := lib.RunOptionalFunc(options...); err != nil {
		return errors.Wrap(err, lib.StringTags("init client", "option error"))
	}
	return nil
}

// SetContextOption set backoff context
func (c *Client) SetContextOption(ctx context.Context) func() error {
	return func() error {
		if ctx == nil {
			return errors.New("receive nil context")
		}
		c.C = backoff.WithContext(c.C, ctx)
		return nil
	}
}

// SetMaxRetriesOption set backoff max retries
func (c *Client) SetMaxRetriesOption(maxRetries uint64) func() error {
	return func() error {
		c.C = backoff.WithMaxRetries(c.C, maxRetries)
		return nil
	}
}

// SetInitialIntervalOption set backoff inital interval
func (c *Client) SetInitialIntervalOption(interval time.Duration) func() error {
	return func() error {
		exponetial, ok := c.C.(*backoff.ExponentialBackOff)
		if !ok {
			return errors.New("retry client is not an exponential backoff")
		}
		exponetial.InitialInterval = interval
		return nil
	}
}

// SetRandomizationFactorOption set backoff randomization factor
func (c *Client) SetRandomizationFactorOption(random float64) func() error {
	return func() error {
		exponetial, ok := c.C.(*backoff.ExponentialBackOff)
		if !ok {
			return errors.New("retry client is not an exponential backoff")
		}
		exponetial.RandomizationFactor = random
		return nil
	}
}

// SetMultiplierOption set backoff multiplier
func (c *Client) SetMultiplierOption(multiplier float64) func() error {
	return func() error {
		exponetial, ok := c.C.(*backoff.ExponentialBackOff)
		if !ok {
			return errors.New("retry client is not an exponential backoff")
		}
		exponetial.Multiplier = multiplier
		return nil
	}
}

// SetMaxIntervalOption set backoff max interval
func (c *Client) SetMaxIntervalOption(maxInterval time.Duration) func() error {
	return func() error {
		exponetial, ok := c.C.(*backoff.ExponentialBackOff)
		if !ok {
			return errors.New("retry client is not an exponential backoff")
		}
		exponetial.MaxInterval = maxInterval
		return nil
	}
}

// SetMaxElapsedTimeOption set backoff max elapsed time
func (c *Client) SetMaxElapsedTimeOption(maxElapsedTime time.Duration) func() error {
	return func() error {
		exponetial, ok := c.C.(*backoff.ExponentialBackOff)
		if !ok {
			return errors.New("retry client is not an exponential backoff")
		}
		exponetial.MaxElapsedTime = maxElapsedTime
		return nil
	}
}

// Do execute operation
func (c *Client) Do(operation func() error) <-chan bool {
	ch := make(chan bool, 1)
	go func(cha chan<- bool) {
		err := backoff.Retry(operation, c.C)
		cha <- err == nil
		close(cha)
	}(ch)
	return ch
}
