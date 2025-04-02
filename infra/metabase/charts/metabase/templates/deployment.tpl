---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.apiLabels" . | nindent 4 }}
    app.nuon.co/name: {{ include "common.fullname" . }}
spec:
  selector:
    matchLabels:
      {{- include "common.apiSelectorLabels" . | nindent 6 }}
      app.nuon.co/name: {{ include "common.fullname" . }}
  template:
    metadata:
      labels:
        {{- include "common.apiSelectorLabels" . | nindent 8 }}
        app.nuon.co/name: {{ include "common.fullname" . }}
    spec:
      serviceAccountName: {{ .Values.serviceAccount.name }}
      automountServiceAccountToken: true

      containers:
        - name: {{ include "common.fullname" . }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          ports:
            - name: http
              containerPort: {{ .Values.api.port }}
              protocol: TCP
          readinessProbe:
            httpGet:
              path: {{ .Values.api.readiness_probe}}
              port: http
            failureThreshold: 10
            periodSeconds: 10
            timeoutSeconds: 1
          livenessProbe:
            httpGet:
              path: {{ .Values.api.liveness_probe}}
              port: http
            periodSeconds: 10
            timeoutSeconds: 1
          resources:
            limits:
              cpu: {{ .Values.api.resources.limits.cpu }}
              memory: {{ .Values.api.resources.limits.memory }}
            requests:
              cpu: {{ .Values.api.resources.requests.cpu }}
              memory: {{ .Values.api.resources.requests.memory }}
          envFrom:
            - configMapRef:
                name: {{ include "common.fullname" . }}
          env:
          {{- range $envSecret := .Values.envSecrets }}
            - name: {{ $envSecret.name }}
              valueFrom:
                secretKeyRef:
                  name: {{ $envSecret.valueFrom.name }}
                  key: {{ $envSecret.valueFrom.key }}
          {{- end}}
            - name: HOST_IP
              valueFrom:
                  fieldRef:
                      fieldPath: status.hostIP
            - name: HOST_NAME
              valueFrom:
                  fieldRef:
                      fieldPath: spec.nodeName
