package otelwrap

import (
	"bytes"
	"github.com/QuangTung97/otelwrap/internal/generate/hello"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ hello.Processor

func TestFindAndGenerate(t *testing.T) {
	var buf bytes.Buffer
	err := findAndGenerate(&buf, CommandArgs{
		Dir:      ".",
		Filename: "command_test.go",
		Name:     "hello.Simple",
	})
	assert.Equal(t, nil, err)
	expected := `
package otelwrap

import (
	"github.com/QuangTung97/otelwrap/internal/generate/hello"
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// SimpleWrapper wraps OpenTelemetry's span
type SimpleWrapper struct {
	hello.Simple
	tracer trace.Tracer
	prefix string
}

// NewSimpleWrapper creates a wrapper
func NewSimpleWrapper(wrapped hello.Simple, tracer trace.Tracer, prefix string) *SimpleWrapper {
	return &SimpleWrapper{
		Simple: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// Handle ...
func (w *SimpleWrapper) Handle(ctx context.Context, u *hello.User) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "Handle")
	defer span.End()

	err = w.Simple.Handle(ctx, u)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
`
	assert.Equal(t, expected, buf.String())
}
