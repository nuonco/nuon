import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import type { TCreateWizardStepStatus } from './wizard-path'

interface ICreateStepper {
  steps: { id: string; label: string; status: TCreateWizardStepStatus }[]
  selectedId: string
  onSelect: (id: string) => void
}

export const CreateStepper = ({ steps, selectedId, onSelect }: ICreateStepper) => {
  const currentIndex = steps.findIndex((step) => step.status === 'current')
  const position = currentIndex === -1 ? steps.length : currentIndex + 1

  return (
    <nav className="flex flex-col gap-3" aria-label="Setup steps">
      <ol className="flex flex-wrap items-center gap-x-2 gap-y-3">
        {steps.map((step, index) => {
          const selected = step.id === selectedId
          const reachable = step.status !== 'upcoming'
          return (
            <li key={step.id} className="flex min-w-0 items-center gap-2">
              <button
                type="button"
                disabled={!reachable}
                aria-current={step.status === 'current' ? 'step' : undefined}
                aria-pressed={selected}
                aria-label={`${index + 1}. ${step.label}`}
                onClick={() => reachable && onSelect(step.id)}
                className={cn(
                  'flex items-center gap-2 rounded-lg px-1 py-0.5 text-left outline-none',
                  'focus-visible:[--tw-outline-style:solid] focus-visible:outline-1 focus-visible:outline-offset-0 focus-visible:outline-primary-400/80',
                  reachable
                    ? 'cursor-pointer hover:bg-black/5 dark:hover:bg-white/5'
                    : 'cursor-not-allowed'
                )}
              >
                {step.status === 'done' ? (
                  <Icon
                    variant="CheckCircleIcon"
                    size={22}
                    className="text-green-600 dark:text-green-500"
                  />
                ) : (
                  <span
                    className={cn(
                      'flex size-6 shrink-0 items-center justify-center rounded-full border',
                      step.status === 'current'
                        ? 'border-primary-600 dark:border-primary-400'
                        : 'border-cool-grey-400 dark:border-dark-grey-500'
                    )}
                  >
                    <Text
                      variant="label"
                      theme={step.status === 'current' ? 'brand' : 'neutral'}
                    >
                      {index + 1}
                    </Text>
                  </span>
                )}
                <Text
                  variant="body"
                  weight={selected || step.status === 'current' ? 'strong' : 'normal'}
                  theme={
                    step.status === 'upcoming'
                      ? 'neutral'
                      : selected || step.status === 'current'
                        ? 'brand'
                        : 'default'
                  }
                >
                  {step.label}
                </Text>
              </button>
              {index === steps.length - 1 ? null : (
                <span
                  aria-hidden
                  className="h-px w-6 shrink-0 bg-cool-grey-300 dark:bg-dark-grey-600"
                />
              )}
            </li>
          )
        })}
      </ol>
      <Text variant="subtext" theme="neutral">
        Step {position} of {steps.length}
      </Text>
    </nav>
  )
}
