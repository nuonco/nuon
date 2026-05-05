import { useState } from 'react'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { RadioInput } from '@/components/common/form/RadioInput'
import { cn } from '@/utils/classnames'
import { SUB_OPS, labelForSubOp } from './subops'
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
// Slack variant collapses approval_requests + approval_responses into a single
// "Include approval events" checkbox. We always keep the two booleans split on
// the underlying ResourceCfg and coerce both to the same value on slack write
// (documented choice — simplest, no indeterminate visual). Webhook variant
// exposes both checkboxes.
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

  const ops = cfg.ops ?? []
  const outcome: Outcome = cfg.outcome ?? 'all'
  const approvalRequests = !!cfg.approval_requests
  const approvalResponses = !!cfg.approval_responses
  const slackApprovalChecked = approvalRequests || approvalResponses
  const driftDetected = !!cfg.drift_detected
  // Only resources whose sub-op vocabulary includes "drift" can reasonably
  // emit a drift_detected event; we hide the toggle elsewhere to avoid
  // suggesting it does anything (the matcher would never fire it).
  const supportsDrift = SUB_OPS[kind].includes('drift')

  const opSummary =
    ops.length === 0 ? 'all ops' : `${ops.length} op${ops.length === 1 ? '' : 's'}`
  const approvalSummary =
    approvalRequests && approvalResponses
      ? 'approvals'
      : approvalRequests
        ? 'approval requests'
        : approvalResponses
          ? 'approval responses'
          : null

  const summaryParts = enabled
    ? [opSummary, OUTCOME_LABELS[outcome].toLowerCase(), approvalSummary].filter(
        Boolean
      ) as string[]
    : ['off']

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

        <Text variant="subtext" theme="neutral" className="ml-auto !mr-2 truncate">
          {summaryParts.join(' · ')}
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
          <Icon variant={expanded ? 'CaretUp' : 'CaretDown'} />
        </button>
      </div>

      {enabled && expanded ? (
        <div
          id={`interests-${kind}-body`}
          className="flex flex-col gap-4 border-t border-neutral-200 px-3 py-3 dark:border-neutral-700"
        >
          <div className="flex flex-col gap-2">
            <Text variant="subtext" weight="strong">
              Sub-operations
            </Text>
            <Text variant="subtext" theme="neutral">
              Leave all unchecked to receive every sub-operation.
            </Text>
            <div className="flex flex-wrap">
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

          <div className="flex flex-col gap-2">
            <Text variant="subtext" weight="strong">
              Notify on
            </Text>
            <div className="flex flex-wrap gap-x-2">
              {(['all', 'completion', 'failures'] as Outcome[]).map((o) => (
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
          </div>

          <div className="flex flex-col gap-2">
            <Text variant="subtext" weight="strong">
              Approval events
            </Text>
            {variant === 'slack' ? (
              <CheckboxInput
                id={`interests-${kind}-approval`}
                checked={slackApprovalChecked}
                onChange={(e) =>
                  setApproval(e.target.checked, e.target.checked)
                }
                disabled={disabled}
                labelProps={{
                  labelText: 'Include approval events',
                  labelTextProps: { variant: 'subtext' },
                }}
              />
            ) : (
              <div className="flex flex-col">
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
            )}
          </div>

          {supportsDrift ? (
            <div className="flex flex-col gap-2">
              <Text variant="subtext" weight="strong">
                Drift detection
              </Text>
              <CheckboxInput
                id={`interests-${kind}-drift-detected`}
                checked={driftDetected}
                onChange={(e) => setDriftDetected(e.target.checked)}
                disabled={disabled}
                labelProps={{
                  labelText:
                    'Notify only when drift is actually detected (not for clean drift scans)',
                  labelTextProps: { variant: 'subtext' },
                }}
              />
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  )
}
