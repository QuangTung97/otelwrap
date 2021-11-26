package otelwrap

import (
	"bytes"
	"errors"
	"github.com/QuangTung97/otelwrap/internal/generate/hello"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ hello.Processor

func TestFindAndGenerate_Interface_From_Another(t *testing.T) {
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

func TestFindAndGenerate_Same_Package_Not_Found(t *testing.T) {
	var buf bytes.Buffer
	err := findAndGenerate(&buf, CommandArgs{
		Dir:      ".",
		Filename: "command_test.go",
		Name:     "Example",
	})
	assert.Equal(t, errors.New("can not find interface 'Example'"), err)
}

func TestFindAndGenerate_Same_Package_OK(t *testing.T) {
	var buf bytes.Buffer
	err := findAndGenerate(&buf, CommandArgs{
		Dir:      ".",
		Filename: "command_test.go",
		Name:     "Sample",
	})
	assert.Equal(t, nil, err)
	expected := `
package otelwrap

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// SampleWrapper wraps OpenTelemetry's span
type SampleWrapper struct {
	Sample
	tracer trace.Tracer
	prefix string
}

// NewSampleWrapper creates a wrapper
func NewSampleWrapper(wrapped Sample, tracer trace.Tracer, prefix string) *SampleWrapper {
	return &SampleWrapper{
		Sample: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// Get ...
func (w *SampleWrapper) Get(ctx context.Context) (a int, err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "Get")
	defer span.End()

	a, err = w.Sample.Get(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return a, err
}
`
	assert.Equal(t, expected, buf.String())
}
