package otelwrap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/QuangTung97/otelwrap/internal/generate"
	"github.com/QuangTung97/otelwrap/internal/generate/hello"
	"go/format"
	"io"
	"os"
	"path"
	"strings"
)

// =========================================
// For Testing Only
// =========================================

// Sample for testing
type Sample interface {
	Get(ctx context.Context) (int, error)
	Check() (bool, error)
}

// Repo for testing
type Repo interface {
	Update(ctx context.Context, id int) error
}

// HandlerAlias ...
type HandlerAlias = hello.Handler

// =========================================

// CommandArgs ...
type CommandArgs struct {
	Dir            string
	SrcFileName    string
	InterfaceNames []string
	InAnother      bool
	PkgName        string
}

func splitPackageNameFromInterfaceNames(interfaceNames []string) (string, []string, error) {
	values := strings.Split(interfaceNames[0], ".")
	if len(values) == 1 {
		for _, interfaceName := range interfaceNames[1:] {
			values = strings.Split(interfaceName, ".")
			if len(values) > 1 {
				return "", nil, errors.New("can not have mixed interface names")
			}
		}
		return "", interfaceNames, nil
	}

	packageName := values[0]
	result := make([]string, 0, len(interfaceNames))
	for _, interfaceName := range interfaceNames {
		values = strings.Split(interfaceName, ".")
		if len(values) != 2 || values[0] != packageName {
			return "", nil, errors.New("can not have mixed interface names")
		}
		result = append(result, values[1])
	}
	return packageName, result, nil
}

func findAndGenerate(w io.Writer, args CommandArgs) error {
	packageName, interfaceNames, err := splitPackageNameFromInterfaceNames(args.InterfaceNames)
	if err != nil {
		fmt.Println("splitPackageNameFromInterfaceNames", err)
		return err
	}

	if len(packageName) == 0 {
		if args.InAnother {
			return generate.LoadAndGenerate(w,
				".", interfaceNames,
				generate.WithInAnotherPackage(args.PkgName),
			)
		}
		return generate.LoadAndGenerate(w,
			".", interfaceNames,
		)
	}

	filePath := path.Join(args.Dir, args.SrcFileName)
	findResult, err := generate.FindPackage(filePath, packageName)
	if err != nil {
		fmt.Println("FindPackage", err)
		return err
	}

	return generate.LoadAndGenerate(w,
		findResult.DestPkgPath, interfaceNames,
		generate.WithInAnotherPackage(findResult.SrcPkgName),
	)
}

// RunCommand ...
func RunCommand(args CommandArgs, outFile string) error {
	var buf bytes.Buffer
	_, _ = buf.WriteString(`// Code generated by otelwrap; DO NOT EDIT.
// github.com/QuangTung97/otelwrap

`)
	err := findAndGenerate(&buf, args)
	if err != nil {
		return err
	}

	data, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println("format.Source", string(buf.Bytes()), err)
		return err
	}

	file, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = file.Write(data)
	return err
}

// CheckInAnother ...
func CheckInAnother(filename string) bool {
	dir := path.Dir(filename)
	return dir != "."
}
