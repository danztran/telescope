package server

import (
	"errors"
	"net/http"

	"github.com/danztran/telescope/pkg/handler"
	"github.com/danztran/telescope/pkg/httpclient"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *server) setupAPIs(e *echo.Echo) error {
	e.GET("/health", func(c echo.Context) error { return c.String(http.StatusOK, "OK") })
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	v1Public := e.Group("/v1/public")
	v1Public.GET("/mesh", wrapHandler(s.getAllConnections))
	v1Public.GET("/mesh/:name", wrapHandler(s.getConnectionsByName))

	return nil
}

func (s *server) getConnectionsByName(c echo.Context) error {
	name := c.Param("name")

	ctx := c.Request().Context()
	data, err := s.handler.GetConnectionsByName(ctx, name)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, data)
}

func (s *server) getAllConnections(c echo.Context) error {
	ctx := c.Request().Context()
	opt := new(handler.GetAllNodesOptions)

	if err := c.Bind(opt); err != nil {
		return err
	}

	data, err := s.handler.GetAllConnections(ctx, *opt)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, data)
}

func wrapHandler(hl func(echo.Context) error) func(echo.Context) error {
	return func(c echo.Context) error {
		err := hl(c)
		if err != nil {
			return catchHandlerError(c, err)
		}
		return nil
	}
}

func catchHandlerError(c echo.Context, err error) error {
	var errNotFound *httpclient.ErrNotFound

	switch true {
	case errors.As(err, &errNotFound):
		err = c.String(http.StatusNotFound, err.Error())
	default:
		c.Error(err)
	}

	return err
}
