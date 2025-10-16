---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  replicas: {{- if .Values.node_pool.enabled }} {{ .Values.node_pool.runner_count }} {{- else }} 1 {{- end }}
  selector:
    matchLabels:
      {{- include "common.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "common.selectorLabels" . | nindent 8 }}
    spec:
      {{- if .Values.node_pool.enabled }}
      nodeSelector:
        {{ include "common.fullname" . }}: "true"  # Matches the label in the NodePool template
      tolerations:
        - key: "deployment"
          operator: "Equal"
          value: {{ include "common.fullname" . }}
          effect: "NoSchedule"
      {{- end }}

      serviceAccountName: {{ .Values.serviceAccount.name | default (include "common.fullname" .)}}
      automountServiceAccountToken: true
      containers:
        - name: {{ include "common.fullname" . }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: Always
          command:
            - /bin/runner
            - org
          envFrom:
            - configMapRef:
                name: {{ include "common.fullname" . }}
          env:
            - name: HOST_IP
              valueFrom:
                  fieldRef:
                      fieldPath: status.hostIP
            - name: HOST_NAME
              valueFrom:
                  fieldRef:
                      fieldPath: spec.nodeName
