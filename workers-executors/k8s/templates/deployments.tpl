{{- range .Values.instances }}
{{- $instance := dict "instance" . -}}
{{- $data := deepCopy $ | merge $instance -}}
{{- $_ := set $data.Values "command" .command -}}
{{ include "lib.deployment" $data }}
{{- end}}
