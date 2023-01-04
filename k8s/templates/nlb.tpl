---
# NOTE: we currently run version 2.4.5 of the load balancer controller
# https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.4/guide/service/nlb/
apiVersion: v1
kind: Service
metadata:
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-nlb-type: ip
    service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: ip
    service.beta.kubernetes.io/aws-load-balancer-healthcheck-path: {{.Values.nlb.healthcheck}}
    service.beta.kubernetes.io/aws-load-balancer-healthcheck-protocol: tcp
    external-dns.alpha.kubernetes.io/internal-hostname: {{.Values.nlb.hostname}}
  name: {{ include "common.fullname" . }}
spec:
  selector:
    app.kubernetes.io/name: {{ include "common.fullname" . }}
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer
  loadBalancerClass: service.k8s.aws/nlb
