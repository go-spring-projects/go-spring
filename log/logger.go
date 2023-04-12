package log

import (
	"path/filepath"

	"go.uber.org/zap"
)

type Logger = zap.Logger

var defLogger = zap.Must(zap.NewDevelopment())

func SetLogger(logger *zap.Logger) {
	defLogger = logger
}

func GetLogger(typeName string) *Logger {
	return defLogger.Named(filepath.Base(typeName))
}
