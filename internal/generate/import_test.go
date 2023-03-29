package generate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImporter_Same_Path(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		name: "stderrors",
		path: "errors",
	})

	assert.Equal(t, []importClause{
		{
			path:     "errors",
			usedName: "stderrors",
		},
	}, i.getImports())
	assert.Equal(t, "stderrors", i.chosenName("errors"))
	assert.Equal(t, "", i.chosenName("context"))

	i.add(importInfo{
		name: "context",
		path: "context",
	})

	assert.Equal(t, []importClause{
		{path: "errors", usedName: "stderrors"},
		{path: "context", usedName: "context"},
	}, i.getImports())
	assert.Equal(t, "context", i.chosenName("context"))

	i.add(importInfo{
		name: "errors",
		path: "errors",
	})

	assert.Equal(t, []importClause{
		{path: "errors", usedName: "stderrors"},
		{aliasName: "", path: "context", usedName: "context"},
	}, i.getImports())
	assert.Equal(t, "stderrors", i.chosenName("errors"))
}

func TestImporter_Same_UsedName(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		name: "codes",
		path: "grpc/codes",
	})
	i.add(importInfo{
		name: "codes",
		path: "domain/codes",
	})

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "grpc/codes",
			usedName:  "codes",
		},
		{
			aliasName: "dcodes",
			path:      "domain/codes",
			usedName:  "dcodes",
		},
	}, i.getImports())
}

func TestImporter_Same_UsedName_With_StdLib(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		name: "codes",
		path: "grpc/codes",
	})
	i.add(importInfo{
		name: "codes",
		path: "codes",
	})

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "grpc/codes",
			usedName:  "codes",
		},
		{
			aliasName: "stdcodes",
			path:      "codes",
			usedName:  "stdcodes",
		},
	}, i.getImports())
}

func TestImporter_Same_UsedName_Path_Multi_Levels(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		name: "codes",
		path: "grpc/codes",
	})
	i.add(importInfo{
		name: "codes",
		path: "sample/hello/codes",
	})

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "grpc/codes",
			usedName:  "codes",
		},
		{
			aliasName: "hcodes",
			path:      "sample/hello/codes",
			usedName:  "hcodes",
		},
	}, i.getImports())
}

func TestImporter_Same_UsedName_New_Name_Still_Existed(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		name: "codes",
		path: "grpc/codes",
	})
	i.add(importInfo{
		name: "codes",
		path: "sample/hello/codes",
	})
	i.add(importInfo{
		name: "codes",
		path: "another/hello/codes",
	})

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "grpc/codes",
			usedName:  "codes",
		},
		{
			aliasName: "hcodes",
			path:      "sample/hello/codes",
			usedName:  "hcodes",
		},
		{
			aliasName: "hcodes1",
			path:      "another/hello/codes",
			usedName:  "hcodes1",
		},
	}, i.getImports())
}

func TestImporter_Same_UsedName_New_Name_Still_Existed_Suffix_2(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		name: "codes",
		path: "grpc/codes",
	})
	i.add(importInfo{
		name: "hcodes",
		path: "sample/hello/codes",
	})
	i.add(importInfo{
		name: "hcodes1",
		path: "another/hello/codes",
	})
	i.add(importInfo{
		name: "codes",
		path: "else/hello/codes",
	})

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "grpc/codes",
			usedName:  "codes",
		},
		{
			path:     "sample/hello/codes",
			usedName: "hcodes",
		},
		{
			path:     "another/hello/codes",
			usedName: "hcodes1",
		},
		{
			aliasName: "hcodes2",
			path:      "else/hello/codes",
			usedName:  "hcodes2",
		},
	}, i.getImports())
}

func TestImporter_Same_UsedName_With_Prefer_Prefix(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		name: "codes",
		path: "sample/codes",
	})
	i.add(importInfo{
		name: "codes",
		path: "opentelemetry/codes",
	}, withPreferPrefix("otel"))

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "sample/codes",
			usedName:  "codes",
		},
		{
			aliasName: "otelcodes",
			path:      "opentelemetry/codes",
			usedName:  "otelcodes",
		},
	}, i.getImports())
}
