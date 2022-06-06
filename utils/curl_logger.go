package utils

import (
	"strings"

	xutils "github.com/azhai/xgen/utils"
)

func GetLoggerIgnoreError(logger *xutils.Logger, err error) *xutils.Logger {
	return logger
}

// CurlLogger 日志
type CurlLogger struct {
	*xutils.Logger
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
