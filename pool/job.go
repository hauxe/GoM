package pool

import (
	"context"
)

// Job defines job interface
type Job interface {
	Name() string
	Execute() error
	GetContext() context.Context
}
