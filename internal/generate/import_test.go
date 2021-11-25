package generate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImporter_Same_Path(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		aliasName: "stderrors",
		usedName:  "stderrors",
		path:      "errors",
	})

	assert.Equal(t, []importClause{
		{aliasName: "stderrors", path: "errors"},
	}, i.getImports())
	assert.Equal(t, "stderrors", i.chosenName("errors"))
	assert.Equal(t, "", i.chosenName("context"))

	i.add(importInfo{
		aliasName: "",
		usedName:  "context",
		path:      "context",
	})

	assert.Equal(t, []importClause{
		{aliasName: "stderrors", path: "errors"},
		{aliasName: "", path: "context"},
	}, i.getImports())
	assert.Equal(t, "context", i.chosenName("context"))

	i.add(importInfo{
		aliasName: "",
		usedName:  "errors",
		path:      "errors",
	})

	assert.Equal(t, []importClause{
		{aliasName: "stderrors", path: "errors"},
		{aliasName: "", path: "context"},
	}, i.getImports())
	assert.Equal(t, "stderrors", i.chosenName("errors"))
}

func TestImporter_Same_UsedName(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		usedName: "codes",
		path:     "grpc/codes",
	})
	i.add(importInfo{
		usedName: "codes",
		path:     "domain/codes",
	})

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "grpc/codes",
		},
		{
			aliasName: "dcodes",
			path:      "domain/codes",
		},
	}, i.getImports())
}

func TestImporter_Same_UsedName_With_StdLib(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		usedName: "codes",
		path:     "grpc/codes",
	})
	i.add(importInfo{
		usedName: "codes",
		path:     "codes",
	})

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "grpc/codes",
		},
		{
			aliasName: "stdcodes",
			path:      "codes",
		},
	}, i.getImports())
}

func TestImporter_Same_UsedName_Path_Multi_Levels(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		usedName: "codes",
		path:     "grpc/codes",
	})
	i.add(importInfo{
		usedName: "codes",
		path:     "sample/hello/codes",
	})

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "grpc/codes",
		},
		{
			aliasName: "hcodes",
			path:      "sample/hello/codes",
		},
	}, i.getImports())
}

func TestImporter_Same_UsedName_New_Name_Still_Existed(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		usedName: "codes",
		path:     "grpc/codes",
	})
	i.add(importInfo{
		aliasName: "hcodes",
		usedName:  "hcodes",
		path:      "sample/hello/codes",
	})
	i.add(importInfo{
		usedName: "codes",
		path:     "another/hello/codes",
	})

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "grpc/codes",
		},
		{
			aliasName: "hcodes",
			path:      "sample/hello/codes",
		},
		{
			aliasName: "hcodes1",
			path:      "another/hello/codes",
		},
	}, i.getImports())
}

func TestImporter_Same_UsedName_New_Name_Still_Existed_Suffix_2(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		usedName: "codes",
		path:     "grpc/codes",
	})
	i.add(importInfo{
		aliasName: "hcodes",
		usedName:  "hcodes",
		path:      "sample/hello/codes",
	})
	i.add(importInfo{
		aliasName: "hcodes1",
		usedName:  "hcodes1",
		path:      "another/hello/codes",
	})
	i.add(importInfo{
		usedName: "codes",
		path:     "else/hello/codes",
	})

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "grpc/codes",
		},
		{
			aliasName: "hcodes",
			path:      "sample/hello/codes",
		},
		{
			aliasName: "hcodes1",
			path:      "another/hello/codes",
		},
		{
			aliasName: "hcodes2",
			path:      "else/hello/codes",
		},
	}, i.getImports())
}

func TestImporter_Same_UsedName_With_Prefer_Prefix(t *testing.T) {
	i := newImporter()
	i.add(importInfo{
		usedName: "codes",
		path:     "sample/codes",
	})
	i.add(importInfo{
		usedName: "codes",
		path:     "opentelemetry/codes",
	}, withPreferPrefix("otel"))

	assert.Equal(t, []importClause{
		{
			aliasName: "",
			path:      "sample/codes",
		},
		{
			aliasName: "otelcodes",
			path:      "opentelemetry/codes",
		},
	}, i.getImports())
}
