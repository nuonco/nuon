import { useState } from 'react'
import { InterestsPicker } from './InterestsPicker'
import { allEvents } from './defaults'
import type { Interests } from './types'

export default { title: 'Interests/InterestsPicker' }

const Wrapper = ({
  initial,
  disabled,
}: {
  initial: Interests
  disabled?: boolean
}) => {
  const [value, setValue] = useState<Interests>(initial)
  return (
    <div className="max-w-md p-6">
      <InterestsPicker value={value} onChange={setValue} disabled={disabled} />
      <pre className="mt-6 rounded-md bg-neutral-100 p-3 text-xs dark:bg-neutral-800">
        {JSON.stringify(value, null, 2)}
      </pre>
    </div>
  )
}

// New-subscription baseline — every event matches. The button summary reads
// "All events". Click to open the modal with the All-events radio selected
// and the checklist hidden.
export const AllEvents = () => <Wrapper initial={allEvents()} />

// "Choose specific events" mode populated with a typical four-resource set.
// Summary chip shows the rolled-up "N events selected" count.
export const SpecificEventsPopulated = () => (
  <Wrapper
    initial={{
      resources: {
        installs: {
          outcome: 'completion',
          approval_requests: true,
          approval_responses: true,
        },
        components: {
          outcome: 'completion',
          approval_requests: true,
          approval_responses: true,
          drift_detected: true,
        },
        sandboxes: {
          outcome: 'completion',
          drift_detected: true,
        },
      },
    }}
  />
)

// Only drift_detected is on — lifecycle category is muted (outcome: 'none').
// Exercises the per-category collapsing that happens when toggling lifecycle
// off but leaving drift on.
export const DriftOnly = () => (
  <Wrapper
    initial={{
      resources: {
        components: {
          outcome: 'none',
          drift_detected: true,
        },
      },
    }}
  />
)

// Explicitly-empty resources map — backend matches nothing. The summary
// switches to the warn tone "No events selected" so it's obvious at a glance.
export const EmptyWarn = () => <Wrapper initial={{ resources: {} }} />

// Disabled — button is non-interactive. Useful for read-only contexts.
export const Disabled = () => <Wrapper initial={allEvents()} disabled />
