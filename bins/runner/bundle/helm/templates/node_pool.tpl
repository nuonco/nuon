{{- if .Values.node_pool.enabled }}

apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  {{- /* https://karpenter.sh/docs/concepts/disruption/#consolidation */}}
  disruption:
    consolidateAfter: 120s
    consolidationPolicy: WhenEmpty
    budgets:
    - nodes: 50%
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
      {{- with randNumeric 3 }}
      expireAfter: {{ cat "2635" . "s" | replace " " "" | quote }}
      {{- end }}
      taints:
        - key: deployment
          effect: NoSchedule
          value: {{ include "common.fullname" . }}
      nodeClassRef:
        group: karpenter.k8s.aws
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
