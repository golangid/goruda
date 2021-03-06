// GENERATED BY goruda
// This file was generated automatically at
// {{ .TimeStamp }}

package {{.PackageName}}

type {{ .Name }} interface {
{{- range $key, $val := .Methods }}
	{{ $key | camelcase }}({{ range $index, $element := $val.Attributes }}{{ $element.Name }} {{ $element.Type }}{{ if ne $index $val.Attributes.GetLastIndex }},{{ end }}{{ end }}){{ if eq (len $val.ReturnValue) 0 }} error {{ else }} ({{ range $index, $element := .ReturnValue }}{{ $element.Type }}{{ if ne $index $val.ReturnValue.GetLastIndex }},{{ end }}{{ end }}, error){{ end }}
{{- end }}
}
{{ $receiverName := print .Name "Implementation" }}
type {{ $receiverName }} struct {
}{{ if gt (len .Methods) 0 }}
{{ range $key, $val := .Methods }}
func (s {{ $receiverName }}) {{ $key | camelcase }}({{ range $index, $element := .Attributes }}{{ $element.Name }} {{ $element.Type }}{{ if ne $index $val.Attributes.GetLastIndex }},{{ end }}{{ end }}){{ if eq (len $val.ReturnValue) 0 }} error {{ else }} ({{ range $index, $element := .ReturnValue }}{{ $element.Type }}{{ if ne $index $val.ReturnValue.GetLastIndex }},{{ end }}{{ end }}, error){{ end }}{
	 return {{ if gt (len .ReturnValue) 0 }}{{ range $index, $element := .ReturnValue }}{{ $element.Type }}{{ if ne $index $val.ReturnValue.GetLastIndex }},{{ end }}{{ end }}{},{{end}} nil
}
{{ end }}
{{ end }}
