package hello

import (
	"context"
	"database/sql"
	otelgo "github.com/QuangTung97/otelwrap/internal/generate/hello/otel"
	otelgosdk "github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk"
	"time"
)

type User struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	IsValid   sql.NullBool
}

type Processor interface {
	DoA(ctx context.Context, n int) error
	Handle(ctx context.Context, u *User) error
	Get(ctx context.Context, id int64, content otelgosdk.Content) (otelgo.Person, error)
	NoName(context.Context, int)
}
