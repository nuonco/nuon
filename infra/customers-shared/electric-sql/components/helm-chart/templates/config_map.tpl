---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace | quote | default "default"}}
  labels:
    {{- include "common.labels" . | nindent 4 }}
data:
{{- range $path, $config := .Values.env }}
  {{ $path }}: |
{{ $config | indent 4 -}}
{{- end -}}
