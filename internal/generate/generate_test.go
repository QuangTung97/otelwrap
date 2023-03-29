package generate

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

const rootPackagePath = "github.com/QuangTung97/otelwrap/internal/generate"

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
				name: "StartTimer",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgList:    pkgListContext(),
					},
					{
						name:    "d",
						typeStr: "int32",
					},
				},
			},
			{
				name: "Scan",
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
				name: "Convert",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgList:    pkgListContext(),
					},
					{
						name:    "d",
						typeStr: "time.Duration",
						pkgList: []tupleTypePkg{
							{
								path:  "time",
								begin: 0,
								end:   len("time"),
							},
						},
					},
				},
			},
			{
				name: "SetInfo",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgList:    pkgListContext(),
					},
					{
						name:    "info",
						typeStr: "ScannerInfo",
						pkgList: []tupleTypePkg{
							{
								path:  "github.com/QuangTung97/otelwrap/internal/generate/hello/embed",
								begin: 0,
								end:   0,
							},
						},
					},
				},
			},
			{
				name: "Compute",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgList:    pkgListContext(),
					},
					{
						name:    "x",
						typeStr: "string",
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
								begin: len("[]*"),
								end:   len("[]*") + len("otelgosdk"),
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
			{
				name: "UseMap",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgList:    pkgListContext(),
					},
					{
						name:    "m",
						typeStr: "map[otelgosdk.Content]otelgosdk.Content",
						pkgList: []tupleTypePkg{
							{
								path:  "github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk",
								begin: len("map["),
								end:   len("map[") + len("otelgosdk"),
							},
							{
								path:  "github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk",
								begin: len("map[otelgosdk.Content]"),
								end:   len("map[otelgosdk.Content]") + len("otelgosdk"),
							},
						},
					},
				},
				results: []tupleType{
					{
						typeStr: "map[User]User",
						pkgList: []tupleTypePkg{
							{
								path:  "github.com/QuangTung97/otelwrap/internal/generate/hello",
								begin: len("map["),
								end:   len("map["),
							},
							{
								path:  "github.com/QuangTung97/otelwrap/internal/generate/hello",
								begin: len("map[User]"),
								end:   len("map[User]"),
							},
						},
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
				path: "context",
				name: "context",
			},
			{
				path: "time",
				name: "time",
			},
			{
				path: "github.com/QuangTung97/otelwrap/internal/generate/hello/embed",
				name: "embed",
			},
			{
				path: "github.com/QuangTung97/otelwrap/internal/generate/hello/otel",
				name: "otelgo",
			},
			{
				path: "github.com/QuangTung97/otelwrap/internal/generate/hello/otel/sdk",
				name: "otelgo",
			},
		},
		interfaces: []interfaceInfo{interface1},
	}, info)
}

func TestLoadPackageTypeInfo_For_Type_Alias(t *testing.T) {
	info, err := loadPackageTypeData("./hello", "SimpleAlias")
	assert.Equal(t, nil, err)

	interface1 := interfaceInfo{
		name: "SimpleAlias",
		methods: []methodType{
			{
				name: "Scan",
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
						name:       "",
						typeStr:    "error",
						recognized: recognizedTypeError,
					},
				},
			},
			{
				name: "Convert",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgList:    pkgListContext(),
					},
					{
						name:    "d",
						typeStr: "time.Duration",
						pkgList: []tupleTypePkg{
							{
								path:  "time",
								begin: 0,
								end:   len("time"),
							},
						},
					},
				},
				results: nil,
			},
			{
				name: "SetInfo",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgList:    pkgListContext(),
					},
					{
						name:    "info",
						typeStr: "ScannerInfo",
						pkgList: []tupleTypePkg{
							{
								path: "github.com/QuangTung97/otelwrap/internal/generate/hello/embed",
							},
						},
					},
				},
				results: nil,
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
						name:       "",
						typeStr:    "error",
						recognized: recognizedTypeError,
					},
				},
			},
			{
				name: "Variadic",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgList:    pkgListContext(),
					},
					{
						name:       "names",
						typeStr:    "...string",
						isVariadic: true,
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
				path: "context",
				name: "context",
			},
			{
				path: "time",
				name: "time",
			},
			{
				path: "github.com/QuangTung97/otelwrap/internal/generate/hello/embed",
				name: "embed",
			},
		},
		interfaces: []interfaceInfo{interface1},
	}, info)
}

func TestLoadPackageTypeInfo_Not_Found(t *testing.T) {
	info, err := loadPackageTypeData("./hello", "RandomInterface")
	assert.Equal(t, errors.New("can not find interface 'RandomInterface'"), err)
	assert.Equal(t, packageTypeInfo{}, info)
}

func TestLoadPackageTypeInfo_Not_Found_Second_Interface(t *testing.T) {
	info, err := loadPackageTypeData("./hello", "Simple", "AnotherInterface")
	assert.Equal(t, errors.New("can not find interface 'AnotherInterface'"), err)
	assert.Equal(t, packageTypeInfo{}, info)
}

func TestLoadPackageTypeInfo_Not_An_Interface(t *testing.T) {
	info, err := loadPackageTypeData("./hello", "User")
	assert.Equal(t, errors.New("name 'User' is not an interface"), err)
	assert.Equal(t, packageTypeInfo{}, info)
}

func TestLoadPackageTypeInfo_Interface_With_Underscore(t *testing.T) {
	info, err := loadPackageTypeData("./hello", "InterfaceWithUnderscore")
	assert.Equal(t, nil, err)

	methods := []methodType{
		{
			name: "GetName",
			params: []tupleType{
				{
					name:       "ctx",
					typeStr:    "context.Context",
					recognized: recognizedTypeContext,
					pkgList:    pkgListContext(),
				},
				{
					name:    "_",
					typeStr: "string",
				},
				{
					name:    "_",
					typeStr: "int",
				},
			},
			results: []tupleType{
				{
					name:    "_",
					typeStr: "int32",
				},
				{
					name:       "_",
					typeStr:    "error",
					recognized: recognizedTypeError,
				},
			},
		},
	}

	assert.Equal(t, packageTypeInfo{
		name: "hello",
		path: rootPackagePath + "/hello",
		imports: []importInfo{
			{
				name: "context",
				path: "context",
			},
		},
		interfaces: []interfaceInfo{
			{
				name:    "InterfaceWithUnderscore",
				methods: methods,
			},
		},
	}, info)
}

func TestLoadPackageTypeInfo_For_Type_Alias__In_Another_Package(t *testing.T) {
	info, err := loadPackageTypeData("./hello/another", "HandlerAlias")
	assert.Equal(t, nil, err)

	interface1 := interfaceInfo{
		name: "HandlerAlias",
		methods: []methodType{
			{
				name: "Process",
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
						name:       "",
						typeStr:    "error",
						recognized: recognizedTypeError,
					},
				},
			},
		},
	}

	assert.Equal(t, packageTypeInfo{
		name: "another",
		path: "github.com/QuangTung97/otelwrap/internal/generate/hello/another",
		imports: []importInfo{
			{
				name: "context",
				path: "context",
			},
		},
		interfaces: []interfaceInfo{interface1},
	}, info)
}

func TestLoadPackageTypeInfo_For_InterfaceType_With_Generic_Params_And_Returns(t *testing.T) {
	info, err := loadPackageTypeData("./hello", "GenericHandler")
	assert.Equal(t, nil, err)

	interface1 := interfaceInfo{
		name: "GenericHandler",
		methods: []methodType{
			{
				name: "GetNull",
				params: []tupleType{
					{
						name:       "ctx",
						typeStr:    "context.Context",
						recognized: recognizedTypeContext,
						pkgList:    pkgListContext(),
					},
					{
						name:    "info",
						typeStr: "Null[otelgo.AnotherInfo]",
						pkgList: []tupleTypePkg{
							{
								path:  "github.com/QuangTung97/otelwrap/internal/generate/hello",
								begin: 0,
								end:   0,
							},
							{
								path:  "github.com/QuangTung97/otelwrap/internal/generate/hello/otel",
								begin: len("Null["),
								end:   len("Null[otelgo"),
							},
						},
					},
				},
				results: []tupleType{
					{
						name:    "",
						typeStr: "Null[otelgo.Person]",
						pkgList: []tupleTypePkg{
							{
								path:  "github.com/QuangTung97/otelwrap/internal/generate/hello",
								begin: 0,
								end:   0,
							},
							{
								path:  "github.com/QuangTung97/otelwrap/internal/generate/hello/otel",
								begin: len("Null["),
								end:   len("Null[otelgo"),
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
				name: "context",
				path: "context",
			},
			{
				name: "otelgo",
				path: "github.com/QuangTung97/otelwrap/internal/generate/hello/otel",
			},
		},
		interfaces: []interfaceInfo{interface1},
	}, info)
}
