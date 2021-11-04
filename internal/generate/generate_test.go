package generate

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go/types"
	"golang.org/x/tools/go/packages"
	"testing"
)

func TestLoadPackageTypeInfo(t *testing.T) {
	info, err := loadPackageTypeData("./hello", "Processor")
	assert.Equal(t, nil, err)
	assert.Equal(t, packageTypeInfo{
		interfaceName: "Processor",
		packageMap: map[string]packageInfo{
			"context": {
				path: "context",
			},
			"time": {
				path: "time",
			},
			"otelgo": {
				path: "github.com/QuangTung97/otelwrap/internal/generate/hello/otel",
			},
		},
		methods: map[string]packageTypeMethod{
			"DoA": {
				params: []packageTypeTuple{
					{
						name:        "ctx",
						typeName:    "Context",
						packageName: "context",
					},
					{
						name:     "n",
						typeName: "int",
					},
				},
				results: []packageTypeTuple{
					{
						typeName: "error",
					},
				},
			},
			"Handle": {
				params: []packageTypeTuple{
					{
						name:        "ctx",
						typeName:    "Context",
						packageName: "context",
					},
					{
						name:     "u",
						typeName: "User",
					},
				},
				results: []packageTypeTuple{
					{
						typeName: "error",
					},
				},
			},
			"Get": {
				params: []packageTypeTuple{
					{
						name:        "ctx",
						typeName:    "Context",
						packageName: "context",
					},
					{
						name:     "id",
						typeName: "int64",
					},
				},
				results: []packageTypeTuple{
					{
						typeName:    "Person",
						packageName: "otelgo",
					},
					{
						typeName: "error",
					},
				},
			},
		},
	}, info)
}

func TestGenerate(t *testing.T) {
	mode := packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports
	pkgList, err := packages.Load(&packages.Config{
		Mode: mode,
	}, "./hello")
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgList {
		fmt.Println(pkg.PkgPath)
		interf := pkg.Types.Scope().Lookup("Processor")
		fmt.Println(interf)
		fmt.Println(interf.Name())
		fmt.Println(interf.Id())
		undertype := interf.Type().Underlying().(*types.Interface)
		method := undertype.Method(0)
		methodType := method.Type().(*types.Signature)
		fmt.Println("PARAMS", methodType.Params())
		fmt.Println("RETURN", methodType.Results())
	}
}
