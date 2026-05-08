import { useState } from 'react'
import { InterestsPicker } from './InterestsPicker'
import { allEvents, defaultInterests } from './defaults'
import type { Interests } from './types'

export default { title: 'Interests/InterestsPicker' }

const Wrapper = ({
  variant,
  initial,
  disabled,
}: {
  variant: 'slack' | 'webhook'
  initial: Interests
  disabled?: boolean
}) => {
  const [value, setValue] = useState<Interests>(initial)
  return (
    <div className="max-w-2xl p-6">
      <InterestsPicker
        variant={variant}
        value={value}
        onChange={setValue}
        disabled={disabled}
      />
      <pre className="mt-6 rounded-md bg-neutral-100 p-3 text-xs dark:bg-neutral-800">
        {JSON.stringify(value, null, 2)}
      </pre>
    </div>
  )
}

// Collapsed default — every per-resource row is closed; the user has not yet
// expanded anything. Useful baseline for the rest of the stories.
export const SlackCollapsedDefault = () => (
  <Wrapper variant="slack" initial={defaultInterests()} />
)

export const WebhookCollapsedDefault = () => (
  <Wrapper variant="webhook" initial={defaultInterests()} />
)

// "Send all events" sentinel — Resources is omitted from the wire shape.
export const SlackAllEvents = () => (
  <Wrapper variant="slack" initial={allEvents()} />
)

export const WebhookAllEvents = () => (
  <Wrapper variant="webhook" initial={allEvents()} />
)

// Expanded with all three categories on for the components row. Drift is only
// rendered for components + sandboxes.
export const SlackExpandedAllCategories = () => (
  <Wrapper
    variant="slack"
    initial={{
      resources: {
        components: {
          outcome: 'completion',
          approval_requests: true,
          approval_responses: true,
          drift_detected: true,
        },
      },
    }}
  />
)

// Lifecycle off (outcome: 'none') with drift on — the "drift only" summary
// case for components.
export const SlackDriftOnly = () => (
  <Wrapper
    variant="slack"
    initial={{
      resources: {
        components: {
          outcome: 'none',
          approval_requests: false,
          approval_responses: false,
          drift_detected: true,
        },
      },
    }}
  />
)

// Webhook variant with the split approvals sub-row visible. approval_requests
// is on, approval_responses is off — the canonical reason to use the webhook
// variant over slack.
export const WebhookSplitApprovals = () => (
  <Wrapper
    variant="webhook"
    initial={{
      resources: {
        installs: {
          outcome: 'completion',
          approval_requests: true,
          approval_responses: false,
        },
      },
    }}
  />
)

// "No events" warn state — resource is enabled but every category is off, so
// the matcher will never fire for it. The summary chip switches to the warn
// theme to make this visible at the collapsed-row level.
export const SlackNoEventsWarn = () => (
  <Wrapper
    variant="slack"
    initial={{
      resources: {
        installs: {
          outcome: 'none',
          approval_requests: false,
          approval_responses: false,
        },
      },
    }}
  />
)

// Empty resources map — the picker surfaces a banner explaining the
// subscription will receive nothing. Distinct from the per-row "no events"
// state above.
export const SlackEmptyAllOff = () => (
  <Wrapper variant="slack" initial={{ resources: {} }} />
)

export const SlackDisabled = () => (
  <Wrapper variant="slack" initial={defaultInterests()} disabled />
)
