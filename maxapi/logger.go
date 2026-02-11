package maxapi

import (
	"context"
	"log"
)

// Logger — минимальный интерфейс логгера, который можно адаптировать под любую библиотеку.
type Logger interface {
	Debugf(ctx context.Context, format string, args ...any)
	Infof(ctx context.Context, format string, args ...any)
	Warnf(ctx context.Context, format string, args ...any)
	Errorf(ctx context.Context, format string, args ...any)
}

// noopLogger — логгер по умолчанию, ничего не пишет.
type noopLogger struct{}

func (n *noopLogger) Debugf(_ context.Context, _ string, _ ...any) {}
func (n *noopLogger) Infof(_ context.Context, _ string, _ ...any)  {}
func (n *noopLogger) Warnf(_ context.Context, _ string, _ ...any)  {}
func (n *noopLogger) Errorf(_ context.Context, _ string, _ ...any) {}

// StdLogger — адаптер над стандартным log.Logger.
type StdLogger struct {
	l *log.Logger
}

// NewStdLogger создаёт адаптер для стандартного логгера.
func NewStdLogger(l *log.Logger) *StdLogger {
	if l == nil {
		l = log.Default()
	}
	return &StdLogger{l: l}
}

func (s *StdLogger) Debugf(_ context.Context, format string, args ...any) {
	s.l.Printf("[DEBUG] "+format, args...)
}

func (s *StdLogger) Infof(_ context.Context, format string, args ...any) {
	s.l.Printf("[INFO] "+format, args...)
}

func (s *StdLogger) Warnf(_ context.Context, format string, args ...any) {
	s.l.Printf("[WARN] "+format, args...)
}

func (s *StdLogger) Errorf(_ context.Context, format string, args ...any) {
	s.l.Printf("[ERROR] "+format, args...)
}

