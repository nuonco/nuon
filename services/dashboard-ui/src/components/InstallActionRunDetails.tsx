'use client'

import { Duration, EventStatus, JsonView, Section, Text } from '@/components'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import type { TInstallActionWorkflowRun, TActionConfig } from '@/types'
import { sentanceCase } from '@/utils'

function hydrateRunSteps(
  steps: TInstallActionWorkflowRun['steps'],
  stepConfigs: TActionConfig['steps']
) {
  return steps?.map((step) => {
    const config = stepConfigs?.find((cfg) => cfg.id === step.step_id)
    return {
      name: config?.name,
      idx: config.idx,
      ...step,
    }
  })
}

export const InstallActionRunDetails = () => {
  const { installActionRun } = useInstallActionRun()
  return (
    <div className="divide-y flex flex-col md:col-span-4">
      <Section
        className="flex-initial"
        heading={`${installActionRun?.steps?.reduce((count, step) => {
          if (step.status === 'finished' || step.status === 'error') count++
          return count
        }, 0)} of ${installActionRun?.config?.steps?.length} Steps`}
      >
        <div className="flex flex-col gap-2">
          {hydrateRunSteps(
            installActionRun?.steps,
            installActionRun?.config?.steps
          )
            ?.sort(({ idx: a }, { idx: b }) => b - a)
            ?.reverse()
            ?.map((step) => {
              return (
                <span key={step.id} className="py-2">
                  <span className="flex items-center gap-3">
                    <EventStatus status={step.status} />
                    <Text variant="med-14">{step?.name}</Text>
                  </span>

                  <Text className="flex items-center ml-7" variant="reg-12">
                    {sentanceCase(step.status)}{' '}
                    {step?.execution_duration > 1000000 ? (
                      <>
                        in <Duration nanoseconds={step?.execution_duration} />
                      </>
                    ) : null}
                  </Text>
                </span>
              )
            })}
        </div>
      </Section>
      {installActionRun?.runner_job?.outputs ? (
        <Section className="flex-initial" heading="Action run outputs">
          <JsonView data={installActionRun?.runner_job?.outputs} />
        </Section>
      ) : null}
    </div>
  )
}
