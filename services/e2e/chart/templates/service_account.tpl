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
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{ .Values.serviceAccount.name }}-rolebinding
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
subjects:
  - kind: ServiceAccount
    name: {{ .Values.serviceAccount.name }}
    namespace: {{ .Release.Namespace }}
roleRef:
  kind: ClusterRole
  name: {{ .Values.serviceAccount.name }}
  apiGroup: rbac.authorization.k8s.io
