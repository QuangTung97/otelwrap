package generate

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type recognizedType int

const (
	recognizedTypeUnknown recognizedType = iota
	recognizedTypeContext
	recognizedTypeError
)

type tupleType struct {
	name       string
	typeStr    string
	recognized recognizedType
	isVariadic bool
}

type methodType struct {
	name    string
	params  []tupleType
	results []tupleType
}

type importInfo struct {
	aliasName string
	path      string
}

type interfaceInfo struct {
	name    string
	methods []methodType
}

type packageTypeInfo struct {
	imports    []importInfo
	interfaces []interfaceInfo
}

func getRecognizedType(field *ast.Field, info *types.Info) recognizedType {
	fieldType := info.TypeOf(field.Type)
	namedType, ok := fieldType.(*types.Named)
	if ok {
		name := namedType.Obj().Name()
		pkg := namedType.Obj().Pkg()
		if name == "Context" && pkg != nil && pkg.Path() == "context" {
			return recognizedTypeContext
		}
		if name == "error" && pkg == nil {
			return recognizedTypeError
		}
	}
	return recognizedTypeUnknown
}

func fieldListToTupleList(
	fileList *ast.FieldList, fset *token.FileSet,
	fileMap map[string]string, info *types.Info,
) []tupleType {
	if fileList == nil {
		return nil
	}

	var tuples []tupleType
	for _, field := range fileList.List {
		begin := field.Type.Pos()
		end := field.Type.End()
		file := fset.File(begin)

		filename := file.Name()
		typeStr := fileMap[filename][file.Offset(begin):file.Offset(end)]

		isVariadic := false
		_, ok := field.Type.(*ast.Ellipsis)
		if ok {
			isVariadic = true
		}

		recognized := getRecognizedType(field, info)
		for _, resultName := range field.Names {
			tuples = append(tuples, tupleType{
				name:       resultName.Name,
				typeStr:    typeStr,
				recognized: recognized,
				isVariadic: isVariadic,
			})
		}
		if len(field.Names) == 0 {
			tuples = append(tuples, tupleType{
				typeStr:    typeStr,
				recognized: recognized,
				isVariadic: isVariadic,
			})
		}
	}
	return tuples
}

func readFiles(files []string) map[string]string {
	fileMap := map[string]string{}
	for _, filename := range files {
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
	return fileMap
}

func findInterfaceType(interfaceName string, syntax []*ast.File) (*ast.InterfaceType, error) {
	var foundTypeSpec *ast.TypeSpec
	for _, syntax := range syntax {
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
		return nil, fmt.Errorf("name '%s' is not an interface", interfaceName)
	}
	return interfaceType, nil
}

func getImportInfos(syntaxFiles []*ast.File, acceptedPackages map[string]struct{}) []importInfo {
	var imports []importInfo
	for _, syntax := range syntaxFiles {
		for _, importDetail := range syntax.Imports {
			pathValue, err := strconv.Unquote(importDetail.Path.Value)
			if err != nil {
				panic(err)
			}

			aliasName := ""
			usedName := path.Base(pathValue)
			if importDetail.Name != nil {
				aliasName = importDetail.Name.Name
				usedName = aliasName
			}

			if _, ok := acceptedPackages[usedName]; !ok {
				continue
			}

			imports = append(imports, importInfo{
				aliasName: aliasName,
				path:      pathValue,
			})
		}
	}
	return imports
}

type importVisitor struct {
	info         *types.Info
	packageNames map[string]struct{}
}

func newImportVisitor(info *types.Info) *importVisitor {
	return &importVisitor{
		info:         info,
		packageNames: map[string]struct{}{},
	}
}

func (v *importVisitor) Visit(node ast.Node) ast.Visitor {
	ident, ok := node.(*ast.Ident)
	if !ok {
		return v
	}
	object, ok := v.info.Uses[ident]
	if !ok {
		return v
	}
	pkgName, ok := object.(*types.PkgName)
	if !ok {
		return v
	}
	v.packageNames[pkgName.Name()] = struct{}{}
	return v
}

func getInterfaceInfo(
	interfaceName string, foundPkg *packages.Package,
	fileMap map[string]string,
	visitor *importVisitor,
) (interfaceInfo, error) {
	interfaceType, err := findInterfaceType(interfaceName, foundPkg.Syntax)
	if err != nil {
		return interfaceInfo{}, err
	}

	ast.Walk(visitor, interfaceType)

	methods := make([]methodType, 0, len(interfaceType.Methods.List))
	for _, field := range interfaceType.Methods.List {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		params := fieldListToTupleList(funcType.Params, foundPkg.Fset, fileMap, foundPkg.TypesInfo)
		results := fieldListToTupleList(funcType.Results, foundPkg.Fset, fileMap, foundPkg.TypesInfo)

		methods = append(methods, methodType{
			name:    field.Names[0].Name,
			params:  params,
			results: results,
		})
	}

	return interfaceInfo{
		name:    interfaceName,
		methods: methods,
	}, nil
}

func loadPackageTypeData(pattern string, interfaceNames ...string) (packageTypeInfo, error) {
	mode := packages.NeedName | packages.NeedSyntax | packages.NeedCompiledGoFiles |
		packages.NeedTypes | packages.NeedTypesInfo

	pkgList, err := packages.Load(&packages.Config{
		Mode: mode,
	}, pattern)
	if err != nil {
		return packageTypeInfo{}, err
	}

	var foundPkg *packages.Package
	for _, pkg := range pkgList {
		if pkg.Types.Scope().Lookup(interfaceNames[0]) != nil {
			foundPkg = pkg
			break
		}
	}

	if foundPkg == nil {
		return packageTypeInfo{}, fmt.Errorf("can not find interface '%s'", interfaceNames[0])
	}
	for _, otherName := range interfaceNames[1:] {
		if foundPkg.Types.Scope().Lookup(otherName) == nil {
			return packageTypeInfo{}, fmt.Errorf("can not find interface '%s'", otherName)
		}
	}

	fileMap := readFiles(foundPkg.CompiledGoFiles)

	visitor := newImportVisitor(foundPkg.TypesInfo)

	var interfaces []interfaceInfo
	for _, interfaceName := range interfaceNames {
		info, err := getInterfaceInfo(interfaceName, foundPkg, fileMap, visitor)
		if err != nil {
			return packageTypeInfo{}, err
		}
		interfaces = append(interfaces, info)
	}

	imports := getImportInfos(foundPkg.Syntax, visitor.packageNames)

	return packageTypeInfo{
		imports:    imports,
		interfaces: interfaces,
	}, nil
}
