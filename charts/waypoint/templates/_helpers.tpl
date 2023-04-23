{{/*
Expand the name of the chart.
*/}}
{{- define "waypoint.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "waypoint.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "waypoint.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "waypoint.labels" -}}
helm.sh/chart: {{ include "waypoint.chart" . }}
{{ include "waypoint.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "waypoint.selectorLabels" -}}
app.kubernetes.io/name: {{ include "waypoint.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "waypoint.serviceAccount.name" -}}
{{- if .Values.server.serviceAccount.create -}}
    {{ default (include "waypoint.fullname" .) .Values.server.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.server.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Sets extra service account annotations
*/}}
{{- define "waypoint.serviceAccount.annotations" -}}
  {{- if .Values.server.serviceAccount.annotations }}
  annotations:
    {{- $tp := typeOf .Values.server.serviceAccount.annotations }}
    {{- if eq $tp "string" }}
      {{- tpl .Values.server.serviceAccount.annotations . | nindent 4 }}
    {{- else }}
      {{- toYaml .Values.server.serviceAccount.annotations | nindent 4 }}
    {{- end }}
  {{- end }}
{{- end -}}

{{/*
Sets extra pod annotations
*/}}
{{- define "waypoint.annotations" -}}
  {{- if .Values.server.annotations }}
      annotations:
        {{- $tp := typeOf .Values.server.annotations }}
        {{- if eq $tp "string" }}
          {{- tpl .Values.server.annotations . | nindent 8 }}
        {{- else }}
          {{- toYaml .Values.server.annotations | nindent 8 }}
        {{- end }}
  {{- end }}
{{- end -}}

{{/*
Sets extra statefulset annotations
*/}}
{{- define "waypoint.statefulSet.annotations" -}}
  {{- if .Values.server.statefulSet.annotations }}
  annotations:
    {{- $tp := typeOf .Values.server.statefulSet.annotations }}
    {{- if eq $tp "string" }}
      {{- tpl .Values.server.statefulSet.annotations . | nindent 4 }}
    {{- else }}
      {{- toYaml .Values.server.statefulSet.annotations | nindent 4 }}
    {{- end }}
  {{- end }}
{{- end -}}

{{/*
Sets the server token secret name.
*/}}
{{- define "waypoint.serverTokenSecret" -}}
{{- include "waypoint.fullname" . -}}-server-token
{{- end -}}

{{/*
Sets the runner token secret name.
*/}}
{{- define "waypoint.runnerTokenSecret" -}}
{{- include "waypoint.fullname" . -}}-runner-token
{{- end -}}

{{/*
Create the name of the static runner service account to use
*/}}
{{- define "waypoint.runner.serviceAccount.name" -}}
{{- if .Values.runner.serviceAccount.create -}}
    {{ default (printf "%s-runner" (include "waypoint.fullname" .)) .Values.runner.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.runner.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Sets extra static runner service account annotations
*/}}
{{- define "waypoint.runner.serviceAccount.annotations" -}}
  {{- if .Values.runner.serviceAccount.annotations }}
  annotations:
    {{- $tp := typeOf .Values.runner.serviceAccount.annotations }}
    {{- if eq $tp "string" }}
      {{- tpl .Values.runner.serviceAccount.annotations . | nindent 4 }}
    {{- else }}
      {{- toYaml .Values.runner.serviceAccount.annotations | nindent 4 }}
    {{- end }}
  {{- end }}
{{- end -}}

{{/*
Create the name of the odr runner service account to use
*/}}
{{- define "waypoint.runner.odr.serviceAccount.name" -}}
{{- if .Values.runner.odr.serviceAccount.create -}}
    {{ default (printf "%s-runner-odr" (include "waypoint.fullname" .)) .Values.runner.odr.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.runner.odr.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Sets extra odr runner service account annotations
*/}}
{{- define "waypoint.runner.odr.serviceAccount.annotations" -}}
  {{- if .Values.runner.odr.serviceAccount.annotations }}
  annotations:
    {{- $tp := typeOf .Values.runner.odr.serviceAccount.annotations }}
    {{- if eq $tp "string" }}
      {{- tpl .Values.runner.odr.serviceAccount.annotations . | nindent 4 }}
    {{- else }}
      {{- toYaml .Values.runner.odr.serviceAccount.annotations | nindent 4 }}
    {{- end }}
  {{- end }}
{{- end -}}

{{/*
Sets extra server Service annotations
*/}}
{{- define "waypoint.service.annotations" -}}
  {{- if .Values.server.service.annotations }}
    {{- $tp := typeOf .Values.server.service.annotations }}
    {{- if eq $tp "string" }}
      {{- tpl .Values.server.service.annotations . | nindent 4 }}
    {{- else }}
      {{- toYaml .Values.server.service.annotations | nindent 4 }}
    {{- end }}
  {{- end }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "waypoint.bootstrap.serviceAccount.name" -}}
{{- if .Values.bootstrap.serviceAccount.create -}}
    {{ default (printf "%s-bootstrap" (include "waypoint.fullname" .)) .Values.bootstrap.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.bootstrap.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Sets extra service account annotations
*/}}
{{- define "waypoint.bootstrap.serviceAccount.annotations" -}}
  {{- if .Values.bootstrap.serviceAccount.annotations }}
  annotations:
    {{- $tp := typeOf .Values.bootstrap.serviceAccount.annotations }}
    {{- if eq $tp "string" }}
      {{- tpl .Values.bootstrap.serviceAccount.annotations . | nindent 4 }}
    {{- else }}
      {{- toYaml .Values.bootstrap.serviceAccount.annotations | nindent 4 }}
    {{- end }}
  {{- end }}
{{- end -}}

{{/*
imagePullSecrets generates pull secrets from either string or map values.
A map value must be indexable by the key 'name'.
*/}}
{{- define "imagePullSecrets" -}}
{{- with .Values.global.imagePullSecret -}}
imagePullSecrets:
  - name: {{ . }}
{{- end -}}
{{- end -}}

{{/*
Sets extra ingress annotations
*/}}
{{- define "waypoint.ingress.annotations" -}}
  {{- if .Values.ui.ingress.annotations }}
    {{- $tp := typeOf .Values.ui.ingress.annotations }}
    {{- if eq $tp "string" }}
      {{- tpl .Values.ui.ingress.annotations . | nindent 4 }}
    {{- else }}
      {{- toYaml .Values.ui.ingress.annotations | nindent 4 }}
    {{- end }}
  {{- end }}
{{- end -}}
