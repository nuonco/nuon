export default {
  title: 'Install Overrides/ComponentOverrideCard',
}

import { ComponentOverrideCard } from './ComponentOverrideCard'
import type { TComponentOverrideCard } from '@/utils/install-utils'
import type { TAppInput } from '@/types'

const hex = (s: string) =>
  Array.from(new TextEncoder().encode(s))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')

const inputName = (kind: string, component: string) =>
  `nuon_component_override_v1_${kind}_${hex(component)}`

const mkInput = (name: string, defaultValue = ''): TAppInput =>
  ({ id: name, name, default: defaultValue }) as unknown as TAppInput

const helmValues = `replicaCount: 3
image:
  repository: nginx
  tag: "1.27"
resources:
  limits:
    cpu: 500m`

const tfVars = `instance_type = "t3.large"
replicas      = 2`

const helmCard: TComponentOverrideCard<TAppInput> = {
  component: 'whoami',
  componentType: 'helm_chart',
  enabledInput: mkInput(inputName('enabled', 'whoami'), 'true'),
  configInput: mkInput(inputName('helm_values', 'whoami')),
  configKind: 'helm_values',
  index: 0,
}

const tfCard: TComponentOverrideCard<TAppInput> = {
  component: 'database',
  componentType: 'terraform_module',
  enabledInput: null,
  configInput: mkInput(inputName('tf_vars', 'database')),
  configKind: 'tf_vars',
  index: 1,
}

export const ReadOnlyEnabled = () => (
  <ComponentOverrideCard
    card={helmCard}
    readOnly
    values={{
      [helmCard.enabledInput!.name!]: 'true',
      [helmCard.configInput!.name!]: helmValues,
    }}
  />
)

export const ReadOnlyDisabled = () => (
  <ComponentOverrideCard
    card={helmCard}
    readOnly
    values={{
      [helmCard.enabledInput!.name!]: 'false',
      [helmCard.configInput!.name!]: helmValues,
    }}
  />
)

export const ReadOnlyNoOverrideSet = () => (
  <ComponentOverrideCard
    card={helmCard}
    readOnly
    values={{ [helmCard.enabledInput!.name!]: 'true' }}
  />
)

export const ReadOnlyTerraform = () => (
  <ComponentOverrideCard
    card={tfCard}
    readOnly
    values={{ [tfCard.configInput!.name!]: tfVars }}
  />
)

export const ReadOnlyConfigOnly = () => (
  <ComponentOverrideCard
    card={helmCard}
    readOnly
    showEnabled={false}
    values={{ [helmCard.configInput!.name!]: helmValues }}
  />
)

export const Editable = () => (
  <form className="max-w-2xl">
    <ComponentOverrideCard
      card={helmCard}
      values={{
        [helmCard.enabledInput!.name!]: 'true',
        [helmCard.configInput!.name!]: helmValues,
      }}
    />
  </form>
)
