{{- define "common.secret" -}}
NUON_API_TOKEN: {{ .Values.nuon_api_token | b64enc }}
{{- end -}}
