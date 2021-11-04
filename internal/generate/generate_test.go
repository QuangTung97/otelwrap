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
