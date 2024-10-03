---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
data:
{{- merge .Values.env (fromYaml (include "common.config-map" .)) | toYaml | nindent 2 }}
