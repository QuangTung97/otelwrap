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

	// only for generating
	recognizedTypeSpan
)

type tupleType struct {
	name       string
	typeStr    string
	recognized recognizedType
	isVariadic bool

	pkgPath  string
	pkgBegin int
	pkgEnd   int
}

type methodType struct {
	name    string
	params  []tupleType
	results []tupleType
}

type importInfo struct {
	aliasName string
	usedName  string
	path      string
}

type interfaceInfo struct {
	name    string
	methods []methodType
}

type packageTypeInfo struct {
	name string
	path string

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

type tupleVisitor struct {
	begin token.Pos
	info  *types.Info

	packagePath  string
	packageBegin int
	packageEnd   int

	identBegin int
}

func (v *tupleVisitor) Visit(node ast.Node) ast.Visitor {
	ident, ok := node.(*ast.Ident)
	if !ok {
		return v
	}
	object, ok := v.info.Uses[ident]
	if !ok {
		return v
	}
	_, ok = object.(*types.PkgName)
	if ok {
		v.packageBegin = int(ident.Pos() - v.begin)
		v.packageEnd = int(ident.End() - v.begin)
		return v
	}

	pkg := object.Pkg()
	if pkg != nil {
		v.identBegin = int(ident.Pos() - v.begin)
		v.packagePath = pkg.Path()
	}
	return v
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

		visitor := &tupleVisitor{begin: field.Type.Pos(), info: info}
		ast.Walk(visitor, field.Type)
		if visitor.packagePath != "" && visitor.packageEnd == 0 {
			visitor.packageBegin = visitor.identBegin
			visitor.packageEnd = visitor.identBegin
		}

		recognized := getRecognizedType(field, info)
		tupleTemplate := tupleType{
			typeStr:    typeStr,
			recognized: recognized,
			isVariadic: isVariadic,

			pkgPath:  visitor.packagePath,
			pkgBegin: visitor.packageBegin,
			pkgEnd:   visitor.packageEnd,
		}

		for _, resultName := range field.Names {
			tuple := tupleTemplate
			tuple.name = resultName.Name
			tuples = append(tuples, tuple)
		}
		if len(field.Names) == 0 {
			tuples = append(tuples, tupleTemplate)
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

func findInterfaceTypeForDecl(interfaceName string, syntax *ast.File) *ast.TypeSpec {
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
				return typeSpec
			}
		}
	}
	return nil
}

func findInterfaceType(interfaceName string, syntaxFiles []*ast.File) (*ast.InterfaceType, error) {
	var foundTypeSpec *ast.TypeSpec
	for _, syntax := range syntaxFiles {
		typeSpec := findInterfaceTypeForDecl(interfaceName, syntax)
		if typeSpec != nil {
			foundTypeSpec = typeSpec
			break
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
				usedName:  usedName,
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
		name:       foundPkg.Name,
		path:       foundPkg.PkgPath,
		imports:    imports,
		interfaces: interfaces,
	}, nil
}
