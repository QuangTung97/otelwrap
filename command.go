package otelwrap

import (
	"github.com/QuangTung97/otelwrap/internal/generate"
	"io"
	"path"
	"strings"
)

// CommandArgs ...
type CommandArgs struct {
	Dir      string
	Filename string
	Name     string
}

func findAndGenerate(w io.Writer, args CommandArgs) error {
	filePath := path.Join(args.Dir, args.Filename)

	values := strings.Split(args.Name, ".")

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
