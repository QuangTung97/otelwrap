package generate

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"reflect"
)

type packageTypeTuple struct {
	name        string
	typeName    string
	packageName string
}

type packageTypeMethod struct {
	params  []packageTypeTuple
	results []packageTypeTuple
}

type packageInfo struct {
	path string
}

type packageTypeInfo struct {
	interfaceName string
	packageMap    map[string]packageInfo
	methods       map[string]packageTypeMethod
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
	for _, field := range interfaceType.Methods.List {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		for _, paramField := range funcType.Params.List {
			fmt.Println(paramField.Names)
			fmt.Println(paramField.Type)
			fmt.Println(paramField.Type.Pos(), paramField.Type.End())
			fmt.Println(reflect.TypeOf(paramField.Type))

			selectorExpr, ok := paramField.Type.(*ast.SelectorExpr)
			if ok {
				fmt.Println("SEL:", selectorExpr.Sel)
				fmt.Println("Expr:", selectorExpr.X, reflect.TypeOf(selectorExpr.X))
			}
		}
	}

	return packageTypeInfo{}, nil
}
