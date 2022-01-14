package embed

import (
	"context"
	"time"
)

// ScannerInfo ...
type ScannerInfo struct {
	Name string
}

// Scanner ...
type Scanner interface {
	Scan(ctx context.Context, n int) error
	Convert(ctx context.Context, d time.Duration)
	SetInfo(ctx context.Context, info ScannerInfo)
}

// Parser ...
type Parser interface {
	Compute(ctx context.Context, x string) error
}
