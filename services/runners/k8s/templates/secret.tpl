---
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
data:
{{- include "common.secret" . | nindent 2 }}
