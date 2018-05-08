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
	sdk "github.com/hauxe/gom"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logger delegates all calls to the underlying zap.Logger
type logger struct {
	Logger *zap.Logger
}

// Info logs an info msg with fields
func (l logger) Info(msg string, fields ...zapcore.Field) {
	l.Logger.Info(msg, fields...)
}

// Error logs an error msg with fields
func (l logger) Error(msg string, fields ...zapcore.Field) {
	l.Logger.Error(msg, fields...)
}

// Fatal logs a fatal error msg with fields
func (l logger) Fatal(msg string, fields ...zapcore.Field) {
	l.Logger.Fatal(msg, fields...)
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (l logger) With(fields ...zapcore.Field) sdk.Logger {
	return logger{Logger: l.Logger.With(fields...)}
}
