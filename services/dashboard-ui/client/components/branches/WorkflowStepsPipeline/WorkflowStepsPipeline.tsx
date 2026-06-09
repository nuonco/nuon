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

function stepStatusIcon(status?: string) {
  if (status === 'in-progress') return 'PlayIcon'
  if (status === 'success') return 'CheckIcon'
  if (status === 'error') return 'XIcon'
  return 'ClockIcon'
}

function miniStatusColor(status?: string) {
  if (status === 'success') return 'bg-green-500'
  if (status === 'error') return 'bg-red-500'
  if (status === 'in-progress') return 'bg-blue-500 animate-pulse'
  if (status === 'skipped') return 'bg-cool-grey-300 dark:bg-dark-grey-500'
  return 'bg-cool-grey-400 dark:bg-dark-grey-500'
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
      <div className="flex items-start gap-6 py-6 px-4 min-w-max">
        {steps.map((step, idx) => {
          const stepStatus = step.status?.status || 'pending'
          const isInProgress = stepStatus === 'in-progress'
          const isSuccess = stepStatus === 'success'
          const isError = stepStatus === 'error'
          const isSelected = selectedStepId === step.id

          // Check for build metadata to show mini build rows
          const builds = (step.status?.metadata?.builds as any[]) || []
          const hasBuildRows = builds.length > 0

          return (
            <div key={step.id || idx} className="flex items-start gap-4">
              <div
                className={`flex flex-col min-w-[220px] max-w-[260px] rounded-lg transition-all cursor-pointer border-2 ${
                  isSelected
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
                {/* Step header */}
                <div className="flex flex-col items-center p-5 pb-3 w-full">
                  <div
                    className={`w-10 h-10 rounded-full flex items-center justify-center mb-2 transition-all ${
                      isInProgress
                        ? 'bg-blue-500 dark:bg-blue-600 text-white shadow-lg'
                        : isSuccess
                        ? 'bg-green-500 dark:bg-green-600 text-white shadow-md'
                        : isError
                        ? 'bg-red-500 dark:bg-red-600 text-white shadow-md'
                        : 'bg-cool-grey-300 dark:bg-dark-grey-400 text-cool-grey-600 dark:text-dark-grey-200'
                    }`}
                  >
                    <Icon variant={stepStatusIcon(stepStatus)} size={20} />
                  </div>

                  <Text variant="subtext" weight="stronger" className="text-center mb-0.5">
                    Step {idx + 1}
                  </Text>
                  <Text variant="subtext" theme="neutral" className="text-center max-w-[200px]">
                    {step.name || 'Unknown'}
                  </Text>

                  {step.execution_time ? (
                    <Text variant="subtext" theme="neutral" family="mono" className="mt-1">
                      {(step.execution_time / 1000000000).toFixed(1)}s
                    </Text>
                  ) : null}
                </div>

                {/* Build mini-rows inside the card */}
                {hasBuildRows && (
                  <div className="border-t border-cool-grey-200 dark:border-dark-grey-700 px-3 py-2 space-y-1">
                    {builds.slice(0, 8).map((build: any, i: number) => (
                      <div key={build.component_id || i} className="flex items-center gap-2">
                        <div className={`w-1.5 h-1.5 rounded-full shrink-0 ${miniStatusColor(build.status)}`} />
                        <Text variant="subtext" className="truncate flex-1 text-xs">
                          {build.component_name || 'unknown'}
                        </Text>
                      </div>
                    ))}
                    {builds.length > 8 && (
                      <Text variant="subtext" theme="neutral" className="text-xs pl-3.5">
                        +{builds.length - 8} more
                      </Text>
                    )}
                  </div>
                )}
              </div>

              {idx < steps.length - 1 && (
                <div className="flex items-center mt-12">
                  <Icon
                    variant="ArrowRightIcon"
                    size={28}
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
