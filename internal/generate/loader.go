package generate

import "golang.org/x/tools/go/packages"

type loadedPackage struct {
	fileMap map[string]string
	pkg     *packages.Package
}

type loadedPackages map[string]loadedPackage

const loadPackageMode = packages.NeedName | packages.NeedSyntax | packages.NeedCompiledGoFiles |
	packages.NeedTypes | packages.NeedTypesInfo

func (loaded loadedPackages) loadPackageForInterfaces(
	pattern string, interfaceNames ...string,
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
