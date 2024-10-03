---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
    {{- with .Values.controller.labels }}
      {{- toYaml . | nindent 4 }}
    {{- end }}
    app.kubernetes.io/name: {{ include "common.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
  {{- with .Values.controller.annotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  {{- if not .Values.autoscaling.enabled }}
  replicas: {{ .Values.replicaCount }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "common.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      {{- with .Values.podAnnotations }}
      annotations:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      labels:
        {{- include "common.selectorLabels" . | nindent 8 }}
    spec:
      serviceAccountName: {{ include "common.serviceAccountName" . }}
      containers:
        - name: {{ include "common.fullname" . }}
          image: {{ printf "%s:%v" .Values.image.repository (default .Chart.AppVersion .Values.image.tag) | quote }}
          env:
            - name: HOST_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.hostIP
            - name: HOST_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
          envFrom:
          - configMapRef:
              name: {{ include "common.fullname" . }}
          command: {{ toYaml .Values.command | nindent 14 }}
          imagePullPolicy: {{ .Values.image.imagePullPolicy | quote | default "IfNotPresent" }}
          resources:
            {{- toYaml .Values.resources | nindent 12 }}

