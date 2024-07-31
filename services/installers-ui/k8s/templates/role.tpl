---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: {{ .Values.serviceAccount.name }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
