package mockc

import (
	"fmt"
	"strings"
	"text/template"
)

const (
	mockTmplString = `type {{ .Name }} struct {
	mockcs struct {
	{{- range .Methods }}
		// method: {{ .Name }}
		{{ .Name }} struct {
			// basics
			Called bool
			CallCount int
		{{- if or (len .Params) (len .Results) }}
			// call history
			History []struct{
			{{- if len .Params }}
				Params struct {
				{{- range .Params }}
					{{ .String }}
				{{- end }}
				}
			{{- end }}
			{{- if len .Results}}
				Results struct {
				{{- range .Results }}
					{{ .String }}
				{{- end }}
				}
			{{- end }}
			}
		{{- end }}
		{{- if .Params }}
			// last params
			Params struct {
			{{- range .Params }}
				{{ .String }}
			{{- end }}
			}
		{{- end }}
		{{- if len .Results}}
			// last results
			Results struct {
			{{- range .Results }}
				{{ .String }}
			{{- end }}
			}
		{{- end }}
			// if Body is not nil, it is called in the middle of the method.
			Body func({{ range $idx, $val := .Params }}{{ if $idx }}, {{ end }}{{ $val.ParamType }}{{ end }}) {{ if len .Results | lt 1 }}({{ end }}{{ range $idx, $val := .Results }}{{ if $idx }}, {{ end }}{{ $val.ResultType }}{{ end }}{{ if len .Results | lt 1 }}){{ end }}
		}
	{{- end }}
	}
}
{{ range $method := .Methods }}
func (recv *{{ $.Name }}) {{ $method.Signature }} { 
	// basics
	recv.mockcs.{{ $method.Name }}.Called = true
	recv.mockcs.{{ $method.Name }}.CallCount++
{{- if len .Params }}
	// params
	{{- range $param := $method.Params }}
	recv.mockcs.{{ $method.Name }}.Params.{{ $param.Name }} = {{ $param.ParamName }}
	{{- end }}
{{- end }}
	// body
	if recv.mockcs.{{ $method.Name }}.Body != nil {
		{{ if len .Results }}{{ range $idx, $val := .Results }}{{ if $idx }}, {{ end }}recv.mockcs.{{ $method.Name }}.Results.{{ $val.Name }}{{ end }} = {{ end }}recv.mockcs.{{ $method.Name }}.Body({{ range $idx, $val := .Params }}{{ if $idx }}, {{ end }}{{ $val.ArgString }}{{ end }})
	}
{{- if or (len .Params) (len .Results) }}
	// call history
	recv.mockcs.{{ $method.Name }}.History = append(recv.mockcs.{{ $method.Name }}.History, struct{
	{{- if len .Params }}
		Params struct {
		{{- range .Params }}
			{{ .String }}
		{{- end }}
		}
	{{- end }}
	{{- if len .Results}}
		Results struct {
		{{- range .Results }}
			{{ .String }}
		{{- end }}
		}
	{{- end }}
	}{
	{{- if len .Params }}
		Params: recv.mockcs.{{ $method.Name }}.Params,
	{{- end }}
	{{- if len .Results}}
		Results: recv.mockcs.{{ $method.Name }}.Results,
	{{- end }}
	})
{{- end }}
{{- if len .Results}}
	// results
	return {{ range $idx, $val := .Results }}{{ if $idx }}, {{ end }}recv.mockcs.{{ $method.Name }}.Results.{{ $val.Name }}{{ end }}
{{- end }}
}
{{ end }}`
)

var (
	mockTmpl = template.Must(template.New("").Parse(mockTmplString))
)

type genInfo struct {
	Name    string
	Methods []methodInfo
}

type methodInfo struct {
	Name    string
	Params  []paramInfo
	Results []resultInfo
}

func (m methodInfo) Signature() string {
	params := make([]string, 0, len(m.Params))
	for _, p := range m.Params {
		params = append(params, p.ParamString())
	}
	param := fmt.Sprintf("%v", strings.Join(params, ", "))

	results := make([]string, 0, len(m.Results))
	for _, r := range m.Results {
		results = append(results, r.ResultType())
	}
	result := strings.Join(results, ", ")
	if len(m.Results) > 1 {
		result = fmt.Sprintf("(%v)", result)
	}

	return fmt.Sprintf("%v(%v) %v", m.Name, param, result)
}

type paramInfo struct {
	name, paramName, typeString string
	isVariadic                  bool
}

func (p *paramInfo) Name() string {
	return p.name
}

func (p *paramInfo) ParamName() string {
	return p.paramName
}

func (p *paramInfo) ParamType() string {
	typeString := p.typeString
	if p.isVariadic {
		typeString = fmt.Sprintf("...%v", typeString[2:])
	}

	return fmt.Sprintf("%v", typeString)
}

func (p *paramInfo) ParamString() string {
	return fmt.Sprintf("%v %v", p.ParamName(), p.ParamType())
}

func (p *paramInfo) ArgString() string {
	var variadic string
	if p.isVariadic {
		variadic = "..."
	}

	return fmt.Sprintf("%v%v", p.paramName, variadic)
}

func (p *paramInfo) String() string {
	return fmt.Sprintf("%v %v", p.Name(), p.typeString)
}

type resultInfo struct {
	name, resultName, typeString string
}

func (r *resultInfo) Name() string {
	return r.name
}

func (r *resultInfo) ResultName() string {
	return r.resultName
}

func (r *resultInfo) ResultString() string {
	return fmt.Sprintf("%v %v", r.Name(), r.ResultType())
}

func (r *resultInfo) ResultType() string {
	return r.typeString
}

func (r *resultInfo) String() string {
	return fmt.Sprintf("%v %v", r.Name(), r.typeString)
}
