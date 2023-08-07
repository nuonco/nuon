{{- define "common.workerLabels" -}}
app: {{ .Release.Name }}-worker
helm.sh/chart: {{ include "common.chart" . }}
{{ include "common.workerSelectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "common.workerSelectorLabels" -}}
app.kubernetes.io/name: {{ include "common.name" . }}-worker
app.kubernetes.io/instance: {{ .Release.Name }}-worker
{{- end }}
