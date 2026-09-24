import {
  MODULES,
  presetModules,
  type TModuleId,
} from '../../../utils/modules'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { ModuleManager, type IModuleState } from './ModuleManager'

export default {
  title: 'lite/organisms/ModuleManager',
}

const states = (
  enabled: ReadonlySet<TModuleId>,
  pinned: readonly TModuleId[] = []
): IModuleState[] =>
  MODULES.map((module) => ({
    module,
    enabled: enabled.has(module.id) && !pinned.includes(module.id),
    pinned: pinned.includes(module.id),
    registered: true,
  }))

const noop = () => {}

export const Overview = () => (
  <ComponentDocs
    name="ModuleManager"
    tier="organism"
    summary="The internal panel where Nuon staff decide which modules an org's dashboard ships."
    use={[
      'Render it on the modules page through ModuleManagerContainer, which reads the org flags and writes them through the admin API.',
      'Use the preset select to apply a whole bundle, and the per-module switches to fine tune one org.',
      'Copy the deployment config block into a bespoke control plane to pin the same selection for every org there.',
    ]}
    avoid={[
      'Do not expose it to org users. It is gated to Nuon staff and changes what their customers see.',
      'Do not use it for feature flags that are not modules. The admin dashboard owns those.',
    ]}
    rules={[
      'The module list and presets come from the registry in utils/modules.ts, never from props.',
      'A pinned module renders read only, because forced_enabled_features wins over the stored org value.',
      'While a change is pending every switch locks and the affected cards show the saving state.',
      'Errors render as a banner above the cards, never as a toast.',
    ]}
    props={[
      {
        name: 'modules',
        type: 'IModuleState[]',
        description:
          'One entry per registry module with its enabled, pinned, and registered state.',
      },
      {
        name: 'onToggle',
        type: '(id: TModuleId, enabled: boolean) => void',
        description: 'Requests one module on or off.',
      },
      {
        name: 'onPreset',
        type: '(id: TModulePresetId) => void',
        description: 'Requests a whole preset.',
      },
      {
        name: 'pending',
        type: "TModuleId | 'preset'",
        description:
          'The change in flight. Locks every switch and marks the affected cards as saving.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'The org has not resolved yet, so counts and switches show skeletons.',
      },
      {
        name: 'error',
        type: 'unknown',
        description: 'The last failed save, rendered as an error banner.',
      },
    ]}
  />
)

export const FullPlatform = () => (
  <div className="max-w-4xl p-8">
    <ModuleManager
      modules={states(presetModules('full'))}
      onToggle={noop}
      onPreset={noop}
    />
  </div>
)

export const InstallOperations = () => (
  <div className="max-w-4xl p-8">
    <ModuleManager
      modules={states(presetModules('install-operations'))}
      onToggle={noop}
      onPreset={noop}
    />
  </div>
)

export const PinnedByDeployment = () => (
  <div className="max-w-4xl p-8">
    <ModuleManager
      modules={states(presetModules('full'), ['apps', 'oidc-federation'])}
      onToggle={noop}
      onPreset={noop}
    />
  </div>
)

export const Saving = () => (
  <div className="max-w-4xl p-8">
    <ModuleManager
      modules={states(presetModules('core'))}
      onToggle={noop}
      onPreset={noop}
      pending="installs"
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-4xl p-8">
    <ModuleManager
      modules={states(presetModules('full'))}
      onToggle={noop}
      onPreset={noop}
      loading
    />
  </div>
)

export const SaveFailed = () => (
  <div className="max-w-4xl p-8">
    <ModuleManager
      modules={states(presetModules('full'))}
      onToggle={noop}
      onPreset={noop}
      error={{ error: 'admin email is not authorized for this org' }}
    />
  </div>
)
