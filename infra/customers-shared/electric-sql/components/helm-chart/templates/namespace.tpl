---
apiVersion: v1
kind: Namespace
metadata:
  name: sync-service
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
