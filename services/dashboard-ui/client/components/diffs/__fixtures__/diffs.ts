export const deploymentBefore = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: acme-api
  labels:
    app: acme-api
spec:
  replicas: 2
  template:
    spec:
      containers:
        - name: api
          image: acme/api:1.4.2
          env:
            - name: LOG_LEVEL
              value: info
          resources:
            limits:
              memory: 512Mi
`

export const deploymentAfter = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: acme-api
  labels:
    app: acme-api
    tier: backend
spec:
  replicas: 4
  template:
    spec:
      containers:
        - name: api
          image: acme/api:1.5.0
          env:
            - name: LOG_LEVEL
              value: debug
            - name: FEATURE_FLAGS
              value: "queues,retries"
          resources:
            limits:
              memory: 1Gi
`

export const terraformBefore = `resource "aws_iam_role" "service" {
  name = "acme-service-role"
  path = "/"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}
`

export const terraformAfter = `resource "aws_iam_role" "service" {
  name                 = "acme-service-role"
  path                 = "/"
  max_session_duration = 7200

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = ["ec2.amazonaws.com", "ecs-tasks.amazonaws.com"] }
      Action    = "sts:AssumeRole"
    }]
  })
}
`

export const longManifest = (replicas: number) =>
  [
    'apiVersion: v1',
    'kind: ConfigMap',
    'metadata:',
    '  name: acme-settings',
    'data:',
    ...Array.from(
      { length: 120 },
      (_, index) => `  key_${index}: "value-${index}"`
    ),
    `  replicas: "${replicas}"`,
  ].join('\n')
