package hello

import (
	"context"
	"database/sql"
	otelgo "github.com/QuangTung97/otelwrap/internal/generate/hello/otel"
	otelgosdk "github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk"
	"time"
)

// User ...
type User struct {
	ID        int64
	Name      string
	CreatedAt time.Time
	IsValid   sql.NullBool
}

// Processor ...
type Processor interface {
	DoA(ctx context.Context, n int) error
	Handle(ctx context.Context, u *User) error
	Get(ctx context.Context, id int64, content otelgosdk.Content) (otelgo.Person, error)
	NoName(context.Context, int)
	ManyParams(ctx context.Context, params ...string)
	UseArray(ctx context.Context, contents []*otelgosdk.Content) (User, error)
	UseMap(ctx context.Context, m map[otelgosdk.Content]otelgosdk.Content) map[User]User
}

// Simple ...
type Simple interface {
	Handle(ctx context.Context, u *User) error
}
