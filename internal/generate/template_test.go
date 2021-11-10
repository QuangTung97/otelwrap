package generate

import (
	"bytes"
	"fmt"
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
				path: "context",
			},
			{
				path: "time",
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
							},
							{
								name:    "n",
								typeStr: "int",
							},
							{
								name:    "createdAt",
								typeStr: "time.Time",
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
	fmt.Println("ERR:", err)
	assert.Equal(t, `
package example

import (
	"context"
	"time"
	"go.opentelemetry.io/otel/trace"
)

// HandlerWrapper wraps OpenTelemetry's span
type HandlerWrapper struct {
	Handler
	tracer trace.Tracer
	prefix string
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
