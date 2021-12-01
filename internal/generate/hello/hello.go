package hello

import (
	"context"
	"database/sql"
	"github.com/QuangTung97/otelwrap/internal/generate/hello/embed"
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

// Timer ...
type Timer interface {
	StartTimer(ctx context.Context, d int32)
}

// Processor ...
type Processor interface {
	Timer
	embed.Scanner
	embed.Parser

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
	embed.Scanner

	Handle(ctx context.Context, u *User) error
	Variadic(ctx context.Context, names ...string)
}
