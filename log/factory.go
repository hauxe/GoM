// Copyright (c) 2017 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"context"

	sdk "github.com/hauxe/gom"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Factory is the default logging wrapper that can create
// logger instances either for a given Context or context-less.
type Factory struct {
	Logger *zap.Logger
}

// NewFactory creates a new Factory.
func NewFactory() (Factory, error) {
	logger, err := zap.NewDevelopment(zap.AddCaller(), zap.AddCallerSkip(1))
	return Factory{Logger: logger}, err

}

// Bg creates a context-unaware logger.
func (b Factory) Bg() sdk.Logger {
	return logger{Logger: b.Logger}
}

// For returns a context-aware Logger. If the context
// contains an OpenTracing span, all logging calls are also
// echo-ed into the span.
func (b Factory) For(ctx context.Context) sdk.Logger {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		// TODO for Jaeger span extract trace/span IDs as fields
		return spanLogger{span: span, Logger: b.Logger}
	}
	return b.Bg()
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (b Factory) With(fields ...zapcore.Field) Factory {
	return Factory{Logger: b.Logger.With(fields...)}
}
