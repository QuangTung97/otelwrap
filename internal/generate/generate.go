package generate

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"io"
	"os"
	"sort"
	"strings"
)

type recognizedType int

const (
	recognizedTypeUnknown recognizedType = iota
	recognizedTypeContext
	recognizedTypeError

	// only for generating
	recognizedTypeSpan
)

// tupleTypePkg for replacing package names
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
	name string
	path string
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

		data, err := io.ReadAll(file)
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

func findInterfaceTypeSpec(
	interfaceName string, syntaxFiles []*ast.File,
) *ast.TypeSpec {
	for _, syntax := range syntaxFiles {
		typeSpec := findInterfaceTypeForDecl(interfaceName, syntax)
		if typeSpec != nil {
			return typeSpec
		}
	}
	return nil
}

func findInterfaceAST(typeSpec *ast.TypeSpec) *ast.InterfaceType {
	interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		return nil
	}
	return interfaceType
}

type emptyStruct = struct{}

type importVisitorData struct {
	rootPackagePath string

	packagePaths map[string]emptyStruct

	existedImports map[string]emptyStruct
	imports        []importInfo
}

type importVisitor struct {
	info *types.Info
	data *importVisitorData
}

func newImportVisitorData(rootPackagePath string) *importVisitorData {
	return &importVisitorData{
		rootPackagePath: rootPackagePath,
		packagePaths:    map[string]emptyStruct{},
		existedImports:  map[string]emptyStruct{},
		imports:         nil,
	}
}

func newImportVisitor(info *types.Info, visitorData *importVisitorData) *importVisitor {
	return &importVisitor{
		info: info,
		data: visitorData,
	}
}

func (v *importVisitorData) resetRootPackage(pkgPath string) {
	v.rootPackagePath = pkgPath
}

func (v *importVisitorData) append(imports []importInfo) {
	for _, importDetail := range imports {
		if importDetail.path == v.rootPackagePath {
			continue
		}

		if _, existed := v.existedImports[importDetail.path]; existed {
			continue
		}

		v.existedImports[importDetail.path] = emptyStruct{}
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

	if object.Pkg() == nil {
		return v
	}

	pkgInfo := object.Pkg()
	pkgPath := pkgInfo.Path()

	v.data.packagePaths[pkgPath] = emptyStruct{}
	v.data.append([]importInfo{
		{
			name: pkgInfo.Name(),
			path: pkgPath,
		},
	})
	return v
}

type embeddedInterface struct {
	name    string
	pkgPath string
}

func getEmbeddedInterfaceForTypeExpr(typeExpr ast.Expr, foundPkg *packages.Package) (embeddedInterface, bool) {
	selector, ok := typeExpr.(*ast.SelectorExpr)
	if !ok {
		ident, ok := typeExpr.(*ast.Ident)
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

func sortImportInfos(imports []importInfo) []importInfo {
	var stdImports []importInfo
	var otherImports []importInfo
	for _, importDetail := range imports {
		if strings.ContainsRune(importDetail.path, '.') {
			otherImports = append(otherImports, importDetail)
		} else {
			stdImports = append(stdImports, importDetail)
		}
	}

	sort.Slice(stdImports, func(i, j int) bool {
		return stdImports[i].path < stdImports[j].path
	})

	sort.Slice(otherImports, func(i, j int) bool {
		return otherImports[i].path < otherImports[j].path
	})

	result := stdImports
	result = append(result, otherImports...)
	return result
}

func loadPackageTypeData(pattern string, interfaceNames ...string) (packageTypeInfo, error) {
	loaded := loadedPackages{}
	foundPkg, err := loaded.loadPackageForInterfaces(pattern, interfaceNames...)
	if err != nil {
		fmt.Println("loadPackageForInterfaces", err)
		return packageTypeInfo{}, err
	}

	visitorData := newImportVisitorData(foundPkg.pkg.PkgPath)

	var interfaces []interfaceInfo
	for _, interfaceName := range interfaceNames {
		finder := newInterfaceInfoFinder(loaded, visitorData)

		info, err := finder.getInterfaceInfo(interfaceName, foundPkg)
		if err != nil {
			fmt.Println("getInterfaceInfo", err)
			return packageTypeInfo{}, err
		}
		interfaces = append(interfaces, info)
	}

	return packageTypeInfo{
		name:       foundPkg.pkg.Name,
		path:       foundPkg.pkg.PkgPath,
		imports:    sortImportInfos(visitorData.imports),
		interfaces: interfaces,
	}, nil
}
