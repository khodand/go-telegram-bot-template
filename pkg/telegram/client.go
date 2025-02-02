package telegram

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"

	"go-template/pkg/logger"
)

const loggerKey = "logger"

type Config struct {
	Token string
}

func NewClient(logger *zap.Logger, cfg Config) (*tele.Bot, error) {
	const timeout = 10
	client, err := tele.NewBot(tele.Settings{
		Token:   cfg.Token,
		Poller:  &tele.LongPoller{Timeout: timeout * time.Second},
		OnError: onError(),
	})
	if err != nil {
		return nil, err
	}

	client.Use(mwLogger(logger))
	client.Use(mwRecover())

	return client, nil
}

func mwLogger(initLogger *zap.Logger) func(next tele.HandlerFunc) tele.HandlerFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			logger := initLogger.With(zap.String("trace", uuid.New().String()))
			logger.Info("handle update", zap.String("text", c.Text()), zap.String("data", c.Data()))

			start := time.Now()
			setLogger(c, logger)
			err := next(c)
			elapsed := time.Since(start)

			logger = logger.With(zap.Stringer("elapsed", elapsed), zap.Int64("elapsed_nanos", int64(elapsed)))
			if err != nil {
				logger.Error("update failed", zap.Error(err))
				return c.Send("Internal Error")
			}

			logger.Info("update processed")
			return nil
		}
	}
}

func mwRecover() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					if rErr, ok := r.(error); ok {
						err = fmt.Errorf("panic: %w", rErr)
					} else {
						err = fmt.Errorf("panic: %v", r)
					}
				}
			}()

			return next(c)
		}
	}
}

func onError() func(err error, c tele.Context) {
	return func(err error, c tele.Context) {
		GetLogger(c).Error("attention! post-middleware unexpected error", zap.Error(err))
	}
}

func setLogger(c tele.Context, logger *zap.Logger) {
	c.Set(loggerKey, logger)
}

func GetLogger(c tele.Context) *zap.Logger {
	if c != nil {
		log, ok := c.Get(loggerKey).(*zap.Logger)
		if ok && log != nil {
			return log
		}
	}

	return logger.New(false).With(zap.String("logger", "default"))
}
