import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Input } from '@/components/common/form/Input'
import { useSurfaces } from '@/hooks/use-surfaces'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { MatchPicker } from '@/components/match/MatchPicker'
import { labelsToQueryString, parseLabelsQuery } from '@/components/match/parse'
import type { SubscriptionMatch, TargetKind } from '@/components/match/types'
import { cn } from '@/utils/classnames'
import {
  CATEGORY_LABELS,
  RESOURCE_CATEGORIES,
  isCategoryOn,
  isResourceEmpty,
  setCategoryOn,
} from './categories'
import { PRESETS, matchPreset, type PresetId } from './presets'
import {
  ALL_RESOURCES,
  RESOURCE_LABELS,
  type Interests,
  type ResourceCfg,
  type ResourceKind,
} from './types'

type ResourcesMap = NonNullable<Interests['resources']>

export interface PresetModalOutput {
  interests: Interests
  match: SubscriptionMatch | undefined
}

const initialResources = (value: Interests): ResourcesMap => value.resources ?? {}

const extractLabels = (
  match: SubscriptionMatch | undefined,
  kind: TargetKind
): { include: string; exclude: string } => {
  const tm = match?.[kind]
  return {
    include: labelsToQueryString(tm?.selector?.match_labels),
    exclude: labelsToQueryString(tm?.selector?.not_match_labels),
  }
}

const buildMatchFromLabels = (
  kind: TargetKind,
  includeRaw: string,
  excludeRaw: string
): SubscriptionMatch | undefined => {
  const include = parseLabelsQuery(includeRaw)
  const exclude = parseLabelsQuery(excludeRaw)
  if (Object.keys(include).length === 0 && Object.keys(exclude).length === 0) {
    return undefined
  }
  const selector: { match_labels?: typeof include; not_match_labels?: typeof exclude } = {}
  if (Object.keys(include).length > 0) selector.match_labels = include
  if (Object.keys(exclude).length > 0) selector.not_match_labels = exclude
  return { [kind]: { selector } }
}

export const InterestsModal = ({
  value,
  matchValue,
  onSave,
  ...props
}: {
  value: Interests
  matchValue?: SubscriptionMatch
  onSave: (output: PresetModalOutput) => void
} & Omit<IModal, 'children' | 'primaryActionTrigger'>) => {
  const { removeModal } = useSurfaces()

  const [presetId, setPresetId] = useState<PresetId>(() => matchPreset(value))
  const [resources, setResources] = useState<ResourcesMap>(() =>
    initialResources(value)
  )

  const selectedPreset = PRESETS.find((p) => p.id === presetId)
  const scopeKind = selectedPreset?.scopeKind

  const [includeLabels, setIncludeLabels] = useState<string>(() =>
    scopeKind ? extractLabels(matchValue, scopeKind).include : ''
  )
  const [excludeLabels, setExcludeLabels] = useState<string>(() =>
    scopeKind ? extractLabels(matchValue, scopeKind).exclude : ''
  )
  const [customMatch, setCustomMatch] = useState<SubscriptionMatch | undefined>(
    () => (presetId === 'custom' ? matchValue : undefined)
  )

  const handlePresetSelect = (id: PresetId) => {
    setPresetId(id)
    const preset = PRESETS.find((p) => p.id === id)
    if (preset?.scopeKind) {
      const labels = extractLabels(matchValue, preset.scopeKind)
      setIncludeLabels(labels.include)
      setExcludeLabels(labels.exclude)
    }
    if (id === 'custom') {
      setCustomMatch(matchValue)
    }
  }

  const toggleCategory = (
    kind: ResourceKind,
    cat: (typeof RESOURCE_CATEGORIES)[ResourceKind][number]
  ) => {
    const cfg = resources[kind]
    const next = setCategoryOn(cfg, cat, !isCategoryOn(cfg, cat))
    setResources((prev) => {
      const out: ResourcesMap = { ...prev }
      if (isResourceEmpty(kind, next)) {
        delete out[kind]
      } else {
        out[kind] = next
      }
      return out
    })
  }

  const totalSelected =
    presetId === 'custom'
      ? ALL_RESOURCES.reduce((sum, kind) => {
          const cfg: ResourceCfg | undefined = resources[kind]
          return (
            sum + RESOURCE_CATEGORIES[kind].filter((c) => isCategoryOn(cfg, c)).length
          )
        }, 0)
      : 0

  const handleSave = () => {
    const preset = PRESETS.find((p) => p.id === presetId)
    let interests: Interests
    let match: SubscriptionMatch | undefined

    if (presetId === 'custom') {
      interests = { resources }
      match = customMatch
    } else if (preset) {
      interests = preset.build()
      match = preset.scopeKind
        ? buildMatchFromLabels(preset.scopeKind, includeLabels, excludeLabels)
        : undefined
    } else {
      interests = { all_events: true }
      match = undefined
    }

    onSave({ interests, match })
    removeModal(props.modalId)
  }

  return (
    <Modal
      heading="Choose events"
      size="default"
      primaryActionTrigger={{
        children: 'Save',
        onClick: handleSave,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-2">
          {PRESETS.map((preset) => (
            <div key={preset.id} className="flex flex-col gap-2">
              <PresetCard
                preset={preset}
                selected={presetId === preset.id}
                onSelect={() => handlePresetSelect(preset.id)}
              />
              {presetId === preset.id && preset.scopeKind ? (
                <div className="ml-6 flex flex-col gap-3 rounded-lg border border-neutral-200 p-3 dark:border-neutral-700">
                  <Text variant="subtext" weight="strong">
                    Filter by {preset.scopeKind} labels (optional)
                  </Text>
                  <Input
                    labelProps={{ labelText: 'Include labels' }}
                    placeholder="env=prod, tier=critical"
                    value={includeLabels}
                    onChange={(e) => setIncludeLabels(e.target.value)}
                  />
                  <Input
                    labelProps={{ labelText: 'Exclude labels' }}
                    placeholder="env=stage"
                    value={excludeLabels}
                    onChange={(e) => setExcludeLabels(e.target.value)}
                  />
                  <Text variant="subtext" theme="neutral">
                    Leave blank to match all {preset.scopeKind}.
                  </Text>
                </div>
              ) : null}
            </div>
          ))}
        </div>

        {presetId === 'custom' ? (
          <div className="flex flex-col gap-4 border-t border-neutral-200 pt-4 dark:border-neutral-700">
            <div className="flex flex-col gap-2">
              <Text variant="body" weight="strong">
                Scope
              </Text>
              <MatchPicker
                value={customMatch}
                onChange={setCustomMatch}
              />
            </div>

            <div className="flex flex-col gap-3">
              <Text variant="body" weight="strong">
                Events
              </Text>
              <div className="flex flex-col">
                {ALL_RESOURCES.flatMap((kind) =>
                  RESOURCE_CATEGORIES[kind].map((cat) => {
                    const cfg = resources[kind]
                    const checked = isCategoryOn(cfg, cat)
                    const id = `interests-${kind}-${cat}`
                    return (
                      <CheckboxInput
                        key={id}
                        id={id}
                        checked={checked}
                        onChange={() => toggleCategory(kind, cat)}
                        labelProps={{
                          labelText: (
                            <span>
                              <Text variant="body" weight="strong">
                                {RESOURCE_LABELS[kind]}
                              </Text>
                              <Text variant="body" theme="neutral">
                                {' \u2014 '}
                                {CATEGORY_LABELS[cat]}
                              </Text>
                            </span>
                          ),
                        }}
                      />
                    )
                  })
                )}
              </div>

              {totalSelected === 0 ? (
                <Banner theme="warn">
                  <Text variant="subtext">
                    No events are selected. This subscription will not receive
                    any events.
                  </Text>
                </Banner>
              ) : null}
            </div>
          </div>
        ) : null}
      </div>
    </Modal>
  )
}

const PresetCard = ({
  preset,
  selected,
  onSelect,
}: {
  preset: (typeof PRESETS)[number]
  selected: boolean
  onSelect: () => void
}) => (
  <button
    type="button"
    onClick={onSelect}
    className={cn(
      'flex flex-col gap-1 rounded-lg border p-3 text-left transition-colors',
      selected
        ? 'border-blue-500 bg-blue-50 dark:border-blue-400 dark:bg-blue-950/30'
        : 'border-neutral-200 hover:border-neutral-300 dark:border-neutral-700 dark:hover:border-neutral-600'
    )}
  >
    <div className="flex items-center gap-2">
      <RadioDot selected={selected} />
      <Text variant="body" weight="strong">
        {preset.label}
      </Text>
      {preset.recommended ? (
        <span className="ml-auto flex items-center gap-1 rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/40 dark:text-blue-300">
          <Icon variant="SparkleIcon" size={12} />
          Recommended
        </span>
      ) : null}
    </div>
    <Text variant="subtext" theme="neutral" className="pl-6">
      {preset.description}
    </Text>
  </button>
)

const RadioDot = ({ selected }: { selected: boolean }) => (
  <span
    className={cn(
      'flex h-4 w-4 shrink-0 items-center justify-center rounded-full border-2',
      selected
        ? 'border-blue-500 dark:border-blue-400'
        : 'border-neutral-300 dark:border-neutral-600'
    )}
  >
    {selected ? (
      <span className="h-2 w-2 rounded-full bg-blue-500 dark:bg-blue-400" />
    ) : null}
  </span>
)
