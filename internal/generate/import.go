package generate

import (
	"fmt"
	"path"
)

type importer struct {
	infos       []importInfo
	importPaths map[string]int
	usedNames   map[string]int
}

type importClause struct {
	aliasName string
	path      string
}

func newImporter() *importer {
	return &importer{
		importPaths: map[string]int{},
		usedNames:   map[string]int{},
	}
}

func (i *importer) add(importDetail importInfo) {
	index, ok := i.importPaths[importDetail.path]
	if ok {
		return
	}

	index, ok = i.usedNames[importDetail.usedName]
	if ok {
		dir := path.Dir(importDetail.path)

		var newName string
		if dir == "." {
			newName = "std" + importDetail.usedName
		} else {
			base := path.Base(dir)
			newName = base[:1] + importDetail.usedName
		}

		prevNewName := newName
		for suffix := 1; ; suffix++ {
			_, existed := i.usedNames[newName]
			if !existed {
				break
			}
			newName = fmt.Sprintf("%s%d", prevNewName, suffix)
		}

		importDetail.aliasName = newName
		importDetail.usedName = newName
	}

	index = len(i.infos)

	i.importPaths[importDetail.path] = index
	i.usedNames[importDetail.usedName] = index

	i.infos = append(i.infos, importDetail)
}

func (i *importer) getImports() []importClause {
	var result []importClause
	for _, info := range i.infos {
		result = append(result, importClause{
			aliasName: info.aliasName,
			path:      info.path,
		})
	}
	return result
}

func (i *importer) chosenName(importPath string) string {
	index, ok := i.importPaths[importPath]
	if !ok {
		return ""
	}
	return i.infos[index].usedName
}
