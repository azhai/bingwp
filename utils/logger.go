package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/azhai/gozzo/logging"
	"go.uber.org/zap"
)

var (
	infoStr = "%s\n[info] "
	warnStr = "%s\n[warn] "
	errStr  = "%s\n[error] "
)

// Logger goe日志
type Logger struct {
	SlowThreshold             time.Duration
	IgnoreRecordNotFoundError bool
	*zap.SugaredLogger
}

// NewLogger 创建日志
func NewLogger(filename string) *Logger {
	l := logging.NewLoggerURL("info", filename)
	return WrapLogger(l)
}

// WrapLogger 封装日志
func WrapLogger(l *zap.SugaredLogger) *Logger {
	if l == nil {
		l = zap.NewNop().Sugar()
	}
	s := 200 * time.Millisecond
	return &Logger{SlowThreshold: s, SugaredLogger: l}
}

// InfoContext print info
func (l *Logger) InfoContext(_ context.Context, msg string, data ...any) {
	preStr := fmt.Sprintf(infoStr, logging.FileWithLineNum())
	l.SugaredLogger.Infof(preStr+msg, data...)
}

// WarnContext print warn messages
func (l *Logger) WarnContext(_ context.Context, msg string, data ...any) {
	preStr := fmt.Sprintf(warnStr, logging.FileWithLineNum())
	l.SugaredLogger.Warnf(preStr+msg, data...)
}

// ErrorContext print error messages
func (l *Logger) ErrorContext(_ context.Context, msg string, data ...any) {
	preStr := fmt.Sprintf(errStr, logging.FileWithLineNum())
	l.SugaredLogger.Errorf(preStr+msg, data...)
}
