package vision

import (
	"context"
	"errors"
)

// Coordinates represents pixel coordinates on the screen.
type Coordinates struct {
	X int
	Y int
}

// Resolver resolves locate text descriptions to pixel coordinates
// via an external vision service.
type Resolver interface {
	Resolve(ctx context.Context, locate string, screenshot []byte) (Coordinates, error)
}

// Sentinel errors for the vision resolver.
var (
	ErrServiceUnavailable = errors.New("vision service unavailable")
	ErrElementNotFound    = errors.New("element not found")
	ErrInvalidResponse    = errors.New("invalid response from vision service")
)
