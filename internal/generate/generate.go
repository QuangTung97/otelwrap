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

type packageTypeTuple struct {
	name       string
	typeStr    string
	recognized recognizedType
}

type packageTypeMethod struct {
	name    string
	params  []packageTypeTuple
	results []packageTypeTuple
}

type packageImportInfo struct {
	aliasName string
	path      string
}

type packageTypeInfo struct {
	interfaceName string
	imports       []packageImportInfo
	methods       []packageTypeMethod
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
) []packageTypeTuple {
	if fileList == nil {
		return nil
	}

	var tuples []packageTypeTuple
	for _, field := range fileList.List {
		begin := field.Type.Pos()
		end := field.Type.End()
		file := fset.File(begin)

		filename := file.Name()
		typeStr := fileMap[filename][file.Offset(begin):file.Offset(end)]

		recognized := getRecognizedType(field, info)
		for _, resultName := range field.Names {
			tuples = append(tuples, packageTypeTuple{
				name:       resultName.Name,
				typeStr:    typeStr,
				recognized: recognized,
			})
		}
		if len(field.Names) == 0 {
			tuples = append(tuples, packageTypeTuple{
				typeStr:    typeStr,
				recognized: recognized,
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

func getImportInfos(syntaxFiles []*ast.File, acceptedPackages map[string]struct{}) []packageImportInfo {
	var imports []packageImportInfo
	for _, syntax := range syntaxFiles {
		for _, importInfo := range syntax.Imports {
			pathValue, err := strconv.Unquote(importInfo.Path.Value)
			if err != nil {
				panic(err)
			}

			aliasName := ""
			usedName := path.Base(pathValue)
			if importInfo.Name != nil {
				aliasName = importInfo.Name.Name
				usedName = aliasName
			}

			if _, ok := acceptedPackages[usedName]; !ok {
				continue
			}

			imports = append(imports, packageImportInfo{
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

func loadPackageTypeData(pattern string, interfaceName string) (packageTypeInfo, error) {
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
		if pkg.Types.Scope().Lookup(interfaceName) != nil {
			foundPkg = pkg
			break
		}
	}

	if foundPkg == nil {
		return packageTypeInfo{}, fmt.Errorf("can not find interface '%s'", interfaceName)
	}

	fileMap := readFiles(foundPkg.CompiledGoFiles)
	interfaceType, err := findInterfaceType(interfaceName, foundPkg.Syntax)
	if err != nil {
		return packageTypeInfo{}, err
	}

	visitor := newImportVisitor(foundPkg.TypesInfo)
	ast.Walk(visitor, interfaceType)

	imports := getImportInfos(foundPkg.Syntax, visitor.packageNames)

	methods := make([]packageTypeMethod, 0, len(interfaceType.Methods.List))
	for _, field := range interfaceType.Methods.List {
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		params := fieldListToTupleList(funcType.Params, foundPkg.Fset, fileMap, foundPkg.TypesInfo)
		results := fieldListToTupleList(funcType.Results, foundPkg.Fset, fileMap, foundPkg.TypesInfo)

		methods = append(methods, packageTypeMethod{
			name:    field.Names[0].Name,
			params:  params,
			results: results,
		})
	}

	return packageTypeInfo{
		interfaceName: interfaceName,
		imports:       imports,
		methods:       methods,
	}, nil
}
