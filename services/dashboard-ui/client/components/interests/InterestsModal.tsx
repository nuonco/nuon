import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { RadioInput } from '@/components/common/form/RadioInput'
import { useSurfaces } from '@/hooks/use-surfaces'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import {
  CATEGORY_LABELS,
  RESOURCE_CATEGORIES,
  isCategoryOn,
  isResourceEmpty,
  setCategoryOn,
} from './categories'
import {
  ALL_RESOURCES,
  RESOURCE_LABELS,
  type Interests,
  type ResourceCfg,
  type ResourceKind,
} from './types'

type Mode = 'all' | 'specific'

// Resources map is the source of truth for the "Choose specific events" branch.
// We keep it on local state regardless of the current Mode so toggling
// 'All events' → 'Choose specific events' → 'All events' doesn't blow away
// whatever the user has already ticked.
type ResourcesMap = NonNullable<Interests['resources']>

const initialMode = (value: Interests): Mode =>
  value.all_events ? 'all' : 'specific'

const initialResources = (value: Interests): ResourcesMap => value.resources ?? {}

// Popup form for editing the Interests config. Opened from InterestsPicker via
// useSurfaces; commits the draft back through onSave only on Save. The Cancel
// button (provided by Modal) discards the draft and the parent form keeps the
// previous value.
//
// The flat checklist deliberately drops the per-resource outcome filter
// (succeeded vs failed vs started) and sub-op narrowing from the picker UI.
// Both are still expressible on the wire — power users can craft them by API
// — but they make the picker overwhelming for the 99% case.
export const InterestsModal = ({
  value,
  onSave,
  ...props
}: {
  value: Interests
  onSave: (next: Interests) => void
} & Omit<IModal, 'children' | 'primaryActionTrigger'>) => {
  const { removeModal } = useSurfaces()

  const [mode, setMode] = useState<Mode>(() => initialMode(value))
  const [resources, setResources] = useState<ResourcesMap>(() =>
    initialResources(value)
  )

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
    mode === 'all'
      ? 0
      : ALL_RESOURCES.reduce((sum, kind) => {
          const cfg: ResourceCfg | undefined = resources[kind]
          return (
            sum + RESOURCE_CATEGORIES[kind].filter((c) => isCategoryOn(cfg, c)).length
          )
        }, 0)

  const handleSave = () => {
    const next: Interests =
      mode === 'all' ? { all_events: true } : { resources }
    onSave(next)
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
        <div className="flex flex-col gap-1">
          <RadioInput
            id="interests-mode-all"
            name="interests-mode"
            checked={mode === 'all'}
            onChange={() => setMode('all')}
            labelProps={{
              labelText: 'All events',
              labelTextProps: { variant: 'body' },
            }}
          />
          <RadioInput
            id="interests-mode-specific"
            name="interests-mode"
            checked={mode === 'specific'}
            onChange={() => setMode('specific')}
            labelProps={{
              labelText: 'Choose specific events',
              labelTextProps: { variant: 'body' },
            }}
          />
        </div>

        {mode === 'specific' ? (
          <div className="flex flex-col gap-3 border-t border-neutral-200 pt-4 dark:border-neutral-700">
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
        ) : null}
      </div>
    </Modal>
  )
}
