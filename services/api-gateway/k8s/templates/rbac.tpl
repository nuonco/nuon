---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Values.namespace | default "default" }}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: view
subjects:
- kind: ServiceAccount
  name: {{ include "common.fullname" . }}
  namespace: {{ .Values.namespace | default "default" }}
