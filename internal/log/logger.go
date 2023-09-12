/*
 * Copyright 2012-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
