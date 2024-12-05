---
apiVersion: v1
kind: ConfigMap
metadata:
  name: helm-configmap
  namespace: {{ .Release.Namespace | quote | default "default"}}
  labels:
    foo: bar
data:
{{- .Values.env | toYaml | nindent 2 }}
