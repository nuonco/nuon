---
apiVersion: v1
kind: Service
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.apiLabels" . | nindent 4 }}
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: ip
    service.beta.kubernetes.io/aws-load-balancer-scheme: internal
    service.beta.kubernetes.io/aws-load-balancer-target-group-attributes: preserve_client_ip.enabled=false
    service.beta.kubernetes.io/aws-load-balancer-target-group-attributes: deregistration_delay.draining_interval=10
    service.beta.kubernetes.io/aws-load-balancer-target-group-attributes: deregistration_delay.connection_termination.enabled=true
    service.beta.kubernetes.io/aws-load-balancer-healthcheck-interval: "10"
    service.beta.kubernetes.io/aws-load-balancer-healthcheck-unhealthy-threshold: "2"
    external-dns.alpha.kubernetes.io/hostname: {{ .Values.api.domain }}
spec:
  type: LoadBalancer
  loadBalancerClass: service.k8s.aws/nlb
  allocateLoadBalancerNodePorts: false
  externalTrafficPolicy: Local
  internalTrafficPolicy: Local
  selector:
    {{- include "common.apiSelectorLabels" . | nindent 4 }}
    app.nuon.co/name: {{ include "common.fullname" . }}
  ports:
    - name: http
      port: 80
      targetPort: http
