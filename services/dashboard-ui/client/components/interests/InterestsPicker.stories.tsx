import { useState } from 'react'
import { InterestsPicker } from './InterestsPicker'
import { allEvents } from './defaults'
import { PRESETS } from './presets'
import type { Interests } from './types'
import type { SubscriptionMatch } from '@/components/match/types'
import type { PresetModalOutput } from './InterestsModal'

export default { title: 'Interests/InterestsPicker' }

const Wrapper = ({
  initial,
  initialMatch,
  disabled,
}: {
  initial: Interests
  initialMatch?: SubscriptionMatch
  disabled?: boolean
}) => {
  const [value, setValue] = useState<Interests>(initial)
  const [match, setMatch] = useState<SubscriptionMatch | undefined>(initialMatch)
  return (
    <div className="max-w-md p-6">
      <InterestsPicker
        value={value}
        matchValue={match}
        onChange={({ interests, match: nextMatch }: PresetModalOutput) => {
          setValue(interests)
          setMatch(nextMatch)
        }}
        disabled={disabled}
      />
      <pre className="mt-6 rounded-md bg-neutral-100 p-3 text-xs dark:bg-neutral-800">
        {JSON.stringify({ interests: value, match }, null, 2)}
      </pre>
    </div>
  )
}

export const Failures = () => (
  <Wrapper initial={PRESETS[0].build()} />
)

export const FailuresWithLabels = () => (
  <Wrapper
    initial={PRESETS[0].build()}
    initialMatch={{
      installs: {
        selector: { match_labels: { env: 'prod', tier: 'critical' } },
      },
    }}
  />
)

export const AllEvents = () => <Wrapper initial={allEvents()} />

export const Operations = () => (
  <Wrapper initial={PRESETS[2].build()} />
)

export const ApprovalsOnly = () => (
  <Wrapper initial={PRESETS[3].build()} />
)

export const DeploymentMilestones = () => (
  <Wrapper initial={PRESETS[4].build()} />
)

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
