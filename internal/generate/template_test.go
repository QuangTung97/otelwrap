package generate

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAssignVariableNames(t *testing.T) {
	result := assignVariableNames(packageTypeInfo{
		name: "example",
		imports: []importInfo{
			{
				aliasName: "",
				path:      "context",
				usedName:  "context",
			},
		},
		interfaces: []interfaceInfo{
			{
				name: "Generator",
				methods: []methodType{
					{
						name: "Hello",
						params: []tupleType{
							{
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
							},
							{
								typeStr: "int64",
							},
							{
								typeStr: "string",
							},
						},
						results: []tupleType{
							{
								typeStr: "bool",
							},
							{
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
					{
						name: "DoA",
						params: []tupleType{
							{
								name:       "ctx",
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
							},
							{
								name:    "n",
								typeStr: "int64",
							},
						},
						results: []tupleType{
							{
								name:       "err",
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, packageTypeInfo{
		name: "example",
		imports: []importInfo{
			{
				aliasName: "",
				path:      "context",
				usedName:  "context",
			},
		},
		interfaces: []interfaceInfo{
			{
				name: "Generator",
				methods: []methodType{
					{
						name: "Hello",
						params: []tupleType{
							{
								name:       "ctx",
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
							},
							{
								name:    "a",
								typeStr: "int64",
							},
							{
								name:    "b",
								typeStr: "string",
							},
						},
						results: []tupleType{
							{
								name:    "a1",
								typeStr: "bool",
							},
							{
								name:       "err",
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
					{
						name: "DoA",
						params: []tupleType{
							{
								name:       "ctx",
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
							},
							{
								name:    "n",
								typeStr: "int64",
							},
						},
						results: []tupleType{
							{
								name:       "err",
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
				},
			},
		},
	}, result)
}

func TestCollectVariables(t *testing.T) {
	result := collectVariables(packageTypeInfo{
		name: "example",
		imports: []importInfo{
			{
				aliasName: "",
				path:      "context",
				usedName:  "context",
			},
		},
		interfaces: []interfaceInfo{
			{
				name: "Generator",
				methods: []methodType{
					{
						name: "Hello",
						params: []tupleType{
							{
								name:       "ctx",
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
							},
							{
								name:    "id",
								typeStr: "int64",
							},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, templateVariables{
		globalVariables: map[string]struct{}{
			"example":   {},
			"context":   {},
			"Generator": {},
		},
		interfaces: []templateInterfaceVariables{
			{
				name: "Generator",
				methods: []templateMethodVariables{
					{
						variables: map[string]recognizedType{
							"Hello": recognizedTypeUnknown,
							"ctx":   recognizedTypeContext,
							"id":    recognizedTypeUnknown,
						},
					},
				},
			},
		},
	}, result)
}

func TestGetVariableName(t *testing.T) {
	name := getVariableName(
		map[string]struct{}{},
		map[string]recognizedType{}, 0, recognizedTypeUnknown)
	assert.Equal(t, "a", name)

	name = getVariableName(
		map[string]struct{}{},
		map[string]recognizedType{}, 1, recognizedTypeUnknown)
	assert.Equal(t, "b", name)

	name = getVariableName(
		map[string]struct{}{},
		map[string]recognizedType{}, 0, recognizedTypeContext)
	assert.Equal(t, "ctx", name)

	name = getVariableName(
		map[string]struct{}{},
		map[string]recognizedType{}, 0, recognizedTypeError)
	assert.Equal(t, "err", name)

	name = getVariableName(
		map[string]struct{}{},
		map[string]recognizedType{
			"ctx": recognizedTypeContext,
		}, 0, recognizedTypeContext)
	assert.Equal(t, "ctx1", name)

	name = getVariableName(
		map[string]struct{}{
			"ctx": {},
		},
		map[string]recognizedType{
			"ctx1": recognizedTypeContext,
		}, 0, recognizedTypeContext)
	assert.Equal(t, "ctx2", name)
}

func TestGenerateCode(t *testing.T) {
	var buf bytes.Buffer
	err := generateCode(&buf, packageTypeInfo{
		name: "example",
		imports: []importInfo{
			{
				path:     "context",
				usedName: "context",
			},
			{
				path:     "time",
				usedName: "time",
			},
		},
		interfaces: []interfaceInfo{
			{
				name: "Handler",
				methods: []methodType{
					{
						name: "Hello",
						params: []tupleType{
							{
								name:       "ctx",
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
							{
								name:    "n",
								typeStr: "int",
							},
							{
								name:    "createdAt",
								typeStr: "time.Time",
								pkgList: []tupleTypePkg{
									{
										path:  "time",
										begin: 0,
										end:   len("time"),
									},
								},
							},
						},
						results: []tupleType{
							{
								name:       "",
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
					{
						name: "WithReturn",
						params: []tupleType{
							{
								name:       "rootCtx",
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
							},
							{
								name:    "n",
								typeStr: "int",
							},
							{
								name:    "span",
								typeStr: "string",
							},
						},
						results: []tupleType{
							{
								name:    "count",
								typeStr: "int64",
							},
							{
								name:       "err",
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, `
package example

import (
	"context"
	"time"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// HandlerWrapper wraps OpenTelemetry's span
type HandlerWrapper struct {
	Handler
	tracer trace.Tracer
	prefix string
}

// NewHandlerWrapper creates a wrapper
func NewHandlerWrapper(wrapped Handler, tracer trace.Tracer, prefix string) *HandlerWrapper {
	return &HandlerWrapper{
		Handler: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// Hello ...
func (w *HandlerWrapper) Hello(ctx context.Context, n int, createdAt time.Time) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "Hello")
	defer span.End()

	err = w.Handler.Hello(ctx, n, createdAt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

// WithReturn ...
func (w *HandlerWrapper) WithReturn(rootCtx context.Context, n int, span string) (count int64, err error) {
	rootCtx, span1 := w.tracer.Start(rootCtx, w.prefix + "WithReturn")
	defer span1.End()

	count, err = w.Handler.WithReturn(rootCtx, n, span)
	if err != nil {
		span1.RecordError(err)
		span1.SetStatus(codes.Error, err.Error())
	}
	return count, err
}
`, buf.String())
}

//revive:disable:line-length-limit
func TestGenerateCode_W_In_Param(t *testing.T) {
	var buf bytes.Buffer
	err := generateCode(&buf, packageTypeInfo{
		name: "example",
		imports: []importInfo{
			{
				path:     "context",
				usedName: "context",
			},
			{
				path:     "time",
				usedName: "time",
			},
			{
				path:     "sample/codes",
				usedName: "codes",
			},
			{
				path:     "sample/trace",
				usedName: "trace",
			},
		},
		interfaces: []interfaceInfo{
			{
				name: "Handler",
				methods: []methodType{
					{
						name: "Hello",
						params: []tupleType{
							{
								name:       "ctx",
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
							{
								name:    "n",
								typeStr: "int",
							},
							{
								name:    "createdAt",
								typeStr: "time.Time",
								pkgList: []tupleTypePkg{
									{
										path:  "time",
										begin: 0,
										end:   len("time"),
									},
								},
							},
							{
								name:    "value",
								typeStr: "*codes.Hello",
								pkgList: []tupleTypePkg{
									{
										path:  "sample/codes",
										begin: 1,
										end:   1 + len("codes"),
									},
								},
							},
							{
								name:    "t",
								typeStr: "*trace.Hello",
								pkgList: []tupleTypePkg{
									{
										path:  "sample/trace",
										begin: 1,
										end:   1 + len("trace"),
									},
								},
							},
						},
						results: []tupleType{
							{
								name:       "",
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
					{
						name: "UseW",
						params: []tupleType{
							{
								name:       "ctx",
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
							{
								name:    "w",
								typeStr: "int64",
							},
						},
						results: []tupleType{
							{
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
					{
						name: "ReturnW",
						params: []tupleType{
							{
								name:       "ctx",
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
						},
						results: []tupleType{
							{
								name:       "w",
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
							{
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, `
package example

import (
	"context"
	"time"
	"sample/codes"
	"sample/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	otelcodes "go.opentelemetry.io/otel/codes"
)

// HandlerWrapper wraps OpenTelemetry's span
type HandlerWrapper struct {
	Handler
	tracer oteltrace.Tracer
	prefix string
}

// NewHandlerWrapper creates a wrapper
func NewHandlerWrapper(wrapped Handler, tracer oteltrace.Tracer, prefix string) *HandlerWrapper {
	return &HandlerWrapper{
		Handler: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// Hello ...
func (w *HandlerWrapper) Hello(ctx context.Context, n int, createdAt time.Time, value *codes.Hello, t *trace.Hello) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "Hello")
	defer span.End()

	err = w.Handler.Hello(ctx, n, createdAt, value, t)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
	}
	return err
}

// UseW ...
func (w *HandlerWrapper) UseW(ctx context.Context, a int64) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "UseW")
	defer span.End()

	err = w.Handler.UseW(ctx, a)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
	}
	return err
}

// ReturnW ...
func (w *HandlerWrapper) ReturnW(ctx context.Context) (ctx1 context.Context, err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "ReturnW")
	defer span.End()

	ctx1, err = w.Handler.ReturnW(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otelcodes.Error, err.Error())
	}
	return ctx1, err
}
`, buf.String())
}

//revive:enable:line-length-limit

func TestGenerateCode_Without_Arg_Name(t *testing.T) {
	var buf bytes.Buffer
	err := generateCode(&buf, packageTypeInfo{
		name: "example",
		imports: []importInfo{
			{
				path:     "context",
				usedName: "context",
			},
		},
		interfaces: []interfaceInfo{
			{
				name: "Handler",
				methods: []methodType{
					{
						name: "WithoutName",
						params: []tupleType{
							{
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
							{
								typeStr: "int",
							},
						},
						results: []tupleType{
							{
								name:    "",
								typeStr: "string",
							},
							{
								name:       "",
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, `
package example

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// HandlerWrapper wraps OpenTelemetry's span
type HandlerWrapper struct {
	Handler
	tracer trace.Tracer
	prefix string
}

// NewHandlerWrapper creates a wrapper
func NewHandlerWrapper(wrapped Handler, tracer trace.Tracer, prefix string) *HandlerWrapper {
	return &HandlerWrapper{
		Handler: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// WithoutName ...
func (w *HandlerWrapper) WithoutName(ctx context.Context, a int) (a1 string, err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "WithoutName")
	defer span.End()

	a1, err = w.Handler.WithoutName(ctx, a)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return a1, err
}
`, buf.String())
}

func TestGenerateCode_Trace_As_Var_Name(t *testing.T) {
	var buf bytes.Buffer
	err := generateCode(&buf, packageTypeInfo{
		name: "example",
		imports: []importInfo{
			{
				path:     "context",
				usedName: "context",
			},
		},
		interfaces: []interfaceInfo{
			{
				name: "Handler",
				methods: []methodType{
					{
						name: "WithoutName",
						params: []tupleType{
							{
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
							{
								name:    "trace",
								typeStr: "int",
							},
						},
						results: []tupleType{
							{
								name:       "",
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, `
package example

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// HandlerWrapper wraps OpenTelemetry's span
type HandlerWrapper struct {
	Handler
	tracer trace.Tracer
	prefix string
}

// NewHandlerWrapper creates a wrapper
func NewHandlerWrapper(wrapped Handler, tracer trace.Tracer, prefix string) *HandlerWrapper {
	return &HandlerWrapper{
		Handler: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// WithoutName ...
func (w *HandlerWrapper) WithoutName(ctx context.Context, a int) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "WithoutName")
	defer span.End()

	err = w.Handler.WithoutName(ctx, a)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
`, buf.String())
}

func TestGenerateCode_Use_Type_In_Current_Package(t *testing.T) {
	var buf bytes.Buffer
	err := generateCode(&buf, packageTypeInfo{
		name: "example",
		path: "hello/example",
		imports: []importInfo{
			{
				path:     "context",
				usedName: "context",
			},
		},
		interfaces: []interfaceInfo{
			{
				name: "Handler",
				methods: []methodType{
					{
						name: "WithoutName",
						params: []tupleType{
							{
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
							{
								name:    "u",
								typeStr: "*User",
								pkgList: []tupleTypePkg{
									{
										path:  "hello/example",
										begin: 1,
										end:   1,
									},
								},
							},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, `
package example

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// HandlerWrapper wraps OpenTelemetry's span
type HandlerWrapper struct {
	Handler
	tracer trace.Tracer
	prefix string
}

// NewHandlerWrapper creates a wrapper
func NewHandlerWrapper(wrapped Handler, tracer trace.Tracer, prefix string) *HandlerWrapper {
	return &HandlerWrapper{
		Handler: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// WithoutName ...
func (w *HandlerWrapper) WithoutName(ctx context.Context, u *User) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "WithoutName")
	defer span.End()

	w.Handler.WithoutName(ctx, u)
}
`, buf.String())
}

func TestGenerateCode_To_Another_Package(t *testing.T) {
	var buf bytes.Buffer
	err := generateCode(&buf, packageTypeInfo{
		name: "example",
		path: "hello/example",
		imports: []importInfo{
			{
				path:     "context",
				usedName: "context",
			},
		},
		interfaces: []interfaceInfo{
			{
				name: "Handler",
				methods: []methodType{
					{
						name: "WithoutName",
						params: []tupleType{
							{
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
							{
								name:    "u",
								typeStr: "*User",
								pkgList: []tupleTypePkg{
									{
										path:  "hello/example",
										begin: 1,
										end:   1,
									},
								},
							},
						},
					},
				},
			},
		},
	}, WithInAnotherPackage("example_wrapper"))
	assert.Equal(t, nil, err)
	assert.Equal(t, `
package example_wrapper

import (
	"hello/example"
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// HandlerWrapper wraps OpenTelemetry's span
type HandlerWrapper struct {
	example.Handler
	tracer trace.Tracer
	prefix string
}

// NewHandlerWrapper creates a wrapper
func NewHandlerWrapper(wrapped example.Handler, tracer trace.Tracer, prefix string) *HandlerWrapper {
	return &HandlerWrapper{
		Handler: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// WithoutName ...
func (w *HandlerWrapper) WithoutName(ctx context.Context, u *example.User) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "WithoutName")
	defer span.End()

	w.Handler.WithoutName(ctx, u)
}
`, buf.String())
}

func TestGenerateCode_To_Another_Package_Return_Error(t *testing.T) {
	var buf bytes.Buffer
	err := generateCode(&buf, packageTypeInfo{
		name: "example",
		path: "hello/example",
		imports: []importInfo{
			{
				path:     "context",
				usedName: "context",
			},
		},
		interfaces: []interfaceInfo{
			{
				name: "Handler",
				methods: []methodType{
					{
						name: "HelloWorld",
						params: []tupleType{
							{
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
							{
								name:    "u",
								typeStr: "*User",
								pkgList: []tupleTypePkg{
									{
										path:  "hello/example",
										begin: 1,
										end:   1,
									},
								},
							},
						},
						results: []tupleType{
							{
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
					{
						name: "WithoutContext",
						params: []tupleType{
							{
								name:    "n",
								typeStr: "int",
							},
						},
						results: []tupleType{
							{
								typeStr:    "error",
								recognized: recognizedTypeError,
							},
						},
					},
				},
			},
		},
	}, WithInAnotherPackage("example_wrapper"))
	assert.Equal(t, nil, err)
	assert.Equal(t, `
package example_wrapper

import (
	"hello/example"
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// HandlerWrapper wraps OpenTelemetry's span
type HandlerWrapper struct {
	example.Handler
	tracer trace.Tracer
	prefix string
}

// NewHandlerWrapper creates a wrapper
func NewHandlerWrapper(wrapped example.Handler, tracer trace.Tracer, prefix string) *HandlerWrapper {
	return &HandlerWrapper{
		Handler: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// HelloWorld ...
func (w *HandlerWrapper) HelloWorld(ctx context.Context, u *example.User) (err error) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "HelloWorld")
	defer span.End()

	err = w.Handler.HelloWorld(ctx, u)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
`, buf.String())
}

func TestGenerateCode_Multiple_Interfaces(t *testing.T) {
	var buf bytes.Buffer
	err := generateCode(&buf, packageTypeInfo{
		name: "example",
		path: "hello/example",
		imports: []importInfo{
			{
				path:     "context",
				usedName: "context",
			},
		},
		interfaces: []interfaceInfo{
			{
				name: "Handler",
				methods: []methodType{
					{
						name: "HelloWorld",
						params: []tupleType{
							{
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
							{
								name:    "u",
								typeStr: "*User",
								pkgList: []tupleTypePkg{
									{
										path:  "hello/example",
										begin: 1,
										end:   1,
									},
								},
							},
						},
					},
				},
			},
			{
				name: "IRepo",
				methods: []methodType{
					{
						name: "GetUser",
						params: []tupleType{
							{
								typeStr:    "context.Context",
								recognized: recognizedTypeContext,
								pkgList:    pkgListContext(),
							},
							{
								name:    "id",
								typeStr: "int",
							},
						},
					},
				},
			},
		},
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, `
package example

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
)

// HandlerWrapper wraps OpenTelemetry's span
type HandlerWrapper struct {
	Handler
	tracer trace.Tracer
	prefix string
}

// NewHandlerWrapper creates a wrapper
func NewHandlerWrapper(wrapped Handler, tracer trace.Tracer, prefix string) *HandlerWrapper {
	return &HandlerWrapper{
		Handler: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// HelloWorld ...
func (w *HandlerWrapper) HelloWorld(ctx context.Context, u *User) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "HelloWorld")
	defer span.End()

	w.Handler.HelloWorld(ctx, u)
}

// IRepoWrapper wraps OpenTelemetry's span
type IRepoWrapper struct {
	IRepo
	tracer trace.Tracer
	prefix string
}

// NewIRepoWrapper creates a wrapper
func NewIRepoWrapper(wrapped IRepo, tracer trace.Tracer, prefix string) *IRepoWrapper {
	return &IRepoWrapper{
		IRepo: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}

// GetUser ...
func (w *IRepoWrapper) GetUser(ctx context.Context, id int) {
	ctx, span := w.tracer.Start(ctx, w.prefix + "GetUser")
	defer span.End()

	w.IRepo.GetUser(ctx, id)
}
`, buf.String())
}
