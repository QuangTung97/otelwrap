package generate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadPackageTypeInfo(t *testing.T) {
	info, err := loadPackageTypeData("./hello", "Processor")
	assert.Equal(t, nil, err)

	interface1 := interfaceInfo{
		name: "Processor",
		methods: []methodType{
			{
				name: "DoA",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgPath:    "context",
						pkgBegin:   0,
						pkgEnd:     len("context"),
					},
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
			{
				name: "Handle",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgPath:    "context",
						pkgEnd:     len("context"),
					},
					{
						name:     "u",
						typeStr:  "*User",
						pkgPath:  "github.com/QuangTung97/otelwrap/internal/generate/hello",
						pkgBegin: 1,
						pkgEnd:   1,
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
				name: "Get",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgPath:    "context",
						pkgEnd:     len("context"),
					},
					{
						name:    "id",
						typeStr: "int64",
					},
					{
						name:    "content",
						typeStr: "otelgosdk.Content",
						pkgPath: "github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk",
						pkgEnd:  len("otelgosdk"),
					},
				},
				results: []tupleType{
					{
						typeStr: "otelgo.Person",
						pkgPath: "github.com/QuangTung97/otelwrap/internal/generate/hello/otel",
						pkgEnd:  len("otelgo"),
					},
					{
						typeStr:    "error",
						recognized: recognizedTypeError,
					},
				},
			},
			{
				name: "NoName",
				params: []tupleType{
					{
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgPath:    "context",
						pkgEnd:     len("context"),
					},
					{
						typeStr: "int",
					},
				},
				results: nil,
			},
			{
				name: "ManyParams",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgPath:    "context",
						pkgEnd:     len("context"),
					},
					{
						name:       "params",
						typeStr:    "...string",
						isVariadic: true,
					},
				},
				results: nil,
			},
			{
				name: "UseArray",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgPath:    "context",
						pkgEnd:     len("context"),
					},
					{
						name:     "contents",
						typeStr:  "[]*otelgosdk.Content",
						pkgPath:  "github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk",
						pkgBegin: 3,
						pkgEnd:   3 + len("otelgosdk"),
					},
				},
				results: []tupleType{
					{
						name:    "",
						typeStr: "User",
						pkgPath: "github.com/QuangTung97/otelwrap/internal/generate/hello",
					},
					{
						name:       "",
						typeStr:    "error",
						recognized: recognizedTypeError,
					},
				},
			},
		},
	}

	assert.Equal(t, packageTypeInfo{
		name: "hello",
		path: "github.com/QuangTung97/otelwrap/internal/generate/hello",
		imports: []importInfo{
			{
				aliasName: "",
				path:      "context",
				usedName:  "context",
			},
			{
				aliasName: "otelgo",
				path:      "github.com/QuangTung97/otelwrap/internal/generate/hello/otel",
				usedName:  "otelgo",
			}, {
				aliasName: "otelgosdk",
				path:      "github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk",
				usedName:  "otelgosdk",
			},
		},
		interfaces: []interfaceInfo{interface1},
	}, info)
}
