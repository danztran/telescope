package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danztran/telescope/pkg/handler"
	"github.com/danztran/telescope/pkg/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

var defaultLogger = utils.MustGetLogger("server")

type Deps struct {
	Log     *zap.SugaredLogger
	Config  Config
	Handler handler.Handler
}

type Config struct {
	Port            string `mapstructure:"port"`
	GracefulSeconds int    `mapstructure:"graceful_seconds"`
	LogRequest      bool   `mapstructure:"log_request"`
	LogResponse     bool   `mapstructure:"log_response"`
	CORS            bool   `mapstructure:"cors"`
	Pprof           bool   `mapstructure:"pprof"`
}

type Server interface {
	Run(ctx context.Context) error
}

type server struct {
	config  Config
	log     *zap.SugaredLogger
	handler handler.Handler
}

func MustNew(deps Deps) Server {
	c, err := New(deps)
	if err != nil {
		panic(err)
	}
	return c
}

func New(deps Deps) (Server, error) {
	if deps.Handler == nil {
		return nil, fmt.Errorf("handler is required")
	}
	if deps.Log == nil {
		deps.Log = defaultLogger
	}

	s := &server{
		config:  deps.Config,
		log:     deps.Log,
		handler: deps.Handler,
	}

	return s, nil
}

func (s *server) Run(ctx context.Context) error {
	e := echo.New()

	if s.config.CORS {
		e.Use(middleware.CORS())
	}
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
	}))
	if s.config.Pprof {
		pprofWrap(e)
	}
	e.Use(NewMetric())
	e.Use(LogRequest(LogConfig{
		Logger: s.log,
		Skipper: func(c echo.Context) bool {
			return false
		},
		WithRequestBody: func(c echo.Context) bool {
			return s.config.LogRequest
		},
		WithResponseBody: func(c echo.Context) bool {
			return s.config.LogResponse && c.Request().RequestURI != "/metrics"
		},
	}))

	if err := s.setupAPIs(e); err != nil {
		return err
	}

	return s.listen(ctx, e)
}

func (s *server) listen(ctx context.Context, e *echo.Echo) error {
	done := make(chan error)

	go func() {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(
			context.TODO(),
			time.Duration(s.config.GracefulSeconds)*time.Second,
		)
		defer cancel()
		done <- e.Server.Shutdown(ctx)
	}()

	err := e.Start(fmt.Sprintf(":%s", s.config.Port))
	if err != http.ErrServerClosed {
		return err
	}

	return <-done
}
