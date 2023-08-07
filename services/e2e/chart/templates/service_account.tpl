---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ .Values.serviceAccount.name }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
  annotations:
    {{- toYaml .Values.serviceAccount.annotations | nindent 4 }}

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ .Values.serviceAccount.name }}-rolebinding
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
roleRef:
  kind: Role
  name: {{ .Values.serviceAccount.name }}
  apiGroup: rbac.authorization.k8s.io
subjects:
  - kind: ServiceAccount
    name: {{ .Values.serviceAccount.name }}
    namespace: {{ .Release.Namespace }}

---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: {{ .Values.serviceAccount.name }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "patch"]
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get"]
