package echo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go-template/pkg/contexts"
	"go-template/pkg/errx"
)

type Handler interface {
	RegisterHandlers(e *echo.Echo)
}

type Server struct {
	logger *zap.Logger
	echo   *echo.Echo
}

func New(logger *zap.Logger) *Server {
	e := echo.New()
	e.Validator = &customValidator{validator: validator.New(validator.WithRequiredStructEnabled())}

	e.Use(mwLogger(logger))
	e.Use(mwErrors())
	e.Use(middleware.Recover())

	return &Server{
		logger: logger,
		echo:   e,
	}
}

func (s *Server) Start(address string, handlers ...Handler) error {
	for _, h := range handlers {
		h.RegisterHandlers(s.echo)
	}

	return s.echo.Start(address)
}

func (s *Server) Shutdown() error {
	const shutdownTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return s.echo.Shutdown(ctx)
}

func mwLogger(initLogger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			traceID := c.Request().Header.Get(echo.HeaderXRequestID)
			if traceID == "" {
				traceID = uuid.NewString()
			}

			logger := initLogger.With(zap.String("trace_id", traceID))
			c.SetRequest(c.Request().WithContext(contexts.SetLogger(c.Request().Context(), logger)))

			req := c.Request()
			fields := []zapcore.Field{
				zap.String("remote_ip", c.RealIP()),
				zap.String("host", req.Host),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.RequestURI)),
				zap.String("user_agent", req.UserAgent()),
			}
			logger.Info("Start http request", fields...)

			start := time.Now()
			err := next(c)

			logger.Info("Request processed",
				zap.Duration("processing_time", time.Since(start)),
				zap.Int("http_status", c.Response().Status),
			)

			return err
		}
	}
}

func mwErrors() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			logger := contexts.GetLogger(c.Request().Context())

			var status int
			var message string
			switch {
			case errors.Is(err, errx.ErrNotFound):
				status = http.StatusNotFound
				message = "not found"
			case errors.Is(err, errx.ErrDuplicateKey):
				status = http.StatusConflict
				message = "duplicate entity"
			case errors.Is(err, echo.ErrUnauthorized):
				status = http.StatusUnauthorized
				message = "unauthorized"
			case errors.Is(err, echo.ErrBadRequest):
				status = http.StatusBadRequest
				message = fmt.Sprintf("bad request: %s", err.Error())
			default:
				logger.Error("Request failed", zap.Error(err))
				return c.JSON(http.StatusInternalServerError, errx.HTTPError{Message: "internal server error"})
			}

			logger.Info("Request failed", zap.Error(err))
			return c.JSON(status, errx.HTTPError{Message: message})
		}
	}
}
