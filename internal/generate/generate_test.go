package generate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadPackageTypeInfo(t *testing.T) {
	info, err := loadPackageTypeData("./hello", "Processor")
	assert.Equal(t, nil, err)
	assert.Equal(t, packageTypeInfo{
		interfaceName: "Processor",
		imports: []packageImportInfo{
			{
				aliasName: "",
				path:      "context",
			},
			{
				aliasName: "otelgo",
				path:      "github.com/QuangTung97/otelwrap/internal/generate/hello/otel",
			}, {
				aliasName: "otelgosdk",
				path:      "github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk",
			},
		},
		methods: []packageTypeMethod{
			{
				name: "DoA",
				params: []packageTypeTuple{
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
				results: []packageTypeTuple{
					{
						typeStr:    "error",
						recognized: recognizedTypeError,
					},
				},
			},
			{
				name: "Handle",
				params: []packageTypeTuple{
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
				results: []packageTypeTuple{
					{
						typeStr:    "error",
						recognized: recognizedTypeError,
					},
				},
			},
			{
				name: "Get",
				params: []packageTypeTuple{
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
				results: []packageTypeTuple{
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
				params: []packageTypeTuple{
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
		},
	}, info)
}
