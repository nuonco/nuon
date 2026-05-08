import { useMemo } from 'react'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { Toggle } from '@/components/common/form/Toggle'
import { ResourceBlock } from './ResourceBlock'
import { defaultInterests } from './defaults'
import { RESOURCES_WITH_DRIFT_DETECTED } from './subops'
import {
  ALL_RESOURCES,
  type Interests,
  type ResourceCfg,
  type ResourceKind,
} from './types'

// resourceSupportsDrift mirrors the Go SupportsDriftDetected helper — only
// components and sandboxes can produce a drift_detected event, so the toggle
// is only meaningful (and only rendered) for those resources.
const resourceSupportsDrift = (kind: ResourceKind): boolean =>
  RESOURCES_WITH_DRIFT_DETECTED.has(kind)

// Shared picker that reads + writes the canonical Interests JSON shape (mirror
// of services/ctl-api/internal/pkg/interests/types.go).
//
// Variant differences:
//   - slack:   collapses approval_requests + approval_responses into a single
//              "Approval events" category checkbox. We always store both
//              booleans split internally and coerce both to the same value on
//              write. (Choice: simplest; no indeterminate visual. The webhook
//              variant remains the source of truth for the split shape.)
//   - webhook: ticking Approvals reveals a sub-row with the two booleans as
//              separate checkboxes.
//
// The picker NEVER auto-reverts an empty resources map. Toggling all resources
// off is a valid wire shape (`{}` — backend matcher returns false for every
// event); we surface that with an inline warning so users notice they will
// receive nothing.
export const InterestsPicker = ({
  variant,
  value,
  onChange,
  disabled,
}: {
  variant: 'slack' | 'webhook'
  value: Interests
  onChange: (next: Interests) => void
  disabled?: boolean
}) => {
  const allEventsOn = !!value.all_events
  const resources = value.resources ?? {}

  const enabledKinds = useMemo(
    () =>
      ALL_RESOURCES.filter((k) =>
        Object.prototype.hasOwnProperty.call(resources, k)
      ),
    [resources]
  )

  const allOff = !allEventsOn && enabledKinds.length === 0

  const handleAllEventsToggle = (next: boolean) => {
    if (next) {
      onChange({ all_events: true })
    } else {
      onChange(defaultInterests())
    }
  }

  const setResource = (kind: ResourceKind, cfg: ResourceCfg | undefined) => {
    const nextResources = { ...resources }
    if (cfg === undefined) {
      delete nextResources[kind]
    } else {
      nextResources[kind] = cfg
    }
    onChange({ resources: nextResources })
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-2 rounded-md border border-neutral-200 p-3 dark:border-neutral-700">
        <Toggle
          checked={allEventsOn}
          onChange={handleAllEventsToggle}
          disabled={disabled}
          label="Send all events"
          description="Every workflow, deploy, sandbox, runner, and approval. Toggle off to pick which resources and outcomes to receive."
        />
      </div>

      {!allEventsOn ? (
        <div className="flex flex-col gap-3">
          <div className="flex flex-col gap-2">
            {ALL_RESOURCES.map((kind) => {
              const enabled = Object.prototype.hasOwnProperty.call(
                resources,
                kind
              )
              const cfg = resources[kind] ?? {}
              return (
                <ResourceBlock
                  key={kind}
                  variant={variant}
                  kind={kind}
                  enabled={enabled}
                  cfg={cfg}
                  onToggleEnabled={(next) =>
                    setResource(
                      kind,
                      next
                        ? {
                            outcome: 'completion',
                            approval_requests: true,
                            approval_responses: true,
                            ...(resourceSupportsDrift(kind)
                              ? { drift_detected: true }
                              : {}),
                          }
                        : undefined
                    )
                  }
                  onChange={(nextCfg) => setResource(kind, nextCfg)}
                  disabled={disabled}
                />
              )
            })}
          </div>

          {allOff ? (
            <Banner theme="warn">
              <Text variant="subtext">
                No resources are enabled. This subscription will not receive any
                events.
              </Text>
            </Banner>
          ) : null}
        </div>
      ) : null}
    </div>
  )
}
