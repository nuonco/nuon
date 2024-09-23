{{- define "common.config-map" -}}
ENVIRONMENT: {{ .Values.environment | quote }}
ENV: {{ .Values.environment | quote }}
CONTAINER_PLATFORM: "EKS"
SERVICE_NAME: {{ .Chart.Name }}
SERVICE_VERSION: {{ .Chart.Version }}
{{- end -}}
