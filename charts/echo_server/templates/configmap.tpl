---
apiVersion: v1
kind: ConfigMap
metadata:
  name: echo-server
  namespace: {{ .Release.Namespace }}
  labels:
    foo: bar
data:
{{- .Values.env | toYaml | nindent 2 }}
