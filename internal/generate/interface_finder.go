package generate

import (
	"fmt"
	"go/ast"
)

type interfaceInfoFinder struct {
	methods     []methodType
	loaded      loadedPackages
	visitorData *importVisitorData
}

func newInterfaceInfoFinder(loaded loadedPackages, visitorData *importVisitorData) *interfaceInfoFinder {
	return &interfaceInfoFinder{
		loaded:      loaded,
		visitorData: visitorData,
	}
}

func (f *interfaceInfoFinder) getInterfaceHandleTypeAlias(
	typeSpec *ast.TypeSpec, interfaceName string, foundPkg loadedPackage,
) error {
	embed, ok := getEmbeddedInterfaceForTypeExpr(typeSpec.Type, foundPkg.pkg)
	if !ok {
		return fmt.Errorf("name '%s' is not an interface", interfaceName)
	}

	f.visitorData.resetRootPackage(embed.pkgPath)

	embeddedPkg, err := f.loaded.loadPackageForInterfaces(embed.pkgPath, embed.name)
	if err != nil {
		return err
	}

	return f.getInterfaceInfoRecursive(embed.name, embeddedPkg)
}

func (f *interfaceInfoFinder) getInterfaceInfoRecursive(
	interfaceName string,
	foundPkg loadedPackage,
) error {
	typeSpec := findInterfaceTypeSpec(interfaceName, foundPkg.pkg.Syntax)
	if typeSpec == nil {
		return fmt.Errorf("name '%s' is not a type spec", interfaceName)
	}

	interfaceType := findInterfaceAST(typeSpec)
	if interfaceType == nil {
		return f.getInterfaceHandleTypeAlias(typeSpec, interfaceName, foundPkg)
	}

	visitor := newImportVisitor(foundPkg.pkg.TypesInfo, f.visitorData)

	for _, field := range interfaceType.Methods.List {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			embed, ok := getEmbeddedInterfaceForTypeExpr(field.Type, foundPkg.pkg)
			if !ok {
				continue
			}

			embeddedPkg, err := f.loaded.loadPackageForInterfaces(embed.pkgPath, embed.name)
			if err != nil {
				return err
			}

			err = f.getInterfaceInfoRecursive(embed.name, embeddedPkg)
			if err != nil {
				return err
			}

			continue
		}

		ast.Walk(visitor, field)

		params := fieldListToTupleList(funcType.Params, foundPkg.pkg.Fset, foundPkg.fileMap, foundPkg.pkg.TypesInfo)
		results := fieldListToTupleList(funcType.Results, foundPkg.pkg.Fset, foundPkg.fileMap, foundPkg.pkg.TypesInfo)

		f.methods = append(f.methods, methodType{
			name:    field.Names[0].Name,
			params:  params,
			results: results,
		})
	}

	return nil
}

func (f *interfaceInfoFinder) getInterfaceInfo(
	interfaceName string,
	foundPkg loadedPackage,
) (interfaceInfo, error) {
	err := f.getInterfaceInfoRecursive(interfaceName, foundPkg)
	if err != nil {
		return interfaceInfo{}, err
	}

	return interfaceInfo{
		name:    interfaceName,
		methods: f.methods,
	}, nil
}
