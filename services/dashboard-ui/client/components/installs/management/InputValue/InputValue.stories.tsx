export default {
  title: 'Installs/InputValue',
}

import { InputValue } from './InputValue'

export const Scalar = () => <InputValue name="domain" value="nuon.run" />

export const Empty = () => <InputValue name="domain" value="" />

export const Missing = () => <InputValue name="domain" value={null} />

export const HelmValues = () => (
  <InputValue
    name="nuon_component_override_v1_helm_values_77686f616d69"
    value={
      'replicaCount: 5\nresources:\n  requests:\n    cpu: "150m"\n    memory: 64Mi\n'
    }
  />
)

export const TFVars = () => (
  <InputValue
    name="nuon_component_override_v1_tf_vars_6365727469666963617465"
    value={'domain_name = "whoami.example.com"\n'}
  />
)

const longHelmValues = `replicaCount: 3
image:
  repository: ghcr.io/acme/whoami
  tag: "1.24.0"
  pullPolicy: IfNotPresent
resources:
  requests:
    cpu: "150m"
    memory: 64Mi
  limits:
    cpu: "500m"
    memory: 256Mi
service:
  type: ClusterIP
  port: 80
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: whoami.example.com
      paths:
        - path: /
          pathType: Prefix
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 75
nodeSelector:
  kubernetes.io/os: linux
`

const longTFVars = `region             = "us-east-1"
cidr_block         = "10.1.0.0/16"
instance_count     = 3
instance_type      = "t3.large"
enable_nat_gateway = true
availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
private_subnets    = ["10.1.1.0/24", "10.1.2.0/24", "10.1.3.0/24"]
public_subnets     = ["10.1.101.0/24", "10.1.102.0/24", "10.1.103.0/24"]
tags = {
  environment = "production"
  team        = "platform"
  managed_by  = "nuon"
}
`

export const LongHelmValues = () => (
  <InputValue
    name="nuon_component_override_v1_helm_values_77686f616d69"
    value={longHelmValues}
  />
)

export const LongTFVars = () => (
  <InputValue
    name="nuon_component_override_v1_tf_vars_767063"
    value={longTFVars}
  />
)
