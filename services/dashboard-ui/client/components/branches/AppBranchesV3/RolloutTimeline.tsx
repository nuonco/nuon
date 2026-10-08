import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'

const STEPS = [
  { id: 'starting', label: 'Starting workflow' },
  { id: 'fetch-commit', label: 'Fetch commit' },
  { id: 'app-config', label: 'Compile and diff app config' },
  { id: 'build-components', label: 'Build components' },
  { id: 'rollout', label: 'Rollout' },
]

export const RolloutTimeline = ({ status }: { status: string }) => (
  <ol
    aria-label="Run progress"
    className="flex min-w-0 flex-wrap items-center gap-y-2"
  >
    {STEPS.map((step, index) => {
      const stepStatus = step.id === 'rollout' ? status : 'success'
      return (
        <li key={step.id} className="flex items-center">
          {index > 0 ? (
            <span
              aria-hidden
              className="mx-2.5 h-px w-3 shrink-0 bg-cool-grey-300 dark:bg-white/20"
            />
          ) : null}
          <span className="flex items-center gap-1.5">
            <Status status={stepStatus} isWithoutText />
            <Text
              variant="subtext"
              weight={stepStatus === 'success' ? 'normal' : 'strong'}
              theme={stepStatus === 'pending' ? 'neutral' : undefined}
            >
              {step.label}
            </Text>
          </span>
        </li>
      )
    })}
  </ol>
)
