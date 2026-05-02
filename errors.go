package aquifer

import (
	"errors"
	"fmt"
)

var ErrClosed = errors.New("aquifer: pool has been shutdown")

var ErrExhausted = errors.New("aquifer: pool is at max capacity and no connection bacame available")

var ErrInvalidConfig = errors.New("aquifer: config validation failed")

type DialError struct {
	Addr string
	Err  error
}

func (e *DialError) Error() string {
	return fmt.Sprintf("aquifer: dial %s: %v", e.Addr, e.Err)
}

func (e *DialError) Unwrap() error {
	return e.Err
}
