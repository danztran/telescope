package server

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

type (
	PrometheusConfig struct {
		Subsystem string
		Skipper   middleware.Skipper
	}
)

var DefaultPrometheusConfig = PrometheusConfig{
	Subsystem: "",
	Skipper:   middleware.DefaultSkipper,
}

func NewMetric() echo.MiddlewareFunc {
	return NewMetricWithConfig(DefaultPrometheusConfig)
}

func NewMetricWithConfig(config PrometheusConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultPrometheusConfig.Skipper
	}
	echoReqDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: config.Subsystem,
			Name:      "request_duration_seconds",
			Help:      "HTTP request latencies in seconds.",
			Buckets:   []float64{.005, .01, .02, 0.04, .06, 0.08, .1, 0.15, .25, 0.4, .6, .8, 1, 1.5, 2, 3, 5},
		},
		[]string{"code", "path"},
	)
	prometheus.MustRegister(echoReqDuration)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			timestamp := time.Now()

			err := next(c)

			uri := c.Path()
			status := strconv.Itoa(res.Status)
			elapsed := time.Since(timestamp).Seconds()
			path := req.Method + "_" + uri

			echoReqDuration.WithLabelValues(status, path).Observe(elapsed)

			return err
		}
	}
}
