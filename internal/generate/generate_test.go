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
				path:      `"context"`,
			},
			{
				aliasName: "otelgo",
				path:      `"github.com/QuangTung97/otelwrap/internal/generate/hello/otel"`,
			}, {
				aliasName: "otelgosdk",
				path:      `"github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk"`,
			},
			{
				aliasName: "",
				path:      `"time"`,
			},
		},
		methods: []packageTypeMethod{
			{
				name: "DoA",
				params: []packageTypeTuple{
					{
						name:    "ctx",
						typeStr: "context.Context",
					},
					{
						name:    "n",
						typeStr: "int",
					},
				},
				results: []packageTypeTuple{
					{
						typeStr: "error",
					},
				},
			},
			{
				name: "Handle",
				params: []packageTypeTuple{
					{
						name:    "ctx",
						typeStr: "context.Context",
					},
					{
						name:    "u",
						typeStr: "*User",
					},
				},
				results: []packageTypeTuple{
					{
						typeStr: "error",
					},
				},
			},
			{
				name: "Get",
				params: []packageTypeTuple{
					{
						name:    "ctx",
						typeStr: "context.Context",
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
						typeStr: "error",
					},
				},
			},
			{
				name: "NoName",
				params: []packageTypeTuple{
					{
						typeStr: "context.Context",
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
