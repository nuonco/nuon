---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
  annotations:
    alb.ingress.kubernetes.io/scheme: {{ .Values.ui.alb.scheme }}
    alb.ingress.kubernetes.io/target-type: ip
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTPS":443}]'
    alb.ingress.kubernetes.io/certificate-arn: {{ .Values.ui.alb.domain_certificate }}
    alb.ingress.kubernetes.io/aws-load-balancer-ssl-ports: https
    alb.ingress.kubernetes.io/healthcheck-path: /
    external-dns.alpha.kubernetes.io/hostname: {{ .Values.ui.alb.domain }}
spec:
  ingressClassName: alb
  rules:
    - http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: {{ include "common.fullname" . }}
                port:
                  name: http
---
apiVersion: v1
kind: Service
metadata:
  name: {{ include "common.fullname" . }}
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
  annotations:
    alb.ingress.kubernetes.io/target-type: 'ip'
spec:
  selector:
    {{- include "common.selectorLabels" . | nindent 4 }}
  type: ClusterIP
  ports:
    - name: http
      port: 8080
      targetPort: http-internal
