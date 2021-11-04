package generate

import (
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"os"
)

type packageTypeTuple struct {
	name    string
	typeStr string
}

type packageTypeMethod struct {
	name    string
	params  []packageTypeTuple
	results []packageTypeTuple
}

type packageInfo struct {
	path string
}

type packageTypeInfo struct {
	interfaceName string
	packageMap    []packageInfo
	methods       []packageTypeMethod
}

func fieldListToTupleList(
	fileList *ast.FieldList, fset *token.FileSet, fileMap map[string]string,
) []packageTypeTuple {
	if fileList == nil {
		return nil
	}

	var tuples []packageTypeTuple
	for _, resultField := range fileList.List {
		begin := resultField.Type.Pos()
		end := resultField.Type.End()
		file := fset.File(begin)

		filename := file.Name()
		typeStr := fileMap[filename][file.Offset(begin):file.Offset(end)]

		for _, resultName := range resultField.Names {
			tuples = append(tuples, packageTypeTuple{
				name:    resultName.Name,
				typeStr: typeStr,
			})
		}
		if len(resultField.Names) == 0 {
			tuples = append(tuples, packageTypeTuple{
				typeStr: typeStr,
			})
		}
	}
	return tuples
}

func loadPackageTypeData(pattern string, interfaceName string) (packageTypeInfo, error) {
	mode := packages.NeedName | packages.NeedSyntax | packages.NeedCompiledGoFiles |
		packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports

	pkgList, err := packages.Load(&packages.Config{
		Mode: mode,
	}, pattern)
	if err != nil {
		return packageTypeInfo{}, err
	}

	var foundPkg *packages.Package
	for _, pkg := range pkgList {
		if pkg.Types.Scope().Lookup(interfaceName) != nil {
			foundPkg = pkg
			break
		}
	}

	if foundPkg == nil {
		return packageTypeInfo{}, fmt.Errorf("can not find interface '%s'", interfaceName)
	}

	fileMap := map[string]string{}
	for _, filename := range foundPkg.CompiledGoFiles {
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}

		data, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err)
		}
		fileMap[filename] = string(data)
		_ = file.Close()
	}

	var foundTypeSpec *ast.TypeSpec
	for _, syntax := range foundPkg.Syntax {
		for _, decl := range syntax.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if typeSpec.Name.Name == interfaceName {
					foundTypeSpec = typeSpec
				}
			}
		}
	}

	interfaceType, ok := foundTypeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return packageTypeInfo{}, fmt.Errorf("name '%s' is not an interface", interfaceName)
	}

	methods := make([]packageTypeMethod, 0, len(interfaceType.Methods.List))
	for _, field := range interfaceType.Methods.List {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		params := fieldListToTupleList(funcType.Params, foundPkg.Fset, fileMap)
		results := fieldListToTupleList(funcType.Results, foundPkg.Fset, fileMap)

		methods = append(methods, packageTypeMethod{
			name:    field.Names[0].Name,
			params:  params,
			results: results,
		})
	}

	return packageTypeInfo{
		interfaceName: interfaceName,
		methods:       methods,
	}, nil
}
