export default {
  title: 'Installs/Forms/ComponentOverridesSection',
}

import { ComponentOverridesSection } from './ComponentOverridesSection'
import type { TAppInputConfig } from '@/types'

type TGroup = NonNullable<TAppInputConfig['input_groups']>[number]

// hex-encoded component names (Go hex.EncodeToString of the component name)
const HEX = {
  certificate: '6365727469666963617465',
  api_gateway: '6170695f67617465776179',
  whoami: '77686f616d69',
}

const group: TGroup = {
  id: 'group-overrides',
  name: 'nuon_component_overrides',
  display_name: 'Component overrides',
  index: 3,
  app_inputs: [
    {
      id: 'i1',
      name: `nuon_component_override_v1_tf_vars_${HEX.certificate}`,
      display_name: 'certificate tf vars',
      type: 'hcl',
      required: false,
      default: 'domain = "app.example.com"\n',
      index: 0,
      source: 'vendor',
    },
    {
      id: 'i2',
      name: `nuon_component_override_v1_enabled_${HEX.certificate}`,
      display_name: 'certificate enabled',
      type: 'bool',
      required: false,
      default: 'true',
      index: 1,
      source: 'vendor',
    },
    {
      id: 'i3',
      name: `nuon_component_override_v1_tf_vars_${HEX.api_gateway}`,
      display_name: 'api_gateway tf vars',
      type: 'hcl',
      required: false,
      default: 'stage = "prod"\n',
      index: 2,
      source: 'vendor',
    },
    {
      id: 'i4',
      name: `nuon_component_override_v1_enabled_${HEX.api_gateway}`,
      display_name: 'api_gateway enabled',
      type: 'bool',
      required: false,
      default: 'true',
      index: 3,
      source: 'vendor',
    },
    {
      id: 'i5',
      name: `nuon_component_override_v1_helm_values_${HEX.whoami}`,
      display_name: 'whoami helm values',
      type: 'yaml',
      required: false,
      default: 'replicaCount: 2\n',
      index: 4,
      source: 'vendor',
    },
  ],
} as TGroup

export const Default = () => (
  <form className="max-w-2xl p-6 flex flex-col gap-6">
    <ComponentOverridesSection group={group} />
  </form>
)

// api_gateway starts disabled via a draft value; its config editor greys out.
export const WithDisabledComponent = () => (
  <form className="max-w-2xl p-6 flex flex-col gap-6">
    <ComponentOverridesSection
      group={group}
      draftValues={{
        [`inputs:nuon_component_override_v1_enabled_${HEX.api_gateway}`]: 'false',
      }}
    />
  </form>
)

// A component with only an enabled toggle and no config override.
const toggleOnlyGroup: TGroup = {
  id: 'group-overrides-toggle-only',
  name: 'nuon_component_overrides',
  display_name: 'Component overrides',
  index: 0,
  app_inputs: [
    {
      id: 't1',
      name: `nuon_component_override_v1_enabled_${HEX.whoami}`,
      display_name: 'whoami enabled',
      type: 'bool',
      required: false,
      default: 'false',
      index: 0,
      source: 'vendor',
    },
  ],
} as TGroup

export const ToggleOnly = () => (
  <form className="max-w-2xl p-6 flex flex-col gap-6">
    <ComponentOverridesSection group={toggleOnlyGroup} />
  </form>
)
