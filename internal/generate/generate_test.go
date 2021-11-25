package generate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func pkgListContext() []tupleTypePkg {
	return []tupleTypePkg{
		{
			path:  "context",
			begin: 0,
			end:   len("context"),
		},
	}
}

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
						pkgList:    pkgListContext(),
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
						pkgList:    pkgListContext(),
					},
					{
						name:    "u",
						typeStr: "*User",
						pkgList: []tupleTypePkg{
							{
								path:  "github.com/QuangTung97/otelwrap/internal/generate/hello",
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
				name: "Get",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgList:    pkgListContext(),
					},
					{
						name:    "id",
						typeStr: "int64",
					},
					{
						name:    "content",
						typeStr: "otelgosdk.Content",
						pkgList: []tupleTypePkg{
							{
								path: "github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk",
								end:  len("otelgosdk"),
							},
						},
					},
				},
				results: []tupleType{
					{
						typeStr: "otelgo.Person",
						pkgList: []tupleTypePkg{
							{
								path: "github.com/QuangTung97/otelwrap/internal/generate/hello/otel",
								end:  len("otelgo"),
							},
						},
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
						pkgList:    pkgListContext(),
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
						pkgList:    pkgListContext(),
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
						pkgList:    pkgListContext(),
					},
					{
						name:    "contents",
						typeStr: "[]*otelgosdk.Content",
						pkgList: []tupleTypePkg{
							{
								path:  "github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk",
								begin: 3,
								end:   3 + len("otelgosdk"),
							},
						},
					},
				},
				results: []tupleType{
					{
						name:    "",
						typeStr: "User",
						pkgList: []tupleTypePkg{
							{
								path: "github.com/QuangTung97/otelwrap/internal/generate/hello",
							},
						},
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
