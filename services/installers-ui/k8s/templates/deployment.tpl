---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  selector:
    matchLabels:
      {{- include "common.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "common.selectorLabels" . | nindent 8 }}
    spec:
      serviceAccountName: {{ .Values.serviceAccount.name }}
      automountServiceAccountToken: true
      containers:
        - name: {{ include "common.fullname" . }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          command:
            - npm
            - run
            - start
          ports:
            - name: http
              containerPort: {{ .Values.ui.port }}
              protocol: TCP
          readinessProbe:
            httpGet:
              path: {{ .Values.ui.readiness_probe}}
              port: http
          livenessProbe:
            httpGet:
              path: {{ .Values.ui.liveness_probe}}
              port: http
          resources:
            limits:
              cpu: {{ .Values.ui.resources.limits.cpu }}
              memory: {{ .Values.ui.resources.limits.memory }}
            requests:
              cpu: {{ .Values.ui.resources.requests.cpu }}
              memory: {{ .Values.ui.resources.requests.memory }}
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
