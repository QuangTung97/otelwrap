package generate

import (
	"errors"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path"
)

// FindResult ...
type FindResult struct {
	SrcPkgName  string
	DestPkgPath string
}

// ErrNotFound ...
var ErrNotFound = errors.New("generate: not found")

// FindPackage ...
func FindPackage(filePath string, pkgName string) (FindResult, error) {
	srcFile, err := os.Open(filePath)
	if err != nil {
		return FindResult{}, err
	}
	defer func() {
		_ = srcFile.Close()
	}()

	data, err := io.ReadAll(srcFile)
	if err != nil {
		return FindResult{}, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, string(data), 0)
	if err != nil {
		return FindResult{}, err
	}
	for _, importSpec := range f.Imports {
		importPath := importSpec.Path.Value
		importPath = importPath[1 : len(importPath)-1]
		usedName := path.Base(importPath)
		if importSpec.Name != nil {
			usedName = importSpec.Name.Name
		}
		if usedName == pkgName {
			return FindResult{
				SrcPkgName:  f.Name.Name,
				DestPkgPath: importPath,
			}, nil
		}
	}
	return FindResult{}, ErrNotFound
}
