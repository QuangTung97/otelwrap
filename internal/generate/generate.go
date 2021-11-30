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

type tupleTypePkg struct {
	path  string
	begin int
	end   int
}

type tupleType struct {
	name       string
	typeStr    string
	recognized recognizedType
	isVariadic bool

	pkgList []tupleTypePkg
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

	pkgList []tupleTypePkg

	packageBegin int
	packageEnd   int
	foundPkg     bool
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
		v.foundPkg = true
		return v
	}

	pkg := object.Pkg()
	if pkg != nil {
		var pkgInfo tupleTypePkg
		if v.foundPkg {
			v.foundPkg = false
			pkgInfo = tupleTypePkg{
				path:  pkg.Path(),
				begin: v.packageBegin,
				end:   v.packageEnd,
			}
		} else {
			identBegin := int(ident.Pos() - v.begin)
			pkgInfo = tupleTypePkg{
				path:  pkg.Path(),
				begin: identBegin,
				end:   identBegin,
			}
		}
		v.pkgList = append(v.pkgList, pkgInfo)
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

		recognized := getRecognizedType(field, info)
		tupleTemplate := tupleType{
			typeStr:    typeStr,
			recognized: recognized,
			isVariadic: isVariadic,

			pkgList: visitor.pkgList,
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

			if _, ok := acceptedPackages[pathValue]; !ok {
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

type importVisitorData struct {
	packagePaths map[string]struct{}

	existedImports map[string]struct{}
	imports        []importInfo
}

type importVisitor struct {
	info *types.Info
	data *importVisitorData
}

func newImportVisitorData() *importVisitorData {
	return &importVisitorData{
		packagePaths:   map[string]struct{}{},
		existedImports: map[string]struct{}{},
		imports:        nil,
	}
}

func newImportVisitor(info *types.Info, visitorData *importVisitorData) *importVisitor {
	return &importVisitor{
		info: info,
		data: visitorData,
	}
}

func (v *importVisitorData) append(imports []importInfo) {
	for _, importDetail := range imports {
		if _, existed := v.existedImports[importDetail.path]; existed {
			continue
		}

		v.existedImports[importDetail.path] = struct{}{}
		v.imports = append(v.imports, importDetail)
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
	v.data.packagePaths[pkgName.Imported().Path()] = struct{}{}
	return v
}

type embeddedInterface struct {
	name    string
	pkgPath string
}

func getEmbeddedInterface(field *ast.Field, foundPkg *packages.Package) (embeddedInterface, bool) {
	selector, ok := field.Type.(*ast.SelectorExpr)
	if !ok {
		ident, ok := field.Type.(*ast.Ident)
		if !ok {
			return embeddedInterface{}, false
		}
		object, ok := foundPkg.TypesInfo.Uses[ident]
		if !ok {
			return embeddedInterface{}, false
		}
		return embeddedInterface{
			name:    ident.Name,
			pkgPath: object.Pkg().Path(),
		}, true
	}

	packageIdent, ok := selector.X.(*ast.Ident)
	if !ok {
		return embeddedInterface{}, false
	}

	object, ok := foundPkg.TypesInfo.Uses[packageIdent]
	if !ok {
		return embeddedInterface{}, false
	}

	pkg, ok := object.(*types.PkgName)
	if !ok {
		return embeddedInterface{}, false
	}

	return embeddedInterface{
		name:    selector.Sel.Name,
		pkgPath: pkg.Imported().Path(),
	}, true
}

func getInterfaceInfoRecursive(
	methods []methodType,
	loaded loadedPackages,
	interfaceName string,
	foundPkg loadedPackage,
	visitorData *importVisitorData,
) ([]methodType, error) {
	interfaceType, err := findInterfaceType(interfaceName, foundPkg.pkg.Syntax)
	if err != nil {
		return nil, err
	}

	visitor := newImportVisitor(foundPkg.pkg.TypesInfo, visitorData)

	for _, field := range interfaceType.Methods.List {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			embed, ok := getEmbeddedInterface(field, foundPkg.pkg)
			if !ok {
				continue
			}

			embeddedPkg, err := loadPackageForInterfaces(loaded, embed.pkgPath, embed.name)
			if err != nil {
				return nil, err
			}

			methods, err = getInterfaceInfoRecursive(methods, loaded, embed.name, embeddedPkg, visitorData)
			if err != nil {
				return nil, err
			}

			visitorData.append(getImportInfos(embeddedPkg.pkg.Syntax, visitorData.packagePaths))
			continue
		}

		ast.Walk(visitor, field)

		params := fieldListToTupleList(funcType.Params, foundPkg.pkg.Fset, foundPkg.fileMap, foundPkg.pkg.TypesInfo)
		results := fieldListToTupleList(funcType.Results, foundPkg.pkg.Fset, foundPkg.fileMap, foundPkg.pkg.TypesInfo)

		methods = append(methods, methodType{
			name:    field.Names[0].Name,
			params:  params,
			results: results,
		})
	}

	return methods, nil
}

func getInterfaceInfo(
	loaded loadedPackages,
	interfaceName string,
	foundPkg loadedPackage,
	visitorData *importVisitorData,
) (interfaceInfo, error) {
	methods := make([]methodType, 0)

	methods, err := getInterfaceInfoRecursive(methods, loaded, interfaceName, foundPkg, visitorData)
	if err != nil {
		return interfaceInfo{}, err
	}

	return interfaceInfo{
		name:    interfaceName,
		methods: methods,
	}, nil
}

type loadedPackage struct {
	fileMap map[string]string
	pkg     *packages.Package
}

type loadedPackages = map[string]loadedPackage

const loadPackageMode = packages.NeedName | packages.NeedSyntax | packages.NeedCompiledGoFiles |
	packages.NeedTypes | packages.NeedTypesInfo

func loadPackageForInterfaces(
	loaded loadedPackages, pattern string, interfaceNames ...string,
) (loadedPackage, error) {
	if pkg, existed := loaded[pattern]; existed {
		_, err := checkAndFindPackageForInterfaces([]*packages.Package{pkg.pkg}, interfaceNames...)
		if err != nil {
			return loadedPackage{}, err
		}
		return pkg, nil
	}

	pkgList, err := packages.Load(&packages.Config{
		Mode: loadPackageMode,
	}, pattern)
	if err != nil {
		return loadedPackage{}, err
	}

	foundPkg, err := checkAndFindPackageForInterfaces(pkgList, interfaceNames...)
	if err != nil {
		return loadedPackage{}, err
	}

	result := loadedPackage{
		pkg:     foundPkg,
		fileMap: readFiles(foundPkg.CompiledGoFiles),
	}
	loaded[foundPkg.PkgPath] = result
	return result, nil
}

func checkAndFindPackageForInterfaces(
	pkgList []*packages.Package, interfaceNames ...string,
) (*packages.Package, error) {
	var foundPkg *packages.Package
	for _, pkg := range pkgList {
		if pkg.Types.Scope().Lookup(interfaceNames[0]) != nil {
			foundPkg = pkg
			break
		}
	}
	if foundPkg == nil {
		return nil, fmt.Errorf("can not find interface '%s'", interfaceNames[0])
	}

	for _, interfaceName := range interfaceNames[1:] {
		if foundPkg.Types.Scope().Lookup(interfaceName) == nil {
			return nil, fmt.Errorf("can not find interface '%s'", interfaceName)
		}
	}
	return foundPkg, nil
}

func loadPackageTypeData(pattern string, interfaceNames ...string) (packageTypeInfo, error) {
	loaded := loadedPackages{}
	foundPkg, err := loadPackageForInterfaces(loaded, pattern, interfaceNames...)
	if err != nil {
		return packageTypeInfo{}, err
	}

	visitorData := newImportVisitorData()

	var interfaces []interfaceInfo
	for _, interfaceName := range interfaceNames {
		info, err := getInterfaceInfo(loaded, interfaceName, foundPkg, visitorData)
		if err != nil {
			return packageTypeInfo{}, err
		}
		interfaces = append(interfaces, info)
	}

	visitorData.append(getImportInfos(foundPkg.pkg.Syntax, visitorData.packagePaths))

	return packageTypeInfo{
		name:       foundPkg.pkg.Name,
		path:       foundPkg.pkg.PkgPath,
		imports:    visitorData.imports,
		interfaces: interfaces,
	}, nil
}
