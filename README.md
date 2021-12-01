## OtelWrap: code generation tool for Go OpenTelemetry
[![Build Status](https://app.travis-ci.com/QuangTung97/otelwrap.svg?branch=master)](https://app.travis-ci.com/QuangTung97/otelwrap)
[![Coverage Status](https://coveralls.io/repos/github/QuangTung97/otelwrap/badge.svg)](https://coveralls.io/github/QuangTung97/otelwrap)

### What is OtelWrap?

**OtelWrap** is a tool that generates a decorator implementation of any interfaces that can be used for instrumentation
with Go OpenTelemetry library. Inspired by https://github.com/matryer/moq

Supporting:

* Any interface and any method with **context.Context** as the first parameter.
* Detecting **error** return and set the span's error status accordingly.
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

To use the generated struct, simply wraps the original implementation. The generated code is very easy to read.

```go
package example

import "go.opentelemetry.io/otel"

func InitMyInterface() MyInterface {
    original := NewMyInterfaceImpl()
    return NewMyInterfaceWrapper(original, otel.GetTracerProvider().Tracer("example"), "prefix")
}


```

Can also generate for interfaces in other packages:

```go
package example

import "path/to/another"

var _ another.Interface1 // not necessary, only for keeping the import statement

//go:generate -out interface_wrappers.go . another.Interface1 another.Interface2
```

Or generate to another package:

```go
package example

import "context"

//go:generate otelwrap -out ../another/interface_wrappers.go . MyInterface

type MyInterface interface {
    Method1(ctx context.Context) error
    Method2(ctx context.Context, x int)
    Method3()
}
```
