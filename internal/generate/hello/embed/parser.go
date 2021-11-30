package embed

import "context"

// Scanner ...
type Scanner interface {
	Scan(ctx context.Context, n int) error
}

// Parser ...
type Parser interface {
	Compute(ctx context.Context, x string) error
}
