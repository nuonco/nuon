import { MODULES, moduleById } from '../../utils/modules'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { ModuleCard } from './ModuleCard'

export default {
  title: 'lite/molecules/ModuleCard',
}

const APPS = moduleById('apps')!
const WEBHOOKS = moduleById('webhooks')!

export const Overview = () => (
  <ComponentDocs
    name="ModuleCard"
    tier="molecule"
    summary="One dashboard module as a row: what it is, whether it ships for this org, and the switch that changes that."
    use={[
      'List every module on the modules page, grouped by area.',
      'Show a pinned module so staff can see the deployment config decided it, not the org.',
      'Pass loading while the org resolves so the switch shows a skeleton instead of a wrong value.',
    ]}
    avoid={[
      'Do not use it for feature flags that are not modules. Those stay in the admin dashboard.',
      'Do not use it for resource state. A module is configuration, so it carries a Switch, never a Status.',
    ]}
    rules={[
      'The name, description, and icon come from the module registry and never from props.',
      'A pinned, unregistered, or callback-less card renders its switch disabled.',
      'Readiness renders as a Badge only when the module is not fully built.',
    ]}
    props={[
      {
        name: 'module',
        type: 'IModule',
        description: 'Registry entry that names, describes, and icons the row.',
      },
      {
        name: 'enabled',
        type: 'boolean',
        description: 'Whether the module ships for this org.',
      },
      {
        name: 'pinned',
        type: 'boolean',
        default: 'false',
        description:
          'The deployment config forces this module off. Shows a badge and disables the switch.',
      },
      {
        name: 'registered',
        type: 'boolean',
        default: 'true',
        description:
          'The API knows the module flag. False shows a badge and disables the switch.',
      },
      {
        name: 'saving',
        type: 'boolean',
        default: 'false',
        description: 'A change to this module is in flight.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'The enabled value has not resolved yet.',
      },
      {
        name: 'onToggle',
        type: '(enabled: boolean) => void',
        description:
          'Called with the requested value. Omit it to render the switch read only.',
      },
      {
        name: 'className',
        type: 'string',
        description: 'Extra classes for the card.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="max-w-2xl p-8">
    <ModuleCard module={APPS} enabled onToggle={() => {}} />
  </div>
)

export const Hidden = () => (
  <div className="max-w-2xl p-8">
    <ModuleCard module={WEBHOOKS} enabled={false} onToggle={() => {}} />
  </div>
)

export const Pinned = () => (
  <div className="max-w-2xl p-8">
    <ModuleCard module={APPS} enabled={false} pinned onToggle={() => {}} />
  </div>
)

export const Unregistered = () => (
  <div className="max-w-2xl p-8">
    <ModuleCard
      module={WEBHOOKS}
      enabled
      registered={false}
      onToggle={() => {}}
    />
  </div>
)

export const Saving = () => (
  <div className="max-w-2xl p-8">
    <ModuleCard module={APPS} enabled saving onToggle={() => {}} />
  </div>
)

export const Loading = () => (
  <div className="max-w-2xl p-8">
    <ModuleCard module={APPS} enabled loading onToggle={() => {}} />
  </div>
)

export const Registry = () => (
  <ul className="flex max-w-2xl flex-col gap-2 p-8">
    {MODULES.map((module) => (
      <li key={module.id}>
        <ModuleCard module={module} enabled onToggle={() => {}} />
      </li>
    ))}
  </ul>
)
