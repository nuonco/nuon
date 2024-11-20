---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo-server
  namespace: {{ .Release.Namespace }}
  labels:
    foo: bar
spec:
  selector:
    matchLabels:
      foo: bar
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
        - name: echo-server
          image: "jmalloc/echo-server"
          ports:
            - name: http
              containerPort: {{ .Values.port }}
              protocol: TCP
