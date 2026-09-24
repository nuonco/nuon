import type { TAPIError } from '@/types'
import {
  MODULE_AREAS,
  MODULE_AREA_LABELS,
  MODULE_IDS,
  MODULE_PRESETS,
  deploymentConfig,
  matchPreset,
  presetById,
  type IModule,
  type TModuleId,
  type TModulePresetId,
} from '../../../utils/modules'
import { Banner } from '../../atoms/Banner'
import { Text } from '../../atoms/Text'
import { CodeBlock } from '../../molecules/CodeBlock'
import { Field } from '../../molecules/Field'
import { ModuleCard } from '../../molecules/ModuleCard'
import { OverviewCard, OverviewCardGrid } from '../../molecules/OverviewCard'
import { Select, type ISelectOption } from '../../molecules/Select'

export interface IModuleState {
  module: IModule
  enabled: boolean
  pinned: boolean
  registered: boolean
}

export type TModulePending = TModuleId | 'preset'

export interface IModuleManager {
  modules: IModuleState[]
  onToggle: (id: TModuleId, enabled: boolean) => void
  onPreset: (id: TModulePresetId) => void
  pending?: TModulePending
  loading?: boolean
  error?: unknown
}

const PRESET_OPTIONS: ISelectOption[] = [
  ...MODULE_PRESETS.map((preset) => ({
    value: preset.id,
    label: preset.name,
    description: preset.description,
  })),
  {
    value: 'custom',
    label: 'Custom',
    description: 'A selection that matches no preset.',
    disabled: true,
  },
]

const errorMessage = (error: unknown) =>
  (error as TAPIError | undefined)?.error ||
  'Unable to save the change. This is usually temporary.'

const hiddenSummary = (hidden: number) =>
  hidden === 0
    ? 'Every module ships.'
    : `${hidden} module${hidden === 1 ? '' : 's'} hidden from this org.`

const pinnedSummary = (pinned: number) =>
  pinned === 0
    ? 'No modules are pinned by this control plane.'
    : 'Set in the control plane config. They cannot change here.'

export const ModuleManager = ({
  modules,
  onToggle,
  onPreset,
  pending,
  loading = false,
  error,
}: IModuleManager) => {
  const enabled: ReadonlySet<TModuleId> = new Set(
    modules.filter((state) => state.enabled).map((state) => state.module.id)
  )
  const preset = matchPreset(enabled)
  const presetInfo = preset === 'custom' ? undefined : presetById(preset)
  const pinnedCount = modules.filter((state) => state.pinned).length
  const busy = pending !== undefined

  return (
    <div className="flex w-full flex-col gap-8">
      <Banner theme="info" heading="Modules shape what this org sees">
        Hiding a module removes its pages, navigation, and keyboard shortcuts
        for everyone in the org. The API and CLI are not affected.
      </Banner>
      {error ? (
        <Banner theme="error" heading="Module update failed">
          {errorMessage(error)}
        </Banner>
      ) : null}

      <OverviewCardGrid columns={3}>
        <OverviewCard title="Enabled">
          <Text
            variant="title"
            loading={loading}
            loadingWidth={6}
            className="tabular-nums"
          >
            {`${enabled.size} of ${MODULE_IDS.length}`}
          </Text>
          <Text variant="caption" color="tertiary">
            {hiddenSummary(MODULE_IDS.length - enabled.size)}
          </Text>
        </OverviewCard>
        <OverviewCard title="Preset">
          <Text variant="heading" loading={loading} loadingWidth={12} lines={1}>
            {presetInfo?.name ?? 'Custom'}
          </Text>
          <Text variant="caption" color="tertiary">
            {presetInfo?.description ?? 'This selection matches no preset.'}
          </Text>
        </OverviewCard>
        <OverviewCard title="Pinned">
          <Text
            variant="title"
            loading={loading}
            loadingWidth={2}
            className="tabular-nums"
          >
            {pinnedCount}
          </Text>
          <Text variant="caption" color="tertiary">
            {pinnedSummary(pinnedCount)}
          </Text>
        </OverviewCard>
      </OverviewCardGrid>

      <Field
        label="Preset"
        description="Apply a bundle of modules in one change. Changing a single module afterwards makes the selection custom."
        className="max-w-md"
      >
        <Select
          value={preset}
          options={PRESET_OPTIONS}
          loading={loading}
          disabled={busy}
          onChange={(value) => {
            if (value !== 'custom') onPreset(value as TModulePresetId)
          }}
        />
      </Field>

      {MODULE_AREAS.map((area) => {
        const states = modules.filter((state) => state.module.area === area)
        if (!states.length) return null

        return (
          <section key={area} className="flex flex-col gap-3">
            <Text as="h2" variant="heading">
              {MODULE_AREA_LABELS[area]}
            </Text>
            <ul className="flex flex-col gap-2">
              {states.map((state) => (
                <li key={state.module.id}>
                  <ModuleCard
                    module={state.module}
                    enabled={state.enabled}
                    pinned={state.pinned}
                    registered={state.registered}
                    loading={loading}
                    saving={pending === state.module.id || pending === 'preset'}
                    onToggle={
                      busy
                        ? undefined
                        : (next) => onToggle(state.module.id, next)
                    }
                  />
                </li>
              ))}
            </ul>
          </section>
        )
      })}

      <section className="flex flex-col gap-3">
        <Text as="h2" variant="heading">
          Deployment config
        </Text>
        <Text as="p" variant="caption" color="secondary">
          Pin this selection into a bespoke control plane by adding these flags
          to the ctl-api forced_enabled_features value. A pinned module stays
          hidden for every org in that deployment.
        </Text>
        <CodeBlock value={deploymentConfig(enabled)} language="yaml" copy />
      </section>
    </div>
  )
}
