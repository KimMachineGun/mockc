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
		{{- if len .Params }}
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
	{{- end }}
	{{- range $param := $method.Params }}
	recv.mockcs.{{ $method.Name }}.Params.{{ $param.Name }} = {{ $param.Name }}
	{{- end }}
	{{- if len .Results}}
	// body
	if recv.mockcs.{{ $method.Name }}.Body != nil {
		{{ if len .Results }}{{ range $idx, $val := .Results }}{{ if $idx }}, {{ end }}recv.mockcs.{{ $method.Name }}.Results.{{ $val.Name }}{{ end }} = {{ end }}recv.mockcs.{{ $method.Name }}.Body({{ range $idx, $val := .Params }}{{ if $idx }}, {{ end }}{{ $val.ArgString }}{{ end }})
	}
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
	Name, TypeString string
	IsVariadic       bool
}

func (p *paramInfo) ParamString() string {
	return fmt.Sprintf("%v %v", p.Name, p.ParamType())
}

func (p *paramInfo) ParamType() string {
	typeString := p.TypeString
	if p.IsVariadic {
		typeString = fmt.Sprintf("...%v", typeString[2:])
	}

	return fmt.Sprintf("%v", typeString)
}

func (p *paramInfo) ArgString() string {
	var variadic string
	if p.IsVariadic {
		variadic = "..."
	}

	return fmt.Sprintf("%v%v", p.Name, variadic)
}

func (p *paramInfo) String() string {
	return fmt.Sprintf("%v %v", p.Name, p.TypeString)
}

type resultInfo struct {
	Name, TypeString string
}

func (r *resultInfo) ResultString() string {
	return fmt.Sprintf("%v %v", r.Name, r.ResultType())
}

func (r *resultInfo) ResultType() string {
	return r.TypeString
}

func (r *resultInfo) String() string {
	return fmt.Sprintf("%v %v", r.Name, r.TypeString)
}
