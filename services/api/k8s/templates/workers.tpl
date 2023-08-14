{{- range .Values.workers }}
{{- $worker := dict "worker" . -}}
{{- $data := deepCopy $ | merge $worker -}}
{{- $_ := set $data.Values "command" .command -}}
{{- $_ := set $data.Values "replicaCount" .replicas -}}
{{- $_ := set $data.Values.probes.liveness "enabled" false -}}
{{- $_ := set $data.Values.probes.readiness "enabled" false -}}
{{ include "lib.deployment" $data }}
{{- end}}
