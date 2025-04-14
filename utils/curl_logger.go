package utils

import (
	"strings"

	"go.uber.org/zap"
)

// CurlLogger 日志
type CurlLogger struct {
	*zap.SugaredLogger
}

func (l *CurlLogger) SetPrefix(prefix string) {
	l.Named(strings.TrimSpace(prefix))
}

func (l CurlLogger) Printf(format string, args ...any) {
	l.Debugf(format, args...)
}

func (l CurlLogger) Println(args ...any) {
	l.Debug(args...)
}
