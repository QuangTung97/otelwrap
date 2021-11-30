package embed

import (
	"context"
	"time"
)

// Scanner ...
type Scanner interface {
	Scan(ctx context.Context, n int) error
	Convert(ctx context.Context, d time.Duration)
}

// Parser ...
type Parser interface {
	Compute(ctx context.Context, x string) error
}
