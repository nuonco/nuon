'use client'

import { CaretRightIcon } from '@phosphor-icons/react'
import { ActionTriggerType } from '@/components/old/ActionTriggerType'
import { Link } from '@/components/old/Link'
import { Loading } from '@/components/old/Loading'
import { Notice } from '@/components/old/Notice'
import { StatusBadge } from '@/components/old/Status'
import { EventStatus } from '@/components/old/Timeline'
import { Duration } from '@/components/old/Time'
import { Text } from '@/components/old/Typography'
import { useOrg } from '@/hooks/use-org'
import { usePolling } from '@/hooks/use-polling'
import type { TActionConfig, TInstallActionRun } from '@/types'
import { sentanceCase } from '@/utils'
import type { IPollStepDetails } from './InstallWorkflowSteps'

// hydrate run steps with idx and name
function hydrateRunSteps(
  steps: TInstallActionRun['steps'],
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

export const ActionStepDetails = ({
  step,
  shouldPoll = false,
  pollInterval = 5000,
}: IPollStepDetails) => {
  const { org } = useOrg()
  const {
    data: actionRun,
    isLoading,
    error,
  } = usePolling<TInstallActionRun>({
    initIsLoading: true,
    path: `/api/orgs/${org.id}/installs/${step?.owner_id}/actions/runs/${step?.step_target_id}`,
    pollInterval,
    shouldPoll,
  })

  const componentName = actionRun?.run_env_vars?.COMPONENT_NAME
  const componentPath = `/${org.id}/installs/${step?.owner_id}/components/${actionRun?.run_env_vars?.COMPONENT_ID}`

  return (
    <>
      {isLoading && !actionRun ? (
        <div className="border rounded-md p-6">
          <Loading
            loadingText="Loading action run details..."
            variant="stack"
          />
        </div>
      ) : (
        <>
          {error?.error ? (
            <Notice>
              {error?.error || 'Unable to load action run details.'}
            </Notice>
          ) : null}
          {actionRun ? (
            <div className="flex flex-col border rounded-md shadow">
              <div className="flex items-center justify-between p-3 border-b">
                <Text variant="med-14" className="inline-flex gap-4">
                  Action run{' '}
                  {componentName &&
                  (actionRun?.triggered_by_type === 'pre-deploy-component' ||
                    actionRun.triggered_by_type === 'post-deploy-component') ? (
                    <ActionTriggerType
                      triggerType={actionRun?.triggered_by_type}
                      componentName={componentName}
                      componentPath={componentPath}
                    />
                  ) : null}
                </Text>
                <div className="flex items-center gap-4">
                  <Link
                    className="text-sm gap-0"
                    href={`/${org.id}/installs/${step?.owner_id}/actions/${actionRun?.config?.action_workflow_id}`}
                  >
                    View action
                    <CaretRightIcon />
                  </Link>
                  <Link
                    className="text-sm gap-0"
                    href={`/${org.id}/installs/${step?.owner_id}/actions/${actionRun?.config?.action_workflow_id}/${actionRun?.id}`}
                  >
                    View run
                    <CaretRightIcon />
                  </Link>
                </div>
              </div>

              <div className="p-6 flex flex-col gap-4">
                <StatusBadge
                  status={actionRun?.status_v2?.status || actionRun?.status}
                  description={
                    actionRun?.status_v2?.status_human_description ||
                    actionRun?.status_description
                  }
                  label="Action status"
                />
                <div className="flex flex-col gap-2">
                  <Text isMuted className="tracking-wide">
                    Action steps
                  </Text>
                  {hydrateRunSteps(actionRun?.steps, actionRun?.config?.steps)
                    ?.sort(({ idx: a }, { idx: b }) => b - a)
                    ?.reverse()
                    ?.map((actionStep) => {
                      return (
                        <span
                          key={actionStep.id}
                          className="py-2 px-4 border rounded-md flex items-center justify-between"
                        >
                          <span className="flex items-center gap-3">
                            <EventStatus status={actionStep.status} />
                            <Text variant="med-14">{actionStep?.name}</Text>
                          </span>

                          <Text
                            className="flex items-center ml-7"
                            variant="reg-12"
                          >
                            {sentanceCase(actionStep.status)}{' '}
                            {actionStep?.execution_duration > 1000000 ? (
                              <>
                                in{' '}
                                <Duration
                                  nanoseconds={actionStep?.execution_duration}
                                />
                              </>
                            ) : null}
                          </Text>
                        </span>
                      )
                    })}
                </div>
              </div>
            </div>
          ) : null}
        </>
      )}
    </>
  )
}
