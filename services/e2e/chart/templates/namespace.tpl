---
apiVersion: v1
kind: Namespace
metadata:
  name: e2e
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
