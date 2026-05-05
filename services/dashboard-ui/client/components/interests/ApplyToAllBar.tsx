import { Button } from '@/components/common/Button'
import { Text } from '@/components/common/Text'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { RadioInput } from '@/components/common/form/RadioInput'
import { OUTCOME_LABELS, type Outcome } from './types'

// Sticky bulk-edit bar that lives above the per-resource accordion. Writes the
// chosen outcome + approval flags into every currently-enabled resource on
// click. Renders a single "Include approval events" checkbox for the slack
// variant and two split checkboxes for the webhook variant.
export const ApplyToAllBar = ({
  variant,
  outcome,
  onOutcomeChange,
  approvalRequests,
  approvalResponses,
  onApprovalRequestsChange,
  onApprovalResponsesChange,
  onApply,
  disabled,
}: {
  variant: 'slack' | 'webhook'
  outcome: Outcome
  onOutcomeChange: (next: Outcome) => void
  approvalRequests: boolean
  approvalResponses: boolean
  onApprovalRequestsChange: (next: boolean) => void
  onApprovalResponsesChange: (next: boolean) => void
  onApply: () => void
  disabled?: boolean
}) => {
  // Slack collapses both approval booleans into ONE checkbox. We keep the
  // booleans split internally and coerce both to the same value on write.
  const slackApprovalChecked = approvalRequests || approvalResponses
  const setSlackApproval = (next: boolean) => {
    onApprovalRequestsChange(next)
    onApprovalResponsesChange(next)
  }

  return (
    <div className="sticky top-0 z-10 flex flex-col gap-3 rounded-md border border-neutral-200 bg-neutral-50 p-3 dark:border-neutral-700 dark:bg-neutral-800/60">
      <Text variant="subtext" weight="strong">
        Apply to all enabled resources
      </Text>

      <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
        <Text variant="subtext" theme="neutral">
          Notify
        </Text>
        {(['all', 'completion', 'failures'] as Outcome[]).map((o) => (
          <RadioInput
            key={o}
            name="apply-to-all-outcome"
            value={o}
            checked={outcome === o}
            onChange={() => onOutcomeChange(o)}
            disabled={disabled}
            labelProps={{
              labelText: OUTCOME_LABELS[o],
              labelTextProps: { variant: 'subtext' },
            }}
          />
        ))}
      </div>

      <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
        {variant === 'slack' ? (
          <CheckboxInput
            id="apply-to-all-approval"
            checked={slackApprovalChecked}
            onChange={(e) => setSlackApproval(e.target.checked)}
            disabled={disabled}
            labelProps={{
              labelText: 'Include approval events',
              labelTextProps: { variant: 'subtext' },
            }}
          />
        ) : (
          <>
            <CheckboxInput
              id="apply-to-all-approval-requests"
              checked={approvalRequests}
              onChange={(e) => onApprovalRequestsChange(e.target.checked)}
              disabled={disabled}
              labelProps={{
                labelText: 'Include approval requests',
                labelTextProps: { variant: 'subtext' },
              }}
            />
            <CheckboxInput
              id="apply-to-all-approval-responses"
              checked={approvalResponses}
              onChange={(e) => onApprovalResponsesChange(e.target.checked)}
              disabled={disabled}
              labelProps={{
                labelText: 'Include approval responses',
                labelTextProps: { variant: 'subtext' },
              }}
            />
          </>
        )}

        <Button
          type="button"
          size="sm"
          variant="secondary"
          onClick={onApply}
          disabled={disabled}
          className="ml-auto"
        >
          Apply to all
        </Button>
      </div>
    </div>
  )
}
