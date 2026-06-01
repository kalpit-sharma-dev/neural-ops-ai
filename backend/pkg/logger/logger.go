package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a structured JSON logger for the given service name.
func New(serviceName string) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = zapcore.DebugLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	return zap.New(core, zap.AddCaller(), zap.Fields(
		zap.String("service", serviceName),
	)), nil
}
