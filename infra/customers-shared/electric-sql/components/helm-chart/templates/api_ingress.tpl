---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{ include "common.fullname" . }}-public
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.apiLabels" . | nindent 4 }}
  annotations:
    external-dns.alpha.kubernetes.io/hostname: {{ .Values.api.ingresses.public_domain }}
    kubernetes.io/ingress.class: nginx
spec:
  tls:
  - hosts:
    - {{ .Values.api.ingresses.public_domain }}
    secretName: sync-service-ingress-public-tls
  rules:
    - host: {{ .Values.api.ingresses.public_domain}}
      http:
        paths:
        - path: /
          pathType: Prefix
          backend:
            service:
              name: {{ include "common.fullname" . }}
              port:
                number: 80
