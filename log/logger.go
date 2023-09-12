package log

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

type Logger = slog.Logger

var defaultLogger atomic.Value

func init() {
	slogOptions := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if slog.SourceKey == attr.Key {
				source := attr.Value.Any().(*slog.Source)
				idx := strings.LastIndexByte(source.File, '/')
				if idx == -1 {
					return attr
				}
				// Find the penultimate separator.
				idx = strings.LastIndexByte(source.File[:idx], '/')
				if idx == -1 {
					return attr
				}

				source.File = source.File[idx+1:]
			}

			return attr
		},
	}
	defaultLogger.Store(slog.New(slog.NewTextHandler(os.Stdout, slogOptions)))
}

func SetLogger(logger *Logger) {
	defaultLogger.Store(logger)
}

func GetLogger(typeName string) *Logger {
	logger := defaultLogger.Load().(*Logger)
	return logger.With("logger", filepath.Base(typeName))
}
