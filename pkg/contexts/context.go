package contexts

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}

func SetLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func GetLogger(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
		return logger
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return zap.NewNop()
	}
	return logger
}
