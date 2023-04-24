package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// MustGetLogger returns a *zap.SugaredLogger, panic if error exists
func MustGetLogger(module string) *zap.SugaredLogger {
	log, err := GetLogger(module)
	if err != nil {
		panic(err)
	}
	return log
}

func GetLogger(module string) (*zap.SugaredLogger, error) {
	levels := []string{
		os.Getenv("LOG_LEVEL_" + strings.ToUpper(module)),
		os.Getenv("LOG_LEVEL"),
		"debug",
	}

	var logLevel string
	for _, lvl := range levels {
		if lvl != "" {
			logLevel = lvl
			break
		}
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.NameKey = ""
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Local().Format(time.RFC3339))
	})
	if os.Getenv("LOG_COLOR") == "true" {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config := zap.NewProductionConfig()
	config.Encoding = "console"
	config.EncoderConfig = encoderCfg

	if err := config.Level.UnmarshalText([]byte(logLevel)); err != nil {
		return nil, fmt.Errorf("error parse log level / %w", err)
	}
	log, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("error build logger / %w", err)
	}

	return log.Named(module).Sugar(), nil
}

func LogDuration() func(*zap.SugaredLogger, string, ...interface{}) {
	ts := time.Now()
	return func(log *zap.SugaredLogger, template string, args ...interface{}) {
		message := fmt.Sprintf(template, args...)
		log.Desugar().
			WithOptions(zap.AddCallerSkip(1)).
			Sugar().
			Debugf("%s took: %v", message, time.Since(ts))
	}
}
