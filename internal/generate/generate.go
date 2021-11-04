package generate

import (
	"fmt"
	"go/types"
	"golang.org/x/tools/go/packages"
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
	mode := packages.NeedName | packages.NeedSyntax |
		packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports

	pkgList, err := packages.Load(&packages.Config{
		Mode: mode,
	}, pattern)
	if err != nil {
		return packageTypeInfo{}, err
	}

	var interfaceObject types.Object
	var foundPkg *types.Package
	for _, pkg := range pkgList {
		interfaceObject = pkg.Types.Scope().Lookup(interfaceName)
		if interfaceObject != nil {
			foundPkg = pkg.Types
			break
		}
	}

	if interfaceObject == nil {
		return packageTypeInfo{}, fmt.Errorf("can not find interface '%s'", interfaceName)
	}

	packageMap := map[string]packageInfo{}
	for _, importInfo := range foundPkg.Imports() {
		packageMap[importInfo.Name()] = packageInfo{
			path: importInfo.Path(),
		}
	}

	underline, ok := interfaceObject.Type().Underlying().(*types.Interface)
	if !ok {
		return packageTypeInfo{}, fmt.Errorf("name '%s' is not an interface", interfaceName)
	}

	packagePath := foundPkg.Path()
	fmt.Println(packagePath)

	numMethod := underline.NumMethods()
	methods := map[string]packageTypeMethod{}
	for i := 0; i < numMethod; i++ {
		m := underline.Method(i)

		sig := m.Type().(*types.Signature)

		params := make([]packageTypeTuple, 0, sig.Params().Len())
		for i := 0; i < sig.Params().Len(); i++ {
			param := sig.Params().At(i)

			typeName := param.Type().String()
			packageName := ""
			namedType, ok := param.Type().(*types.Named)
			if ok {
				typeName = namedType.Obj().Name()
				if namedType.Obj().Pkg().Path() != packagePath {
					packageName = namedType.Obj().Pkg().Name()
				}
			}

			params = append(params, packageTypeTuple{
				name:        param.Name(),
				typeName:    typeName,
				packageName: packageName,
			})
		}

		results := make([]packageTypeTuple, 0, sig.Results().Len())
		for i := 0; i < sig.Results().Len(); i++ {
			r := sig.Results().At(i)

			typeName := r.Type().String()
			packageName := ""
			namedType, ok := r.Type().(*types.Named)
			if ok {
				typeName = namedType.Obj().Name()

				pkg := namedType.Obj().Pkg()
				if pkg != nil && pkg.Path() != packagePath {
					packageName = namedType.Obj().Pkg().Name()
				}
			}

			results = append(results, packageTypeTuple{
				name:        r.Name(),
				typeName:    typeName,
				packageName: packageName,
			})
		}

		methods[m.Name()] = packageTypeMethod{
			params:  params,
			results: results,
		}
	}
	return packageTypeInfo{
		interfaceName: interfaceName,
		packageMap:    packageMap,
		methods:       methods,
	}, nil
}
