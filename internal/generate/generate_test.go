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
					},
					{
						name:    "u",
						typeStr: "*User",
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
					},
					{
						name:    "id",
						typeStr: "int64",
					},
					{
						name:    "content",
						typeStr: "otelgosdk.Content",
					},
				},
				results: []tupleType{
					{
						typeStr: "otelgo.Person",
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
					},
					{
						name:       "params",
						typeStr:    "...string",
						isVariadic: true,
					},
				},
				results: nil,
			},
		},
	}

	assert.Equal(t, packageTypeInfo{
		name: "hello",
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
