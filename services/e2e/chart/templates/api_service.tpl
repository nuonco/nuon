---
apiVersion: v1
kind: Service
metadata:
  name: {{ include "common.fullname" . }}-api
  labels:
    {{- include "common.apiLabels" . | nindent 4 }}
  namespace: {{ .Release.Namespace }}
spec:
  clusterIP: None
  ports:
    - name: https
      port: {{ .Values.api.port }}
      targetPort: https
  selector:
    {{- include "common.apiSelectorLabels" . | nindent 4 }}
