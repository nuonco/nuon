---
apiVersion: v1
kind: Service
metadata:
  name: {{ include "common.fullname" . }}-nlb-public
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.apiLabels" . | nindent 4 }}

  # NOTE: we run the 2.5.* version, please reference annotations here
  # https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.5/guide/service/annotations/
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: ip
    service.beta.kubernetes.io/aws-load-balancer-scheme: internet-facing
    service.beta.kubernetes.io/aws-load-balancer-target-group-attributes: preserve_client_ip.enabled=false
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert: {{ .Values.api.nlbs.public_domain_certificate_arn }}
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: https
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: tcp
    external-dns.alpha.kubernetes.io/hostname: {{ .Values.api.nlbs.public_domain }}
spec:
  type: LoadBalancer
  loadBalancerClass: service.k8s.aws/nlb
  allocateLoadBalancerNodePorts: false
  externalTrafficPolicy: Local
  internalTrafficPolicy: Local
  selector:
    {{- include "common.apiSelectorLabels" . | nindent 4 }}
  ports:
    - name: https
      port: 443
      targetPort: http

---
apiVersion: v1
kind: Service
metadata:
  name: {{ include "common.fullname" . }}-nlb-internal
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.apiLabels" . | nindent 4 }}
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: ip
    service.beta.kubernetes.io/aws-load-balancer-scheme: internal
    service.beta.kubernetes.io/aws-load-balancer-target-group-attributes: preserve_client_ip.enabled=false
    external-dns.alpha.kubernetes.io/hostname: {{ .Values.api.nlbs.internal_domain }}
spec:
  type: LoadBalancer
  loadBalancerClass: service.k8s.aws/nlb
  allocateLoadBalancerNodePorts: false
  externalTrafficPolicy: Local
  internalTrafficPolicy: Local
  selector:
    {{- include "common.apiSelectorLabels" . | nindent 4 }}
  ports:
    - name: http
      port: 443
      targetPort: http
