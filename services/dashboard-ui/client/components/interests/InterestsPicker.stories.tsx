import { useState } from 'react'
import { InterestsPicker } from './InterestsPicker'
import { allEvents, defaultInterests } from './defaults'
import type { Interests } from './types'

export default { title: 'Interests/InterestsPicker' }

const Wrapper = ({
  variant,
  initial,
}: {
  variant: 'slack' | 'webhook'
  initial: Interests
}) => {
  const [value, setValue] = useState<Interests>(initial)
  return (
    <div className="max-w-2xl p-6">
      <InterestsPicker variant={variant} value={value} onChange={setValue} />
      <pre className="mt-6 rounded-md bg-neutral-100 p-3 text-xs dark:bg-neutral-800">
        {JSON.stringify(value, null, 2)}
      </pre>
    </div>
  )
}

export const SlackAllEvents = () => (
  <Wrapper variant="slack" initial={allEvents()} />
)

export const SlackPerResource = () => (
  <Wrapper variant="slack" initial={defaultInterests()} />
)

export const WebhookAllEvents = () => (
  <Wrapper variant="webhook" initial={allEvents()} />
)

export const WebhookPerResource = () => (
  <Wrapper variant="webhook" initial={defaultInterests()} />
)

export const SlackEmptyAllOff = () => (
  <Wrapper variant="slack" initial={{ resources: {} }} />
)

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
        components: {
          ops: ['deploy', 'drift'],
          outcome: 'all',
          approval_requests: false,
          approval_responses: true,
        },
      },
    }}
  />
)

export const SlackDisabled = () => (
  <div className="max-w-2xl p-6">
    <InterestsPicker
      variant="slack"
      value={defaultInterests()}
      onChange={() => {}}
      disabled
    />
  </div>
)
