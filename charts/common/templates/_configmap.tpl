{{- define "common.configmap.tpl" -}}
{{- if .Values.configmap.enabled }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace | quote |default "default"}}
  labels:
    {{- include "common.labels" . | nindent 4 }}
data:
{{- merge .Values.configmap.values (fromYaml (include "common.config-map" .)) | toYaml | nindent 2 }}
{{- end -}}
{{- end -}}

{{- define "common.configmap" -}}
{{- include "common.util.merge" (append . "common.configmap.tpl") -}}
{{- end -}}

