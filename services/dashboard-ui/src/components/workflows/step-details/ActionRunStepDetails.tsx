'use client'

import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { InstallActionRunLogs } from '@/components/actions/InstallActionRunLogs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import { UnifiedLogsProvider } from '@/providers/unified-logs-provider-temp'
import { LogViewerProvider } from '@/providers/log-viewer-provider-temp'
import { LogsSkeleton as LogsViewerSkeleton } from '@/components/log-stream/Logs'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TInstallActionRun, TWorkflowStep, TOTELLog } from '@/types'
import { hydrateActionRunSteps } from '@/utils/action-utils'
import { toSentenceCase } from '@/utils/string-utils'

interface IActionRunStepDetails {
  step?: TWorkflowStep
}

export const ActionRunStepDetails = ({ step }: IActionRunStepDetails) => {
  const { org } = useOrg()

  const {
    data: actionRun,
    error,
    isLoading,
  } = useQuery<TInstallActionRun>({
    dependencies: [step],
    path: `/api/orgs/${org.id}/installs/${step.owner_id}/actions/runs/${step?.step_target_id}`,
  })

  const params = useQueryParams({
    order: actionRun?.log_stream?.open ? 'asc' : 'desc',
  })

  const { data: logs, isLoading: isLoadingLogs } = useQuery<TOTELLog[]>({
    dependencies: [actionRun?.log_stream?.id],
    path: actionRun?.log_stream?.id
      ? `/api/orgs/${org.id}/log-streams/${actionRun?.log_stream?.id}/logs${params}`
      : null,
  })

  return (
    <div className="flex flex-col gap-4">
      {isLoading && !actionRun ? (
        <ActionRunStepDetailsSkeleton />
      ) : error ? (
        <Text variant="base" weight="strong" theme="error">
          Unable to load action run details
        </Text>
      ) : (
        <>
          <div className="flex items-center gap-4">
            <Text variant="base" weight="strong">
              Action run
            </Text>

            <Text variant="subtext">
              <Link
                href={`/${org.id}/installs/${step.owner_id}/actions/${actionRun?.config?.action_workflow_id}`}
              >
                View action <Icon variant="CaretRight" />
              </Link>
            </Text>
            <Text variant="subtext">
              <Link
                href={`/${org.id}/installs/${step.owner_id}/actions/${actionRun?.config?.action_workflow_id}/${actionRun?.id}`}
              >
                View run <Icon variant="CaretRight" />
              </Link>
            </Text>
          </div>
          <>
            <div className="flex items-start gap-6">
              <LabeledStatus
                label="Status"
                statusProps={{
                  status: actionRun?.status_v2?.status,
                }}
                tooltipProps={{
                  position: 'top',
                  tipContent: actionRun?.status_v2?.status_human_description,
                }}
              />

              <LabeledValue label="Triggered by">
                <Badge size="md" variant="code">
                  {actionRun?.triggered_by_type}
                  {actionRun?.run_env_vars?.COMPONENT_ID ? (
                    <Link
                      href={`/${org.id}/installs/${step.owner_id}/components/${actionRun?.run_env_vars?.COMPONENT_ID}`}
                    >
                      {actionRun?.run_env_vars?.COMPONENT_NAME}
                    </Link>
                  ) : null}
                </Badge>
              </LabeledValue>
            </div>

            <div className="flex flex-col gap-2">
              <Text weight="strong">Action steps</Text>
              {hydrateActionRunSteps({
                steps: actionRun.steps,
                stepConfigs: actionRun?.config?.steps,
              })
                ?.sort(({ idx: a }, { idx: b }) => b - a)
                ?.reverse()
                .map((actionStep) => (
                  <span
                    key={actionStep.id}
                    className="py-2 px-4 border rounded-md flex items-center justify-between"
                  >
                    <span className="flex items-center gap-2">
                      <Status status={actionStep.status} isWithoutText />
                      <Text>{toSentenceCase(actionStep?.name)}</Text>
                    </span>

                    <Text
                      className="flex items-center gap-1"
                      variant="subtext"
                      theme="neutral"
                    >
                      {toSentenceCase(actionStep.status)}{' '}
                      {actionStep?.execution_duration > 1000000 ? (
                        <>
                          in{' '}
                          <Duration
                            variant="subtext"
                            nanoseconds={actionStep?.execution_duration}
                            theme="neutral"
                          />
                        </>
                      ) : null}
                    </Text>
                  </span>
                ))}
            </div>

            {actionRun?.log_stream ? (
              <div className="flex flex-col gap-2">
                <Text weight="strong">Action logs</Text>
                {isLoadingLogs && !logs?.length ? (
                  <ActionRunLogsSkeleton />
                ) : (
                  <LogStreamProvider
                    shouldPoll={actionRun?.log_stream?.open}
                    initLogStream={actionRun?.log_stream}
                  >
                    <UnifiedLogsProvider initLogs={logs}>
                      <LogViewerProvider>
                        <InstallActionRunLogs
                          actionConfig={actionRun?.config}
                          layout="horizontal"
                        />
                      </LogViewerProvider>
                    </UnifiedLogsProvider>
                  </LogStreamProvider>
                )}
              </div>
            ) : null}
          </>
        </>
      )}
    </div>
  )
}

const ActionRunStepDetailsSkeleton = () => {
  return (
    <>
      <div className="flex items-center gap-4">
        <Skeleton height="24px" width="76px" />

        <Skeleton height="17" width="85px" />
        <Skeleton height="17" width="70px" />
      </div>
      <div className="flex items-start gap-6">
        <LabeledValue label={<Skeleton height="17px" width="34px" />}>
          <Skeleton height="23px" width="75px" />
        </LabeledValue>

        <LabeledValue label={<Skeleton height="17px" width="34px" />}>
          <Skeleton height="23px" width="162px" />
        </LabeledValue>
      </div>

      <div className="flex flex-col gap-2">
        <Skeleton height="17px" width="80px" />
        {Array.from({ length: 3 }).map((_, idx) => (
          <Skeleton key={idx} height="42px" width="100%" />
        ))}
      </div>
    </>
  )
}

const ActionRunLogsSkeleton = () => {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap gap-2">
        <Skeleton height="32px" width="120px" />
        <Skeleton height="32px" width="100px" />
        <Skeleton height="32px" width="140px" />
        <Skeleton height="32px" width="110px" />
      </div>
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Skeleton height="36px" width="320px" />
            <Skeleton height="17px" width="85px" />
          </div>
          <div className="flex items-center gap-4">
            <Skeleton height="32px" width="86px" />
            <Skeleton height="32px" width="135px" />
            <Skeleton height="32px" width="140px" />
          </div>
        </div>
        <div>
          <LogsViewerSkeleton />
        </div>
      </div>
    </div>
  )
}
