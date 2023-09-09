package log

import (
	"path/filepath"

	"go.uber.org/zap"
)

type Logger = zap.Logger

var defLogger = zap.Must(zap.Config{
	Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
	Development:      true,
	Encoding:         "console",
	EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
	OutputPaths:      []string{"stderr"},
	ErrorOutputPaths: []string{"stderr"},
}.Build())

func SetLogger(logger *zap.Logger) {
	defLogger = logger
}

func GetLogger(typeName string) *Logger {
	return defLogger.Named(filepath.Base(typeName))
}
