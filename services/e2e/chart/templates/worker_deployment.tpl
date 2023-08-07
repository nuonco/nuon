---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "common.fullname" . }}-worker
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.workerLabels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      {{- include "common.workerSelectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "common.workerSelectorLabels" . | nindent 8 }}
    spec:
      serviceAccountName: {{ .Values.serviceAccount.name }}
      automountServiceAccountToken: true
      containers:
        - name: {{ include "common.fullname" . }}-worker
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          command:
            - /bin/worker
          resources:
            limits:
              cpu: 100m
              memory: 128Mi
            requests:
              cpu: 100m
              memory: 128Mi
          envFrom:
            - configMapRef:
                name: {{ include "common.fullname" . }}
