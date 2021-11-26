package generate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindPackage(t *testing.T) {
	result, err := FindPackage("./hello/hello.go", "otelgo")
	assert.Equal(t, nil, err)
	assert.Equal(t, FindResult{
		SrcPkgName:  "hello",
		DestPkgPath: "github.com/QuangTung97/otelwrap/internal/generate/hello/otel",
	}, result)
}

func TestFindPackage_Not_Found(t *testing.T) {
	result, err := FindPackage("./hello/hello.go", "random")
	assert.Equal(t, ErrNotFound, err)
	assert.Equal(t, FindResult{}, result)
}
