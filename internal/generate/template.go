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
	{{ .CtxName }}, {{ .SpanName }} := w.tracer.Start({{ .CtxName }}, w.prefix + "{{ .Name }}")
	defer {{ .SpanName }}.End()

	{{ .ResultsRecvString }} = w.{{ $interface.Name }}.{{ .Name }}({{ .ArgsString }})
	if {{ .ErrString }} != nil {
		{{ .SpanName }}.RecordError({{ .ErrString }})
		{{ .SpanName }}.SetStatus(codes.Error, {{ .ErrString }}.Error())
	}
	return {{ .ResultsRecvString }}
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
	Name     string
	CtxName  string
	SpanName string

	ParamsString      string
	ResultsString     string
	ArgsString        string
	ResultsRecvString string // including assignment
	ErrString         string
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

type templateMethodVariables struct {
	variables map[string]recognizedType
}

type templateInterfaceVariables struct {
	name    string
	methods []templateMethodVariables
}

type templateVariables struct {
	globalVariables map[string]struct{}
	interfaces      []templateInterfaceVariables
}

func collectVariables(info packageTypeInfo) templateVariables {
	global := map[string]struct{}{}
	global[info.name] = struct{}{}

	for _, importDetail := range info.imports {
		global[importDetail.usedName] = struct{}{}
	}

	interfaces := make([]templateInterfaceVariables, 0, len(info.interfaces))
	for _, interfaceDetail := range info.interfaces {
		global[interfaceDetail.name] = struct{}{}

		var methods []templateMethodVariables
		for _, method := range interfaceDetail.methods {
			variables := map[string]recognizedType{
				method.name: recognizedTypeUnknown,
			}

			for _, param := range method.params {
				variables[param.name] = param.recognized
			}

			methods = append(methods, templateMethodVariables{
				variables: variables,
			})
		}

		interfaces = append(interfaces, templateInterfaceVariables{
			name:    interfaceDetail.name,
			methods: methods,
		})
	}

	return templateVariables{
		globalVariables: global,
		interfaces:      interfaces,
	}
}

func assignVariableNamesForMethod(
	global map[string]struct{},
	local map[string]recognizedType,
	method methodType,
) {
	for i, param := range method.params {
		if param.name != "" {
			continue
		}

		varName := getVariableName(
			global, local,
			i-1, param.recognized,
		)
		method.params[i].name = varName
		local[varName] = param.recognized
	}

	for i, result := range method.results {
		if result.name != "" {
			continue
		}

		varName := getVariableName(
			global, local,
			i, result.recognized,
		)
		method.results[i].name = varName
	}
}

func assignVariableNames(info packageTypeInfo) packageTypeInfo {
	variables := collectVariables(info)

	for interfaceIndex, interfaceDetail := range info.interfaces {
		for methodIndex, method := range interfaceDetail.methods {
			local := variables.interfaces[interfaceIndex].methods[methodIndex].variables
			assignVariableNamesForMethod(variables.globalVariables, local, method)
		}
	}
	return info
}

func getNextVariableName(name string, index int) string {
	if index == 0 {
		return name
	}
	return fmt.Sprintf("%s%d", name, index)
}

func getVariableName(
	global map[string]struct{},
	local map[string]recognizedType,
	index int, expectedType recognizedType,
) string {
	var recommendedName string
	switch expectedType {
	case recognizedTypeContext:
		recommendedName = "ctx"
	case recognizedTypeError:
		recommendedName = "err"
	case recognizedTypeSpan:
		recommendedName = "span"
	default:
		ch := 'a' + index
		recommendedName = fmt.Sprintf("%c", ch)
	}

	for retryIndex := 0; ; retryIndex++ {
		name := getNextVariableName(recommendedName, retryIndex)
		if _, existed := global[name]; existed {
			continue
		}
		if _, existed := local[name]; existed {
			continue
		}
		return name
	}
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

func preventShadowMethodRecv(
	global map[string]struct{},
	local map[string]recognizedType,
	tuples []tupleType,
) {
	for i, t := range tuples {
		if t.name == "w" {
			newName := getVariableName(global, local, 0, t.recognized)
			tuples[i].name = newName
		}
	}
}

func generateCodeForMethod(
	global map[string]struct{},
	local map[string]recognizedType,
	method methodType,
) templateMethod {
	preventShadowMethodRecv(global, local, method.params)
	preventShadowMethodRecv(global, local, method.results)

	paramsStr, _ := generateFieldListString(method.params)
	paramsStr = fmt.Sprintf("(%s)", paramsStr)

	ctxName := ""
	for _, param := range method.params {
		if param.recognized == recognizedTypeContext {
			ctxName = param.name
			break
		}
	}

	resultsStr, needBracket := generateFieldListString(method.results)
	if needBracket {
		resultsStr = fmt.Sprintf("(%s)", resultsStr)
	}

	errStr := ""
	var recvVars []string
	for _, result := range method.results {
		recvVars = append(recvVars, result.name)
		if result.recognized == recognizedTypeError {
			errStr = result.name
		}
	}

	spanName := getVariableName(global, local, 0, recognizedTypeSpan)

	return templateMethod{
		Name:     method.name,
		CtxName:  ctxName,
		SpanName: spanName,

		ParamsString:      paramsStr,
		ResultsString:     resultsStr,
		ArgsString:        generateArgsString(method.params),
		ResultsRecvString: strings.Join(recvVars, ", "),
		ErrString:         errStr,
	}
}

func generateCode(writer io.Writer, info packageTypeInfo) error {
	info = assignVariableNames(info)

	var imports []string
	for _, importDetail := range info.imports {
		imports = append(imports, fmt.Sprintf(`"%s"`, importDetail.path))
	}
	imports = append(imports, `"go.opentelemetry.io/otel/trace"`)

	variables := collectVariables(info)
	global := variables.globalVariables

	var interfaces []templateInterface
	for interfaceIndex, interfaceDetail := range info.interfaces {
		var methods []templateMethod
		for methodIndex, method := range interfaceDetail.methods {
			local := variables.interfaces[interfaceIndex].methods[methodIndex].variables
			methods = append(methods, generateCodeForMethod(global, local, method))
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
