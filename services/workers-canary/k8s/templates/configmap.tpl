---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Values.namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
data:
  {{- merge .Values.configmap.values (fromYaml (include "common.config-map" .)) | toYaml | nindent 2 }}
