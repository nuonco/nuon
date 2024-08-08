{{- if .Values.node_pool.enabled }}

apiVersion: karpenter.sh/v1beta1
kind: NodePool
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace }}
  resourceVersion: "1628066"
  uid: 5f5f145c-1623-4167-a132-3034e34c4fad
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  disruption:
    budgets:
    - nodes: 50%
    consolidateAfter: 30s
    consolidationPolicy: WhenEmpty
    expireAfter: 50296s
  limits:
    cpu: {{ mul .Values.node_pool.instance_type.cpu .Values.node_pool.runner_count | add .Values.node_pool.instance_type.cpu }}
    {{- with mul .Values.node_pool.instance_type.memory .Values.node_pool.runner_count | add .Values.node_pool.instance_type.memory }}
    memory: {{ cat . "Mi" | replace " " "" | quote }}
    {{- end }}
  template:
    metadata:
      labels:
        {{ include "common.fullname" . }}: "true"
    spec:
      taints:
        - key: deployment
          effect: NoSchedule
          value: {{ include "common.fullname" . }}
      nodeClassRef:
        apiVersion: karpenter.k8s.aws/v1beta1
        kind: EC2NodeClass
        name: default
      requirements:
      - key: karpenter.sh/capacity-type
        operator: In
        values: {{ .Values.node_pool.capacity_types | toYaml | nindent 8 }}
      - key: node.kubernetes.io/instance-type
        operator: In
        values:
          - {{ .Values.node_pool.instance_type.name }}

{{- end }}
