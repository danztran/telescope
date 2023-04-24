package httpclient

import (
	"fmt"
)

type ErrInternal struct {
	Message string
}

func (e *ErrInternal) Error() string {
	return fmt.Sprintf("internal error / %s", e.Message)
}

type ErrClient struct {
	Message string
}

func (e *ErrClient) Error() string {
	return fmt.Sprintf("client error / %s", e.Message)
}

type ErrNotFound struct {
	Message string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("not found / %s", e.Message)
}
