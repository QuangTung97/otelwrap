package otelwrap

import (
	"bytes"
	"github.com/QuangTung97/otelwrap/internal/generate/hello"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ hello.Processor

func TestFindAndGenerate(t *testing.T) {
	var buf bytes.Buffer
	err := findAndGenerate(&buf, CommandArgs{
		Dir:      ".",
		Filename: "command_test.go",
		Name:     "hello.Simple",
	})
	assert.Equal(t, nil, err)
	expected := `
package otelwrap

`
	assert.Equal(t, expected, buf.String())
}
