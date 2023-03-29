package otelwrap

import (
	"bytes"
	"errors"
	"github.com/QuangTung97/otelwrap/internal/generate/hello"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ hello.Simple

func TestFindAndGenerate_Interface_From_Another(t *testing.T) {
	var buf bytes.Buffer
	err := findAndGenerate(&buf, CommandArgs{
		Dir:            ".",
		SrcFileName:    "command_test.go",
		InterfaceNames: []string{"hello.Simple"},
	})
	assert.Equal(t, nil, err)
	expected := `
package otelwrap

import (
	"github.com/QuangTung97/otelwrap/internal/generate/hello"
	"context"
	"time"
	"github.com/QuangTung97/otelwrap/internal/generate/hello/embed"
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

// Scan ...
func (w *SimpleWrapper) Scan(ctx context.Context, n int) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "Scan")
	defer span.End()

	err = w.Simple.Scan(ctx, n)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

// Convert ...
func (w *SimpleWrapper) Convert(ctx context.Context, d time.Duration) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "Convert")
	defer span.End()

	w.Simple.Convert(ctx, d)
}

// SetInfo ...
func (w *SimpleWrapper) SetInfo(ctx context.Context, info embed.ScannerInfo) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "SetInfo")
	defer span.End()

	w.Simple.SetInfo(ctx, info)
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

// Variadic ...
func (w *SimpleWrapper) Variadic(ctx context.Context, names ...string) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "Variadic")
	defer span.End()

	w.Simple.Variadic(ctx, names...)
}
`
	assert.Equal(t, expected, buf.String())
}

func TestFindAndGenerate_Same_Package_Not_Found(t *testing.T) {
	var buf bytes.Buffer
	err := findAndGenerate(&buf, CommandArgs{
		Dir:            ".",
		SrcFileName:    "command_test.go",
		InterfaceNames: []string{"Example"},
	})
	assert.Equal(t, errors.New("can not find interface 'Example'"), err)
}

func TestFindAndGenerate_Same_Package_OK(t *testing.T) {
	var buf bytes.Buffer
	err := findAndGenerate(&buf, CommandArgs{
		Dir:            ".",
		SrcFileName:    "command_test.go",
		InterfaceNames: []string{"Sample"},
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

func TestFindAndGenerate_Export_In_Another(t *testing.T) {
	var buf bytes.Buffer
	err := findAndGenerate(&buf, CommandArgs{
		Dir:            ".",
		InterfaceNames: []string{"Sample", "Repo"},
		InAnother:      true,
		PkgName:        "another",
	})
	assert.Equal(t, nil, err)
	expected := `
package another

import (
	"github.com/QuangTung97/otelwrap/otelwrap"
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// SampleWrapper wraps OpenTelemetry's span
type SampleWrapper struct {
	otelwrap.Sample
	tracer trace.Tracer
	prefix string
}

// NewSampleWrapper creates a wrapper
func NewSampleWrapper(wrapped otelwrap.Sample, tracer trace.Tracer, prefix string) *SampleWrapper {
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

// RepoWrapper wraps OpenTelemetry's span
type RepoWrapper struct {
	otelwrap.Repo
	tracer trace.Tracer
	prefix string
}

// NewRepoWrapper creates a wrapper
func NewRepoWrapper(wrapped otelwrap.Repo, tracer trace.Tracer, prefix string) *RepoWrapper {
	return &RepoWrapper{
		Repo: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// Update ...
func (w *RepoWrapper) Update(ctx context.Context, id int) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "Update")
	defer span.End()

	err = w.Repo.Update(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
`
	assert.Equal(t, expected, buf.String())
}

func TestCheckInAnother(t *testing.T) {
	inAnother := CheckInAnother("hello.go")
	assert.Equal(t, false, inAnother)

	inAnother = CheckInAnother("sample/hello.go")
	assert.Equal(t, true, inAnother)

	inAnother = CheckInAnother("./hello.go")
	assert.Equal(t, false, inAnother)
}

func TestFindAndGenerate_Alias_Of_Another_Package(t *testing.T) {
	var buf bytes.Buffer
	err := findAndGenerate(&buf, CommandArgs{
		Dir:            ".",
		SrcFileName:    "command_test.go",
		InterfaceNames: []string{"HandlerAlias"},
	})
	assert.Equal(t, nil, err)
	expected := `
package otelwrap

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// HandlerAliasWrapper wraps OpenTelemetry's span
type HandlerAliasWrapper struct {
	HandlerAlias
	tracer trace.Tracer
	prefix string
}

// NewHandlerAliasWrapper creates a wrapper
func NewHandlerAliasWrapper(wrapped HandlerAlias, tracer trace.Tracer, prefix string) *HandlerAliasWrapper {
	return &HandlerAliasWrapper{
		HandlerAlias: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// Process ...
func (w *HandlerAliasWrapper) Process(ctx context.Context, n int) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "Process")
	defer span.End()

	err = w.HandlerAlias.Process(ctx, n)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
`
	assert.Equal(t, expected, buf.String())
}
