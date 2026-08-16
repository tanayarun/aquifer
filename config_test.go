package aquifer_test

import (
	"errors"
	"testing"

	"github.com/tanayarun/aquifer"
)

func TestInvalidConfig(t *testing.T) {
	_, err := aquifer.New(
		fakeDial(),
		aquifer.WithMaxConns(0),
	)
	if !errors.Is(err, aquifer.ErrInvalidConfig) {
		t.Errorf("want ErrInvalidConfig, got %v", err)
	}
}

func TestMinExceedsMax(t *testing.T) {
	_, err := aquifer.New(
		fakeDial(),
		aquifer.WithMinConns(10),
		aquifer.WithMaxConns(3),
	)
	if !errors.Is(err, aquifer.ErrInvalidConfig) {
		t.Errorf("want ErrInvalidConfig, got %v", err)
	}
}
