package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	LogConfig struct {
		Skipper          func(c echo.Context) bool
		Logger           *zap.SugaredLogger
		WithRequestBody  func(c echo.Context) bool
		WithResponseBody func(c echo.Context) bool
	}
	bodyDumpWriter struct {
		io.Writer
		http.ResponseWriter
	}
)

func LogRequest(config LogConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = func(c echo.Context) bool {
			return false
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			// Request
			start := time.Now()
			req := c.Request()
			res := c.Response()
			reqBody := []byte{}
			if req.Body != nil {
				reqBody, _ = ioutil.ReadAll(req.Body)
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

			// Response
			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(res.Writer, resBody)
			writer := &bodyDumpWriter{Writer: mw, ResponseWriter: res.Writer}
			res.Writer = writer

			correlationID := req.Header.Get("x-correlation-id")
			req.Header.Set("x-correlation-id", correlationID)
			res.Header().Set("x-correlation-id", correlationID)

			ctx := context.WithValue(req.Context(), "x-correlation-id", correlationID)
			c.SetRequest(req.WithContext(ctx))

			err := next(c)
			end := time.Since(start)

			message := fmt.Sprintf(`method="%s" status="%d" uri="%s" latency_human="%s" x-correlation-id="%s"`,
				req.Method, res.Status, req.RequestURI, end.String(), correlationID)

			if config.WithRequestBody != nil && config.WithRequestBody(c) {
				message += fmt.Sprintf(` request_body="%s"`, string(reqBody))
			}

			if config.WithResponseBody != nil && config.WithResponseBody(c) {
				message += fmt.Sprintf(` response_body="%s"`, strings.TrimSuffix(resBody.String(), "\n"))
			}

			logFn := config.Logger.Infof
			if err != nil {
				logFn = config.Logger.Errorf
				message += fmt.Sprintf(` error="%s"`, err)
			}

			logFn("%s", message)

			return err
		}
	}
}

func (w *bodyDumpWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *bodyDumpWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
