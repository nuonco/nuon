import { useState } from 'react'
import { InterestsPicker } from './InterestsPicker'
import { allEvents } from './defaults'
import type { Interests } from './types'

export default { title: 'Features / Interests / Interests picker' }

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

export const AllEvents = () => <Wrapper initial={allEvents()} />

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

export const EmptyWarn = () => <Wrapper initial={{ resources: {} }} />

export const Disabled = () => <Wrapper initial={allEvents()} disabled />
