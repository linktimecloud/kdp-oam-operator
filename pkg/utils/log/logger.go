/*
Copyright 2023 KDP(Kubernetes Data Platform).

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"kdp-oam-operator/cmd/apiserver/options"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger Log component
var Logger *zap.SugaredLogger

func init() {
	Logger = zap.NewNop().Sugar()
}

func SetUp(options options.GenericOptions) {
	var zapLogLevel zap.AtomicLevel
	switch options.LogLevel {
	case "debug":
		zapLogLevel = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		zapLogLevel = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		zapLogLevel = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		zapLogLevel = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	case "fatal":
		zapLogLevel = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	default:
		zapLogLevel = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	var zc = zap.Config{
		Level:             zapLogLevel,
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "level",
			TimeKey:        "time",
			NameKey:        "name",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]interface{}{},
	}

	logger, err := zc.Build()
	if err != nil {
		panic(err)
	}
	defer func(logger *zap.Logger) {

	}(logger)
	Logger = logger.Sugar()
}
