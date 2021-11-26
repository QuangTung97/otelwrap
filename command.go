package otelwrap

import (
	"context"
	"github.com/QuangTung97/otelwrap/internal/generate"
	"io"
	"path"
	"strings"
)

// Sample for testing
type Sample interface {
	Get(ctx context.Context) (int, error)
	Check() (bool, error)
}

// CommandArgs ...
type CommandArgs struct {
	Dir      string
	Filename string
	Name     string
}

func findAndGenerate(w io.Writer, args CommandArgs) error {
	filePath := path.Join(args.Dir, args.Filename)

	values := strings.Split(args.Name, ".")

	if len(values) == 1 {
		interfaceName := values[0]
		return generate.LoadAndGenerate(w,
			".", interfaceName,
		)
	}

	pkgName := values[0]
	interfaceName := values[1]

	findResult, err := generate.FindPackage(filePath, pkgName)
	if err != nil {
		return err
	}

	return generate.LoadAndGenerate(w,
		findResult.DestPkgPath, interfaceName,
		generate.WithInAnotherPackage(findResult.SrcPkgName),
	)
}

// RunCommand ...
func RunCommand(_ CommandArgs) error {
	return nil
}
