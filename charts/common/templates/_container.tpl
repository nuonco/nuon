{{- define "common.container" -}}
- name: {{ include "common.fullname" . }}
  image: {{ printf "%s:%v" .Values.image.repository (default .Chart.AppVersion .Values.image.tag) | quote }}
  env:
  {{- if .Values.envSecrets}}
  {{- range $envSecret := .Values.envSecrets }}
    - name: {{ $envSecret.name }}
      valueFrom:
        secretKeyRef:
          name: {{ $envSecret.valueFrom.name }}
          key: {{ $envSecret.valueFrom.key }}
  {{- end}}
  {{- end}}
    - name: HOST_IP
      valueFrom:
          fieldRef:
              fieldPath: status.hostIP
    - name: HOST_NAME
      valueFrom:
          fieldRef:
              fieldPath: spec.nodeName
  envFrom:
  {{- if .Values.configmap.enabled }}
  - configMapRef:
      name: {{ include "common.fullname" . }}
  {{- end }}
  {{- with .Values.command }}
  command:
    {{- if kindIs "string" . }}
    - {{ . }}
    {{- else }}
      {{ toYaml . | nindent 4 }}
    {{- end }}
  {{- end }}
  {{- with .Values.args }}
  args:
    {{- if kindIs "string" . }}
    - {{ . }}
    {{- else }}
    {{ toYaml . | nindent 4 }}
    {{- end }}
  {{- end }}
  imagePullPolicy: {{ .Values.image.imagePullPolicy | quote | default "IfNotPresent" }}
  {{- with .Values.lifecycle }}
  lifecycle:
    {{- toYaml . | nindent 4}}
  {{- end }}
  {{- with .Values.securityContext }}
  securityContext:
    {{- toYaml . | nindent 4 }}
  {{- end }}
   {{- with (include "common.volumeMounts" . | trim) }}
  volumeMounts:
    {{- nindent 4 . }}
  {{- end }}
  {{- include "common.probes" . | trim | nindent 2 }}
  {{- with .Values.resources }}
  resources:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end -}}


