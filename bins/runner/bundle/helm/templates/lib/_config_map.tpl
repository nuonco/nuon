{{- define "common.config-map" -}}
CONTAINER_PLATFORM: "EKS"
SERVICE_NAME: {{ .Chart.Name }}
SERVICE_VERSION: {{ .Chart.Version }}
{{- end -}}
