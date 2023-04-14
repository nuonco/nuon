{{- $workers := dict "workers" . -}}
{{- $data := deepCopy $ | merge $workers -}}
{{- $_ := set $data.Values "command" .command -}}
{{- $_ := set $data.Values "replicaCount" .replicas -}}
{{ include "common.deployment" $data }}
