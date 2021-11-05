package generate

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

var templateString = `
package {{ .PackageName }}

import (
{{- range .Imports }}
	{{ . }}{{ end }}
)
{{ range $interface := .Interfaces }}
// {{ .StructName }} wraps OpenTelemetry's span
type {{ .StructName }} struct {
	{{ .Name }}
	tracer trace.Tracer
	prefix string
}
{{ range .Methods }}
// {{ .Name }} ...
func (w *{{ $interface.StructName }}) {{ .Name }}{{ .ParamsString }} {{ .ResultsString }} {
	ctx, span := w.tracer.Start(ctx, w.prefix + "{{ .Name }}")
	defer span.End()

	err := w.{{ $interface.Name }}.{{ .Name }}({{ .ArgsString }})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}
{{ end -}}
{{ end -}}
`

func initTemplate() *template.Template {
	tmpl, err := template.New("otelwrap").Parse(templateString)
	if err != nil {
		panic(err)
	}
	return tmpl
}

var resultTemplate = initTemplate()

type templateMethod struct {
	Name          string
	ParamsString  string
	ResultsString string
	ArgsString    string
}

type templateInterface struct {
	Name       string
	StructName string
	Methods    []templateMethod
}

type templatePackageInfo struct {
	PackageName string
	Imports     []string
	Interfaces  []templateInterface
}

func generateFieldListString(fields []tupleType) (string, bool) {
	var fieldList []string

	needBracket := false
	if len(fields) > 1 {
		needBracket = true
	}

	for _, f := range fields {
		var s string
		if f.name == "" {
			s = fmt.Sprintf("%s", f.typeStr)
		} else {
			needBracket = true
			s = fmt.Sprintf("%s %s", f.name, f.typeStr)
		}
		fieldList = append(fieldList, s)
	}

	return strings.Join(fieldList, ", "), needBracket
}

func generateArgsString(fields []tupleType) string {
	var args []string
	for _, field := range fields {
		args = append(args, field.name)
	}
	return strings.Join(args, ", ")
}

func generateCode(writer io.Writer, info packageTypeInfo) error {
	var imports []string
	for _, importDetail := range info.imports {
		imports = append(imports, fmt.Sprintf(`"%s"`, importDetail.path))
	}
	imports = append(imports, `"go.opentelemetry.io/otel/trace"`)

	var interfaces []templateInterface
	for _, interfaceDetail := range info.interfaces {
		var methods []templateMethod
		for _, method := range interfaceDetail.methods {
			paramsStr, _ := generateFieldListString(method.params)
			paramsStr = fmt.Sprintf("(%s)", paramsStr)

			resultsStr, needBracket := generateFieldListString(method.results)
			if needBracket {
				resultsStr = fmt.Sprintf("(%s)", resultsStr)
			}

			methods = append(methods, templateMethod{
				Name:          method.name,
				ParamsString:  paramsStr,
				ResultsString: resultsStr,
				ArgsString:    generateArgsString(method.params),
			})
		}

		interfaces = append(interfaces, templateInterface{
			Name:       interfaceDetail.name,
			StructName: interfaceDetail.name + "Wrapper",
			Methods:    methods,
		})
	}

	return resultTemplate.Execute(writer, templatePackageInfo{
		PackageName: info.name,
		Imports:     imports,
		Interfaces:  interfaces,
	})
}
