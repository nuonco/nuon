export default {
  title: 'Installs/Configuration/InstallConfigOverrides',
}

import { Button } from '@/components/common/Button'
import type { TAppInput } from '@/types'
import { InstallConfigOverrides } from './InstallConfigOverrides'

const hex = (value: string) =>
  Array.from(new TextEncoder().encode(value))
    .map((byte) => byte.toString(16).padStart(2, '0'))
    .join('')

const overrideInput = (kind: string, component: string, index: number) =>
  ({
    id: `${kind}-${component}`,
    name: `nuon_component_override_v1_${kind}_${hex(component)}`,
    index,
  }) as TAppInput

const longHelmValues = `replicaCount: 4
image:
  repository: ghcr.io/acme/checkout
  tag: "2.14.0"
resources:
  requests:
    cpu: "250m"
    memory: 128Mi
  limits:
    cpu: "1"
    memory: 512Mi
ingress:
  enabled: true
  hosts:
    - host: checkout.example.com
      paths:
        - path: /
          pathType: Prefix
autoscaling:
  enabled: true
  minReplicas: 4
  maxReplicas: 12
`

const componentTypes = {
  checkout: 'helm_chart',
  database: 'terraform_module',
  search: 'helm_chart',
  worker: 'job',
  gateway: 'kubernetes_manifest',
} as const

const editInputs = <Button variant="secondary">Edit inputs</Button>

export const Default = () => (
  <InstallConfigOverrides
    action={editInputs}
    inputs={[
      overrideInput('enabled', 'checkout', 0),
      overrideInput('helm_values', 'checkout', 1),
    ]}
    values={{
      [overrideInput('enabled', 'checkout', 0).name!]: 'true',
      [overrideInput('helm_values', 'checkout', 1).name!]:
        'replicaCount: 4\nimage:\n  tag: "2.14.0"',
    }}
  />
)

export const MixedKinds = () => (
  <InstallConfigOverrides
    action={editInputs}
    componentTypes={componentTypes}
    inputs={[
      overrideInput('enabled', 'checkout', 0),
      overrideInput('helm_values', 'checkout', 1),
      overrideInput('tf_vars', 'database', 2),
      overrideInput('enabled', 'worker', 3),
      overrideInput('helm_values', 'search', 4),
    ]}
    values={{
      [overrideInput('enabled', 'checkout', 0).name!]: 'true',
      [overrideInput('helm_values', 'checkout', 1).name!]: longHelmValues,
      [overrideInput('tf_vars', 'database', 2).name!]:
        'instance_class = "db.r6g.large"\nmulti_az       = true\n',
      [overrideInput('enabled', 'worker', 3).name!]: 'false',
    }}
  />
)

export const DisabledComponent = () => (
  <InstallConfigOverrides
    action={editInputs}
    componentTypes={componentTypes}
    inputs={[
      overrideInput('enabled', 'search', 0),
      overrideInput('helm_values', 'search', 1),
    ]}
    values={{
      [overrideInput('enabled', 'search', 0).name!]: 'false',
      [overrideInput('helm_values', 'search', 1).name!]: longHelmValues,
    }}
  />
)

export const EnabledOnlyComponents = () => (
  <InstallConfigOverrides
    action={editInputs}
    componentTypes={componentTypes}
    inputs={[
      overrideInput('enabled', 'worker', 0),
      overrideInput('enabled', 'gateway', 1),
    ]}
    values={{
      [overrideInput('enabled', 'worker', 0).name!]: 'true',
      [overrideInput('enabled', 'gateway', 1).name!]: 'false',
    }}
  />
)

export const UnknownComponentTypes = () => (
  <InstallConfigOverrides
    action={editInputs}
    inputs={[
      overrideInput('enabled', 'worker', 0),
      overrideInput('helm_values', 'checkout', 1),
    ]}
    values={{
      [overrideInput('enabled', 'worker', 0).name!]: 'true',
      [overrideInput('helm_values', 'checkout', 1).name!]: longHelmValues,
    }}
  />
)

export const Loading = () => <InstallConfigOverrides inputs={[]} isLoading />

export const Empty = () => <InstallConfigOverrides inputs={[]} values={{}} />
