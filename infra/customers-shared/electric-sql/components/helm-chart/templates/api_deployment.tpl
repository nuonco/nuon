---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "common.fullname" . }}-api
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.apiLabels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      {{- include "common.apiSelectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "common.apiSelectorLabels" . | nindent 8 }}
    spec:
      serviceAccountName: {{ .Values.serviceAccount.name }}
      automountServiceAccountToken: true
      containers:
        - name: {{ include "common.fullname" . }}-api
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          command:
            - /bin/api
          ports:
            - name: http
              containerPort: {{ .Values.api.port }}
              protocol: TCP
          readinessProbe: null
          livenessProbe: null
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
