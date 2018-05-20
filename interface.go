package sdk

import "go.uber.org/zap/zapcore"

// Server interface defines server functions
type Server interface {
	Start(options ...func() error) error
	Stop() error
}

// Client interface defines client functions
type Client interface {
	Connect(options ...func() error) error
	Disconnect() error
}

// Sender interface defines message sender
type Sender interface {
	Send(msg []byte, to string) error
}

// Logger is a simplified abstraction of the zap.Logger
type Logger interface {
	Info(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)
	With(fields ...zapcore.Field) Logger
}
