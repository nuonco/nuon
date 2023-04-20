{{- /*
The pod definition included in the controller.
*/ -}}
{{- define "common.pod" -}}
{{- if .Values.imagePullSecrets }}
  {{- with .Values.imagePullSecrets }}
imagePullSecrets:
    {{- toYaml . | nindent 2 }}
  {{- end }}
{{- end }}
{{- if .Values.serviceAccount.enabled }}
serviceAccountName: {{ include "common.serviceAccountName" . }}
{{- end }}
{{- if .Values.automountServiceAccountToken}}
automountServiceAccountToken: {{ .Values.automountServiceAccountToken }}
{{- end }}
{{- if .Values.podSecurityContext}}
  {{- with .Values.podSecurityContext }}
securityContext:
    {{- toYaml . | nindent 2 }}
  {{- end }}
{{- end }}
{{- if .Values.initContainers }}
  {{- with .Values.initContainers }}
initContainers:
    {{- $initContainers := list }}
    {{- range $index, $key := (keys .Values.initContainers | uniq | sortAlpha) }}
      {{- $container := get $.Values.initContainers $key }}
      {{- if not $container.name -}}
        {{- $_ := set $container "name" $key }}
      {{- end }}
      {{- $initContainers = append $initContainers $container }}
    {{- end }}
    {{- tpl (toYaml $initContainers) $ | nindent 2 }}
  {{- end }}
{{- end }}
containers:
  {{- include "common.container" . | nindent 2 }}
  {{- if .Values.additionalContainers }}
  {{- with .Values.additionalContainers }}
    {{- $additionalContainers := list }}
    {{- range $name, $container := . }}
      {{- if not $container.name -}}
        {{- $_ := set $container "name" $name }}
      {{- end }}
      {{- $additionalContainers = append $additionalContainers $container }}
    {{- end }}
    {{- tpl (toYaml $additionalContainers) $ | nindent 2 }}
    {{- end }}
  {{- end }}
{{/*   {{- with (include "common.volumes" . | trim) }} */}}
{{/* volumes: */}}
{{/*     {{- nindent 2 . }} */}}
{{/*   {{- end }} */}}
{{/*   {{- with .Values.hostAliases }} */}}
{{/* hostAliases: */}}
{{/*     {{- toYaml . | nindent 2 }} */}}
{{/*   {{- end }} */}}
{{/*   {{- with .Values.nodeSelector }} */}}
{{/* nodeSelector: */}}
{{/*     {{- toYaml . | nindent 2 }} */}}
{{/*   {{- end }} */}}
{{- if .Values.affinity }}
  {{- with .Values.affinity }}
affinity:
    {{- toYaml . | nindent 2 }}
  {{- end }}
{{- end }}
{{- if .Values.tolerations }}
  {{- with .Values.tolerations }}
tolerations:
    {{- toYaml . | nindent 2 }}
  {{- end }}
{{- end }}
{{- end -}}
