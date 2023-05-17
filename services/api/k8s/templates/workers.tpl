{{- range .Values.workers }}
{{- $worker := dict "worker" . -}}
{{- $data := deepCopy $ | merge $worker -}}
{{- $_ := set $data.Values "command" .command -}}
{{- $_ := set $data.Values "replicaCount" .replicas -}}
{{ include "lib.deployment" $data }}
{{- end}}
