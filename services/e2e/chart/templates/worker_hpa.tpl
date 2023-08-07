---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{ include "common.fullname" . }}-worker
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.workerLabels" . | nindent 4 }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ include "common.fullname" . }}-worker
  minReplicas: {{ .Values.worker.autoscaling.minReplicas | default 1 }}
  maxReplicas: {{ .Values.worker.autoscaling.maxReplicas | default 3 }}
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{ .Values.worker.autoscaling.targetCPUUtilizationPercentage }}
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: {{ .Values.worker.autoscaling.targetMemoryUtilizationPercentage }}
