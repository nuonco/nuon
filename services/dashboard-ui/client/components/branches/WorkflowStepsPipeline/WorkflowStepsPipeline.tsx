import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import type { TInstallWorkflowStep } from '@/types'

interface IWorkflowStepsPipeline {
  steps: TInstallWorkflowStep[]
  selectedStepId?: string
  onSelectStep: (step: TInstallWorkflowStep) => void
}

export const WorkflowStepsPipeline = ({
  steps,
  selectedStepId,
  onSelectStep,
}: IWorkflowStepsPipeline) => {
  if (steps.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 gap-4">
        <Loading variant="large" />
        <Text variant="body" theme="neutral">
          Generating workflow steps...
        </Text>
      </div>
    )
  }

  return (
    <div
      className="relative overflow-x-auto overflow-y-hidden"
      style={{ scrollbarWidth: 'thin', scrollBehavior: 'smooth' }}
    >
      <div className="flex items-center gap-6 py-6 px-4 min-w-max">
        {steps.map((step, idx) => {
          const stepStatus = step.status?.status || 'pending'
          const isInProgress = stepStatus === 'in-progress'
          const isSuccess = stepStatus === 'success'
          const isError = stepStatus === 'error'

          return (
            <div key={step.id || idx} className="flex items-center gap-4">
              <div
                className={`flex flex-col items-center min-w-[240px] p-8 rounded-lg transition-all cursor-pointer border-2 ${
                  selectedStepId === step.id
                    ? 'ring-2 ring-primary-300 dark:ring-primary-700 shadow-2xl scale-105 bg-primary-50 dark:bg-dark-grey-900 border-primary-200 dark:border-primary-400/50'
                    : isInProgress
                    ? 'ring-2 ring-blue-200 dark:ring-blue-800 shadow-xl hover:shadow-2xl bg-blue-50 dark:bg-dark-grey-900 border-blue-400 dark:border-blue-500/40'
                    : isSuccess
                    ? 'shadow-lg hover:shadow-xl bg-green-50 dark:bg-dark-grey-900 border-green-400 dark:border-green-500/40'
                    : isError
                    ? 'shadow-lg hover:shadow-xl bg-red-50 dark:bg-dark-grey-900 border-red-300 dark:border-red-500/40'
                    : 'border-dashed border-cool-grey-300 dark:border-dark-grey-600 hover:border-solid hover:shadow-md bg-cool-grey-50 dark:bg-dark-grey-900'
                }`}
                onClick={() => onSelectStep(step)}
              >
                <div
                  className={`w-16 h-16 rounded-full flex items-center justify-center mb-4 transition-all ${
                    isInProgress
                      ? 'bg-blue-500 dark:bg-blue-600 text-white shadow-lg'
                      : isSuccess
                      ? 'bg-green-500 dark:bg-green-600 text-white shadow-md'
                      : isError
                      ? 'bg-red-500 dark:bg-red-600 text-white shadow-md'
                      : 'bg-cool-grey-300 dark:bg-dark-grey-400 text-cool-grey-600 dark:text-dark-grey-200'
                  }`}
                >
                  {isInProgress ? (
                    <Icon variant="Play" size={32} />
                  ) : isSuccess ? (
                    <Icon variant="Check" size={32} />
                  ) : isError ? (
                    <Icon variant="X" size={32} />
                  ) : (
                    <Icon variant="Clock" size={28} />
                  )}
                </div>

                <Text variant="base" weight="stronger" className="text-center mb-2">
                  Step {idx + 1}
                </Text>
                <Text variant="base" theme="neutral" className="text-center mb-3 max-w-[200px]">
                  {step.name || 'Unknown'}
                </Text>

                <div className="flex flex-col gap-2 items-center w-full">
                  {step.group_idx !== undefined && (
                    <Badge
                      theme={
                        isInProgress ? 'info'
                        : isSuccess ? 'success'
                        : isError ? 'error'
                        : 'neutral'
                      }
                      size="md"
                    >
                      Group {step.group_idx}
                    </Badge>
                  )}
                  {step.execution_time && (
                    <Text variant="base" theme="neutral" family="mono" weight="strong">
                      {(step.execution_time / 1000000000).toFixed(1)}s
                    </Text>
                  )}
                </div>
              </div>

              {idx < steps.length - 1 && (
                <div className="flex items-center">
                  <Icon
                    variant="ArrowRight"
                    size={36}
                    className={`transition-colors ${
                      isSuccess
                        ? 'text-green-500 dark:text-green-400'
                        : 'text-cool-grey-400 dark:text-dark-grey-500'
                    }`}
                  />
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
