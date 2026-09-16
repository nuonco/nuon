import { cn } from '@/utils/classnames'
import { Icon } from '../../atoms/Icon'
import { Text } from '../../atoms/Text'
import type { TWizardStepStatus } from '../../../utils/wizard'

export interface IWizardStepperStep {
  id: string
  label: string
  status: TWizardStepStatus
}

export interface IWizardStepper {
  steps: IWizardStepperStep[]
  selectedId?: string
  onSelect: (id: string) => void
}

export const WizardStepper = ({
  steps,
  selectedId,
  onSelect,
}: IWizardStepper) => {
  const currentIndex = steps.findIndex((step) => step.status === 'current')
  const position = currentIndex === -1 ? steps.length : currentIndex + 1

  return (
    <nav className="flex flex-col gap-3" aria-label="Setup steps">
      <ol className="flex flex-wrap items-center gap-x-2 gap-y-3">
        {steps.map((step, index) => {
          const last = index === steps.length - 1
          const selected = step.id === selectedId
          const reachable = step.status !== 'upcoming'
          const label = `${index + 1}. ${step.label}`

          return (
            <li key={step.id} className="flex min-w-0 items-center gap-2">
              <button
                type="button"
                disabled={!reachable}
                aria-current={step.status === 'current' ? 'step' : undefined}
                aria-label={label}
                onClick={() => reachable && onSelect(step.id)}
                className={cn(
                  'flex items-center gap-2 rounded-lg px-1 py-0.5 text-left outline-none focus-ring',
                  reachable ? 'cursor-pointer' : 'cursor-not-allowed'
                )}
              >
                <span
                  className={cn(
                    'flex size-6 shrink-0 items-center justify-center rounded-full border',
                    step.status === 'done' &&
                      'border-transparent text-status-success',
                    step.status === 'current' &&
                      'border-divider-accent bg-surface-accent text-accent',
                    step.status === 'upcoming' && 'border-divider text-tertiary'
                  )}
                >
                  {step.status === 'done' ? (
                    <Icon variant="CheckCircleIcon" size={16} />
                  ) : (
                    <Text
                      as="span"
                      variant="label"
                      color="inherit"
                      className="tabular-nums"
                    >
                      {index + 1}
                    </Text>
                  )}
                </span>
                <Text
                  as="span"
                  variant="caption"
                  color={
                    step.status === 'upcoming'
                      ? 'tertiary'
                      : selected
                        ? 'primary'
                        : 'secondary'
                  }
                  weight={selected ? 'medium' : 'normal'}
                >
                  {step.label}
                </Text>
              </button>
              {last ? null : (
                <span
                  aria-hidden
                  className="h-px w-4 shrink-0 bg-divider sm:w-8"
                />
              )}
            </li>
          )
        })}
      </ol>
      <Text variant="caption" color="secondary">
        Step {Math.min(position, steps.length)} of {steps.length}
      </Text>
    </nav>
  )
}
