package generate

import (
	"fmt"
	"path"
)

type importer struct {
	importClauses []importClause
	importPaths   map[string]int
	usedNames     map[string]int
}

type importClause struct {
	aliasName string
	path      string
	usedName  string
}

func newImporter() *importer {
	return &importer{
		importPaths: map[string]int{},
		usedNames:   map[string]int{},
	}
}

type addConfig struct {
	prefix string
}

type addOption func(opts *addConfig)

func withPreferPrefix(prefix string) addOption {
	return func(conf *addConfig) {
		conf.prefix = prefix
	}
}

func computeImporterConfig(options ...addOption) addConfig {
	conf := addConfig{}
	for _, o := range options {
		o(&conf)
	}
	return conf
}

func (i *importer) add(importDetail importInfo, options ...addOption) {
	conf := computeImporterConfig(options...)

	index, ok := i.importPaths[importDetail.path]
	if ok {
		return
	}

	clause := importClause{
		usedName: importDetail.name,
		path:     importDetail.path,
	}

	index, ok = i.usedNames[importDetail.name]
	if ok {
		dir := path.Dir(importDetail.path)

		var newName string
		if dir == "." {
			newName = "std" + importDetail.name
		} else {
			if conf.prefix == "" {
				base := path.Base(dir)
				newName = base[:1] + importDetail.name
			} else {
				newName = conf.prefix + importDetail.name
			}
		}

		prevNewName := newName
		for suffix := 1; ; suffix++ {
			_, existed := i.usedNames[newName]
			if !existed {
				break
			}
			newName = fmt.Sprintf("%s%d", prevNewName, suffix)
		}

		clause.aliasName = newName
		clause.usedName = newName
	}

	index = len(i.importClauses)

	i.importPaths[clause.path] = index
	i.usedNames[clause.usedName] = index

	i.importClauses = append(i.importClauses, clause)
}

func (i *importer) getImports() []importClause {
	result := make([]importClause, len(i.importClauses))
	copy(result, i.importClauses)
	return result
}

func (i *importer) chosenName(importPath string) string {
	index, ok := i.importPaths[importPath]
	if !ok {
		return ""
	}
	return i.importClauses[index].usedName
}
