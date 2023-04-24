package utils

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danztran/telescope/pkg/httpclient"
)

// IsErrNotFound check if an error is an http not found type
func IsErrNotFound(err error) bool {
	var errNotFound *httpclient.ErrNotFound
	return errors.As(err, &errNotFound)
}

// SinceTime calculate, round and format time since a timestamp
func SinceTime(ts time.Time, round time.Duration) string {
	dur := time.Since(ts)
	if round != 0 {
		dur = dur.Round(round)
	}
	return fmt.Sprintf("%s ago", dur)
}

// WaitToStop wait for signals (INT, TERM)
func WaitToStop() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
