
---
apiVersion: v1
kind: Service
metadata:
  name: echo-server
  namespace: {{ .Release.Namespace }}
  labels:
    foo: bar
spec:
  type: LoadBalancer
  loadBalancerClass: service.k8s.aws/nlb
  allocateLoadBalancerNodePorts: false
  externalTrafficPolicy: Local
  internalTrafficPolicy: Local
  selector:
      foo: bar
  ports:
    - name: http
      port: 80
      targetPort: http-internal
