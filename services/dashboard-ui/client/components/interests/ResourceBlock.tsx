import { useRef, useState } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { RadioInput } from '@/components/common/form/RadioInput'
import { cn } from '@/utils/classnames'
import { RESOURCES_WITH_DRIFT_DETECTED, SUB_OPS, labelForSubOp } from './subops'
import {
  OUTCOME_LABELS,
  RESOURCE_DESCRIPTIONS,
  RESOURCE_LABELS,
  type Outcome,
  type ResourceCfg,
  type ResourceKind,
} from './types'

// Per-resource accordion row used inside InterestsPicker. Owns expand state
// locally; the outer picker owns the canonical Interests value and just
// receives diff callbacks.
//
// Progressive-disclosure layout:
//   - Body is a single "Which event categories?" checkbox group
//     (Lifecycle / Approvals / Drift detection).
//   - Drift only renders for components + sandboxes (mirrors the Go
//     SupportsDriftDetected helper). The matcher silently drops drift workflow
//     lifecycle events entirely, so this toggle is the *only* knob that
//     surfaces drift notifications.
//   - Lifecycle details (outcome radio + the "Narrow to specific operations"
//     sub-ops disclosure) only appear when Lifecycle is ticked.
//   - Webhook approvals reveal exposes both approval_requests +
//     approval_responses checkboxes; slack collapses both into the single
//     category checkbox and writes the same value to both booleans.
//
// Outcome semantics on the wire:
//   - outcome: 'none' = lifecycle category off (drift / approvals still flow if
//     enabled). We toggle Lifecycle off by writing 'none' and remember the
//     previously-selected outcome locally so re-enabling restores it.
export const ResourceBlock = ({
  variant,
  kind,
  enabled,
  cfg,
  onToggleEnabled,
  onChange,
  disabled,
}: {
  variant: 'slack' | 'webhook'
  kind: ResourceKind
  enabled: boolean
  cfg: ResourceCfg
  onToggleEnabled: (next: boolean) => void
  onChange: (next: ResourceCfg) => void
  disabled?: boolean
}) => {
  const [expanded, setExpanded] = useState(false)
  const [subOpsExpanded, setSubOpsExpanded] = useState(false)

  const ops = cfg.ops ?? []
  const outcome: Outcome = cfg.outcome ?? 'all'
  const approvalRequests = !!cfg.approval_requests
  const approvalResponses = !!cfg.approval_responses
  const approvalsOn = approvalRequests || approvalResponses
  const driftDetected = !!cfg.drift_detected
  const supportsDrift = RESOURCES_WITH_DRIFT_DETECTED.has(kind)

  const lifecycleOn = outcome !== 'none'

  // Remember the last non-'none' outcome so toggling Lifecycle off → on
  // restores what the user had picked instead of resetting to 'completion'.
  const lastLifecycleOutcomeRef = useRef<Exclude<Outcome, 'none'>>(
    outcome === 'none' ? 'completion' : outcome
  )
  if (outcome !== 'none') {
    lastLifecycleOutcomeRef.current = outcome
  }

  const buildSummary = (): { text: string; theme: 'neutral' | 'warn' } => {
    if (!enabled) return { text: 'off', theme: 'neutral' }

    const driftActive = supportsDrift && driftDetected
    const parts: string[] = []
    if (lifecycleOn) parts.push(OUTCOME_LABELS[outcome].toLowerCase())
    if (approvalsOn) parts.push('approvals')
    if (driftActive) parts.push('drift')

    if (parts.length === 0) {
      return { text: 'no events', theme: 'warn' }
    }
    if (parts.length === 1 && driftActive && !lifecycleOn && !approvalsOn) {
      return { text: 'drift only', theme: 'neutral' }
    }
    return { text: parts.join(' · '), theme: 'neutral' }
  }
  const summary = buildSummary()

  const setOps = (nextOps: string[]) => {
    onChange({ ...cfg, ops: nextOps })
  }

  const toggleOp = (op: string) => {
    if (ops.includes(op)) {
      setOps(ops.filter((o) => o !== op))
    } else {
      setOps([...ops, op])
    }
  }

  const setOutcome = (next: Outcome) => {
    onChange({ ...cfg, outcome: next })
  }

  const setApproval = (req: boolean, res: boolean) => {
    onChange({ ...cfg, approval_requests: req, approval_responses: res })
  }

  const setDriftDetected = (next: boolean) => {
    onChange({ ...cfg, drift_detected: next })
  }

  const toggleLifecycle = (next: boolean) => {
    if (next) {
      setOutcome(lastLifecycleOutcomeRef.current)
    } else {
      setOutcome('none')
    }
  }

  const toggleApprovals = (next: boolean) => {
    setApproval(next, next)
  }

  return (
    <div
      className={cn(
        'flex flex-col rounded-md border transition-colors',
        enabled
          ? 'border-neutral-200 dark:border-neutral-700'
          : 'border-dashed border-neutral-200 dark:border-neutral-700 opacity-70'
      )}
    >
      <div className="flex items-center gap-2 p-3">
        <CheckboxInput
          id={`interests-${kind}-enabled`}
          checked={enabled}
          onChange={(e) => onToggleEnabled(e.target.checked)}
          disabled={disabled}
          labelProps={{
            labelText: (
              <span className="flex flex-col gap-[2px]">
                <Text variant="body" weight="strong">
                  {RESOURCE_LABELS[kind]}
                </Text>
                <Text variant="subtext" theme="neutral">
                  {RESOURCE_DESCRIPTIONS[kind]}
                </Text>
              </span>
            ),
            labelTextProps: { variant: 'body' },
          }}
        />

        <Text
          variant="subtext"
          theme={summary.theme}
          className="ml-auto !mr-2 truncate"
        >
          {summary.text}
        </Text>

        <button
          type="button"
          aria-expanded={expanded}
          aria-controls={`interests-${kind}-body`}
          onClick={() => setExpanded((v) => !v)}
          disabled={!enabled || disabled}
          className={cn(
            'flex items-center justify-center rounded-md p-1 hover:bg-black/5 dark:hover:bg-white/5',
            (!enabled || disabled) && 'opacity-50 cursor-not-allowed'
          )}
        >
          <Icon variant={expanded ? 'CaretUpIcon' : 'CaretDownIcon'} />
        </button>
      </div>

      {enabled && expanded ? (
        <div
          id={`interests-${kind}-body`}
          className="flex flex-col gap-4 border-t border-neutral-200 px-3 py-3 dark:border-neutral-700"
        >
          <div className="flex flex-col gap-2">
            <Text variant="subtext" weight="strong">
              Which event categories?
            </Text>
            <div className="flex flex-col gap-1">
              <CheckboxInput
                id={`interests-${kind}-cat-lifecycle`}
                checked={lifecycleOn}
                onChange={(e) => toggleLifecycle(e.target.checked)}
                disabled={disabled}
                labelProps={{
                  labelText: 'Lifecycle events (deploys, teardowns)',
                  labelTextProps: { variant: 'subtext' },
                }}
              />
              <CheckboxInput
                id={`interests-${kind}-cat-approvals`}
                checked={approvalsOn}
                onChange={(e) => toggleApprovals(e.target.checked)}
                disabled={disabled}
                labelProps={{
                  labelText: 'Approval events',
                  labelTextProps: { variant: 'subtext' },
                }}
              />
              {supportsDrift ? (
                <CheckboxInput
                  id={`interests-${kind}-cat-drift`}
                  checked={driftDetected}
                  onChange={(e) => setDriftDetected(e.target.checked)}
                  disabled={disabled}
                  labelProps={{
                    labelText: 'Drift detection',
                    labelTextProps: { variant: 'subtext' },
                  }}
                />
              ) : null}
            </div>
          </div>

          {lifecycleOn ? (
            <div className="flex flex-col gap-2 rounded-md border border-neutral-200 bg-neutral-50 p-3 dark:border-neutral-700 dark:bg-neutral-800/40">
              <Text variant="subtext" weight="strong">
                Lifecycle filter
              </Text>
              <div className="flex flex-wrap gap-x-2">
                {(['all', 'completion', 'failures'] as Exclude<
                  Outcome,
                  'none'
                >[]).map((o) => (
                  <RadioInput
                    key={o}
                    name={`interests-${kind}-outcome`}
                    value={o}
                    checked={outcome === o}
                    onChange={() => setOutcome(o)}
                    disabled={disabled}
                    labelProps={{
                      labelText: OUTCOME_LABELS[o],
                      labelTextProps: { variant: 'subtext' },
                    }}
                  />
                ))}
              </div>

              <button
                type="button"
                aria-expanded={subOpsExpanded}
                aria-controls={`interests-${kind}-subops`}
                onClick={() => setSubOpsExpanded((v) => !v)}
                disabled={disabled}
                className={cn(
                  'mt-1 flex items-center gap-1 self-start rounded-md text-left',
                  'hover:underline',
                  disabled && 'opacity-50 cursor-not-allowed'
                )}
              >
                <Icon variant={subOpsExpanded ? 'CaretDownIcon' : 'CaretRightIcon'} />
                <Text variant="subtext" theme="neutral">
                  Narrow to specific operations (optional)
                </Text>
              </button>

              {subOpsExpanded ? (
                <div
                  id={`interests-${kind}-subops`}
                  className="flex flex-col gap-1"
                >
                  <Text variant="subtext" theme="neutral">
                    Leave all unchecked to receive every sub-operation.
                  </Text>
                  <div className="flex flex-wrap gap-x-3">
                    {SUB_OPS[kind].map((op) => (
                      <CheckboxInput
                        key={op}
                        id={`interests-${kind}-op-${op}`}
                        checked={ops.includes(op)}
                        onChange={() => toggleOp(op)}
                        disabled={disabled}
                        labelProps={{
                          labelText: labelForSubOp(op),
                          labelTextProps: { variant: 'subtext' },
                        }}
                      />
                    ))}
                  </div>
                </div>
              ) : null}
            </div>
          ) : null}

          {variant === 'webhook' && approvalsOn ? (
            <div className="flex flex-col gap-2 rounded-md border border-neutral-200 bg-neutral-50 p-3 dark:border-neutral-700 dark:bg-neutral-800/40">
              <Text variant="subtext" weight="strong">
                Approval filter
              </Text>
              <div className="flex flex-col gap-1">
                <CheckboxInput
                  id={`interests-${kind}-approval-req`}
                  checked={approvalRequests}
                  onChange={(e) =>
                    setApproval(e.target.checked, approvalResponses)
                  }
                  disabled={disabled}
                  labelProps={{
                    labelText: 'Include approval requests',
                    labelTextProps: { variant: 'subtext' },
                  }}
                />
                <CheckboxInput
                  id={`interests-${kind}-approval-res`}
                  checked={approvalResponses}
                  onChange={(e) =>
                    setApproval(approvalRequests, e.target.checked)
                  }
                  disabled={disabled}
                  labelProps={{
                    labelText: 'Include approval responses',
                    labelTextProps: { variant: 'subtext' },
                  }}
                />
              </div>
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  )
}
