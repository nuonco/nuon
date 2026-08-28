export default {
  title: 'Install Overrides/ComponentOverridesList',
}

import { ComponentOverridesList } from './ComponentOverridesList'
import type { TAppInput } from '@/types'

const hex = (s: string) =>
  Array.from(new TextEncoder().encode(s))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')

const inputName = (kind: string, component: string) =>
  `nuon_component_override_v1_${kind}_${hex(component)}`

const mkInput = (name: string, index: number): TAppInput =>
  ({ id: name, name, index }) as unknown as TAppInput

const inputs: TAppInput[] = [
  mkInput(inputName('enabled', 'whoami'), 0),
  mkInput(inputName('helm_values', 'whoami'), 1),
  mkInput(inputName('tf_vars', 'database'), 2),
  mkInput(inputName('helm_values', 'redis'), 3),
]

const values: Record<string, string> = {
  [inputName('enabled', 'whoami')]: 'true',
  [inputName('helm_values', 'whoami')]:
    'replicaCount: 3\nimage:\n  tag: "1.27"',
  [inputName('tf_vars', 'database')]: 'instance_type = "t3.large"',
}

export const Default = () => (
  <div className="max-w-2xl">
    <ComponentOverridesList inputs={inputs} values={values} />
  </div>
)

export const WithoutEnabledState = () => (
  <div className="max-w-2xl">
    <ComponentOverridesList
      inputs={inputs}
      values={values}
      showEnabled={false}
    />
  </div>
)

export const Empty = () => (
  <div className="max-w-2xl">
    <ComponentOverridesList inputs={[]} values={{}} />
  </div>
)
