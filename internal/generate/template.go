package generate

import (
	"fmt"
	"io"
	"path"
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
	tracer {{ .ChosenOtelTracer }}
	prefix string
}

// New{{ .StructName }} creates a wrapper
func New{{ .StructName }}(wrapped {{ .Name}}, tracer {{ .ChosenOtelTracer }}, prefix string) *{{ .StructName }} {
	return &{{ .StructName }}{
		{{ .UsedName }}: wrapped,
		tracer: tracer,
		prefix: prefix,
	}
}
{{ range .Methods }}
// {{ .Name }} ...
func (w *{{ $interface.StructName }}) {{ .Name }}{{ .ParamsString }}{{ .ResultsString }}{
	{{ .CtxName }}, {{ .SpanName }} := w.tracer.Start({{ .CtxName }}, w.prefix + "{{ .Name }}")
	defer {{ .SpanName }}.End()

	{{ if .WithReturn -}}
	{{ .ResultsRecvString }} = w.{{ $interface.UsedName }}.{{ .Name }}({{ .ArgsString }})
	{{ if .WithError -}}
	if {{ .ErrString }} != nil {
		{{ .SpanName }}.RecordError({{ .ErrString }})
		{{ .SpanName }}.SetStatus({{ .ChosenOtelCodes }}, {{ .ErrString }}.Error())
	}
	{{- end }}
	return {{ .ResultsRecvString }}
	{{- else -}}
	w.{{ $interface.UsedName }}.{{ .Name }}({{ .ArgsString }})
	{{- end }}
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

	ParamsString  string
	ResultsString string
	ArgsString    string

	WithReturn        bool
	WithError         bool
	ResultsRecvString string
	ErrString         string
	ChosenOtelCodes   string
}

type templateInterface struct {
	Name             string
	UsedName         string
	StructName       string
	Methods          []templateMethod
	ChosenOtelTracer string
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
	globalVariables map[string]emptyStruct
	interfaces      []templateInterfaceVariables
}

func collectVariables(info packageTypeInfo) templateVariables {
	global := map[string]struct{}{}
	global[info.name] = struct{}{}

	for _, importDetail := range info.imports {
		global[importDetail.name] = struct{}{}
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

func nameIsEmpty(name string) bool {
	return name == "" || name == "_"
}

func assignVariableNamesForFields(
	global map[string]struct{},
	local map[string]recognizedType,
	fieldList []tupleType,
	startPosition int,
) {
	for i, field := range fieldList {
		_, globalExisted := global[field.name]
		if !nameIsEmpty(field.name) && !globalExisted && field.name != "w" {
			continue
		}

		varName := getVariableName(
			global, local,
			i-startPosition, field.recognized,
		)
		fieldList[i].name = varName
		local[varName] = field.recognized
	}
}

func assignVariableNamesForMethod(
	global map[string]struct{},
	local map[string]recognizedType,
	method methodType,
) {
	assignVariableNamesForFields(global, local, method.params, 1)
	assignVariableNamesForFields(global, local, method.results, 0)
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

func replacePackageName(typeStr string, pkgList []tupleTypePkg, importController *importer) string {
	var buf strings.Builder

	var fromOffset int

	replace := func(pkg tupleTypePkg, newName string) {
		_, _ = buf.WriteString(typeStr[fromOffset:pkg.begin])
		_, _ = buf.WriteString(newName)
		fromOffset = pkg.end
	}

	for _, pkg := range pkgList {
		chosenName := importController.chosenName(pkg.path)
		if pkg.begin == pkg.end && chosenName != "" {
			replace(pkg, chosenName+".")
		} else {
			replace(pkg, chosenName)
		}
	}

	_, _ = buf.WriteString(typeStr[fromOffset:])

	return buf.String()
}

func generateFieldListString(fields []tupleType, importController *importer) string {
	var fieldList []string

	for _, f := range fields {
		modifiedTypeStr := replacePackageName(f.typeStr, f.pkgList, importController)
		s := fmt.Sprintf("%s %s", f.name, modifiedTypeStr)
		fieldList = append(fieldList, s)
	}

	return strings.Join(fieldList, ", ")
}

func generateArgsString(fields []tupleType) string {
	var args []string
	for _, field := range fields {
		name := field.name
		if field.isVariadic {
			name = name + "..."
		}
		args = append(args, name)
	}
	return strings.Join(args, ", ")
}

const (
	otelTracePkgPath = "go.opentelemetry.io/otel/trace"
	otelCodesPkgPath = "go.opentelemetry.io/otel/codes"
)

func generateCodeForMethod(
	global map[string]struct{},
	local map[string]recognizedType,
	method methodType,
	importController *importer,
) templateMethod {
	paramsStr := generateFieldListString(method.params, importController)
	paramsStr = fmt.Sprintf("(%s)", paramsStr)

	ctxName := ""
	for _, param := range method.params {
		if param.recognized == recognizedTypeContext {
			ctxName = param.name
			break
		}
	}

	var resultsStr string
	if len(method.results) == 0 {
		resultsStr = " "
	} else {
		resultsStr = generateFieldListString(method.results, importController)
		resultsStr = fmt.Sprintf(" (%s) ", resultsStr)
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

		ParamsString:  paramsStr,
		ResultsString: resultsStr,
		ArgsString:    generateArgsString(method.params),

		WithReturn:        resultsStr != " ",
		WithError:         errStr != "",
		ResultsRecvString: strings.Join(recvVars, ", "),
		ErrString:         errStr,
		ChosenOtelCodes: replacePackageName("codes.Error", []tupleTypePkg{
			{
				path:  otelCodesPkgPath,
				begin: 0,
				end:   len("codes"),
			},
		}, importController),
	}
}

//revive:disable-next-line:flag-parameter
func importControllerAddImports(importController *importer, imports []importInfo, addOtelCodes bool) {
	for _, importDetail := range imports {
		importController.add(importDetail)
	}

	importController.add(importInfo{
		path: otelTracePkgPath,
		name: "trace",
	}, withPreferPrefix("otel"))

	if addOtelCodes {
		importController.add(importInfo{
			path: otelCodesPkgPath,
			name: "codes",
		}, withPreferPrefix("otel"))
	}
}

type generateConfig struct {
	inAnotherPackage bool
	pkgName          string
}

// Option ...
type Option func(conf *generateConfig)

// WithInAnotherPackage ...
func WithInAnotherPackage(packageName string) Option {
	return func(conf *generateConfig) {
		conf.inAnotherPackage = true
		conf.pkgName = packageName
	}
}

func computeGenerateConfig(options ...Option) generateConfig {
	conf := generateConfig{
		inAnotherPackage: false,
	}
	for _, o := range options {
		o(&conf)
	}
	return conf
}

func containsErrorReturns(info packageTypeInfo) bool {
	for _, interfaceDetail := range info.interfaces {
		for _, method := range interfaceDetail.methods {
			for _, result := range method.results {
				if result.recognized == recognizedTypeError {
					return true
				}
			}
		}
	}
	return false
}

func generateCode(writer io.Writer, info packageTypeInfo, options ...Option) error {
	conf := computeGenerateConfig(options...)

	importController := newImporter()
	if conf.inAnotherPackage {
		importController.add(importInfo{
			path: info.path,
			name: path.Base(info.path),
		})
	}
	addOtelCodes := containsErrorReturns(info)
	importControllerAddImports(importController, info.imports, addOtelCodes)

	controllerImports := importController.getImports()
	newImports := make([]importInfo, 0, len(controllerImports))
	for _, clause := range controllerImports {
		newImports = append(newImports, importInfo{
			name: clause.usedName,
			path: clause.path,
		})
	}
	info.imports = newImports

	variables := collectVariables(info)
	info = assignVariableNames(info)

	global := variables.globalVariables

	var interfaces []templateInterface
	for interfaceIndex, interfaceDetail := range info.interfaces {
		var methods []templateMethod
		for methodIndex, method := range interfaceDetail.methods {
			if len(method.params) == 0 || method.params[0].recognized != recognizedTypeContext {
				continue
			}
			local := variables.interfaces[interfaceIndex].methods[methodIndex].variables
			methods = append(methods, generateCodeForMethod(global, local, method, importController))
		}

		embeddedInterfaceName := replacePackageName(interfaceDetail.name,
			[]tupleTypePkg{
				{
					path:  info.path,
					begin: 0,
					end:   0,
				},
			},
			importController,
		)
		interfaces = append(interfaces, templateInterface{
			Name:       embeddedInterfaceName,
			UsedName:   interfaceDetail.name,
			StructName: interfaceDetail.name + "Wrapper",
			Methods:    methods,
			ChosenOtelTracer: replacePackageName("trace.Tracer", []tupleTypePkg{
				{
					path:  otelTracePkgPath,
					begin: 0,
					end:   len("trace"),
				},
			}, importController),
		})
	}

	var importStmts []string
	for _, clause := range importController.getImports() {
		if clause.aliasName == "" {
			importStmts = append(importStmts, fmt.Sprintf(`"%s"`, clause.path))
		} else {
			importStmts = append(importStmts, fmt.Sprintf(`%s "%s"`, clause.aliasName, clause.path))
		}
	}

	packageName := info.name
	if conf.pkgName != "" {
		packageName = conf.pkgName
	}

	return resultTemplate.Execute(writer, templatePackageInfo{
		PackageName: packageName,
		Imports:     importStmts,
		Interfaces:  interfaces,
	})
}

// LoadAndGenerate ...
func LoadAndGenerate(w io.Writer, pattern string, interfaceNames []string, options ...Option) error {
	info, err := loadPackageTypeData(pattern, interfaceNames...)
	if err != nil {
		return err
	}
	err = generateCode(w, info, options...)
	if err != nil {
		fmt.Println("generateCode", err)
		return err
	}
	return nil
}
