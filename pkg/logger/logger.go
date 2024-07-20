package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(debug bool) *zap.Logger {
	var config zap.Config
	if debug {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}
	zapLogger, _ := config.Build()
	return zapLogger
}
