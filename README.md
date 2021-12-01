## OtelWrap: code generation tool for Go OpenTelemetry

### What is OtelWrap?

**OtelWrap** is a tool that generates a decorator implementation of any interfaces that can be used for instrumentation
with Go OpenTelemetry library. Inspired by https://github.com/matryer/moq

Supporting:

* Any interface and any method with **context.Context** as the first parameter.
* Detecting **error** return and the set span's error status accordingly.
* Only tested using **go generate**.
* Interface embedding.
* Generating inside / outside of current package.

### Installing

Using conventional **tools.go** file for pinning version in **go.mod** / **go.sum**.

```go
// +build tools

package tools

import (
    _ "github.com/QuangTung97/otelwrap"
)
```

And then download and install the binary with commands:

```shell
$ go mod tidy
$ go install github.com/QuangTung97/otelwrap
```

### Usage

```
otelwrap [flags] -source-dir interface [interface2 interface3 ...]
    --out string (required)
        output file
```

Using **go generate**:

```go
package example

import "context"

//go:generate otelwrap -out interface_wrappers.go . MyInterface

type MyInterface interface {
    Method1(ctx context.Context) error
    Method2(ctx context.Context, x int)
    Method3()
}
```

The run ``go generate ./...`` in your module.

To use the generated struct, simply wraps the original implementation. The generated code  is very easy to read.

```go
package example

import "go.opentelemetry.io/otel"

func InitMyInterface() MyInterface {
    original := NewMyInterfaceImpl()
    return NewMyInterfaceWrapper(original, otel.GetTracerProvider().Tracer("example"), "prefix")
}


```