import { InstallActionRunOutputs } from '@/components/actions/InstallActionRunOutputs/InstallActionRunOutputs'
import { TerraformOutputs } from '@/components/terraform-outputs/TerraformOutputs'
import { Card } from '@/components/common/Card'
import { Code } from '@/components/common/Code'
import { CompositeError } from '@/components/common/CompositeError'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { JSONViewer } from '@/components/common/JSONViewer'
import { KeyValueList } from '@/components/common/KeyValueList'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Stack } from '@/components/common/Stack'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TCompositeError, TInstallActionRun, TWorkflowStep } from '@/types'

type TStatusHistoryEntry = NonNullable<
  NonNullable<TWorkflowStep['status']>['history']
>[number]

const StepHistoryStatus = ({
  status,
  id,
}: {
  status: TStatusHistoryEntry
  id: string
}) => {
  const description = status.status_human_description

  if (!description) {
    return (
      <span className="flex items-center gap-4 py-2">
        <Status status={status.status} variant="badge" />
        <Time
          seconds={status.created_at_ts}
          variant="subtext"
          theme="neutral"
        />
      </span>
    )
  }

  return (
    <Expand
      id={id}
      hasNoHoverStyle
      headerClassName="!p-0"
      heading={
        <span className="flex items-center gap-4 py-2">
          <Status status={status.status} variant="badge" />
          <Time
            seconds={status.created_at_ts}
            variant="subtext"
            theme="neutral"
          />
        </span>
      }
    >
      <Code className="mb-2 !text-xs">{description}</Code>
    </Expand>
  )
}

export interface IRunbookStepCard {
  step: TWorkflowStep
  workflowUrl: string
  eventHref?: (eventId: string, triggerId: string) => string
  targetData?: unknown
  deployOutputs?: Record<string, unknown>
  isLoading?: boolean
}

export const RunbookStepCard = ({
  step,
  workflowUrl,
  eventHref,
  targetData,
  deployOutputs,
  isLoading,
}: IRunbookStepCard) => {
  const isDeploy = step.step_target_type === 'install_deploys'
  const isActionRun = step.step_target_type === 'install_action_workflow_runs'
  const isSandboxRun = step.step_target_type === 'install_sandbox_runs'
  const targetLabel = isDeploy
    ? 'deploy'
    : isSandboxRun
      ? 'sandbox run'
      : 'action run'
  const actionRun = isActionRun ? (targetData as TInstallActionRun) : undefined
  const compositeError =
    typeof targetData === 'object' &&
    targetData !== null &&
    'composite_error' in targetData
      ? (targetData as { composite_error?: TCompositeError }).composite_error
      : undefined
  const envVarEntries = Object.entries(actionRun?.run_env_vars ?? {})
  const stepStatus =
    typeof step.status === 'object' ? step.status?.status : step.status
  const canShowCompositeError =
    stepStatus === 'error' ||
    stepStatus === 'failed-pending-retry' ||
    stepStatus === 'cancelled' ||
    stepStatus === 'discarded'
  const eventWait = step?.links?.event_wait

  return (
    <Card className="flex flex-col gap-4 p-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Text variant="body" className="!inline-flex gap-2">
            {stepStatus ? <Status status={stepStatus} isWithoutText /> : null}
            {step.name}
          </Text>
        </div>
        <Link href={`${workflowUrl}?panel=${step.id}`}>
          <Text variant="subtext" className="!inline-flex gap-1 items-center">
            View workflow <Icon size="12" variant="CaretRightIcon" />
          </Text>
        </Link>
      </div>

      {compositeError && canShowCompositeError ? (
        <CompositeError error={compositeError} />
      ) : null}

      {isActionRun && actionRun ? (
        <div className="flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Outputs
          </Text>
          <InstallActionRunOutputs installActionRun={actionRun} />
          {envVarEntries.length > 0 ? (
            <>
              <Text variant="base" weight="strong">
                Environment variables
              </Text>
              <KeyValueList
                values={envVarEntries.map(([key, value]) => ({ key, value }))}
              />
            </>
          ) : null}
        </div>
      ) : null}

      {isDeploy && deployOutputs && Object.keys(deployOutputs).length > 0 ? (
        <TerraformOutputs heading="Outputs" outputs={deployOutputs} />
      ) : null}

      {eventWait ? (
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between gap-4">
            <Text variant="base" weight="strong">
              Event wait
            </Text>
            <Status
              status={
                eventWait?.status === 'matched'
                  ? 'success'
                  : eventWait?.status === 'active'
                    ? 'info'
                    : eventWait?.status === 'expired'
                      ? 'error'
                      : 'neutral'
              }
              variant="badge"
            >
              {eventWait?.status || 'unknown'}
            </Status>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="flex flex-col gap-1">
              <Text variant="subtext" theme="neutral">
                Trigger
              </Text>
              <Text variant="subtext">
                {eventWait?.trigger_name || eventWait?.trigger_id || 'Unknown'}
              </Text>
            </div>
            <div className="flex flex-col gap-1">
              <Text variant="subtext" theme="neutral">
                Event types
              </Text>
              <Text variant="subtext">
                {eventWait?.event_types?.join(', ') || 'All event types'}
              </Text>
            </div>
            {eventWait?.activated_at ? (
              <div className="flex flex-col gap-1">
                <Text variant="subtext" theme="neutral">
                  Activated
                </Text>
                <Time
                  time={eventWait.activated_at}
                  format="long-datetime"
                  variant="subtext"
                />
              </div>
            ) : null}
            {eventWait?.matched_at ? (
              <div className="flex flex-col gap-1">
                <Text variant="subtext" theme="neutral">
                  Matched
                </Text>
                <Time
                  time={eventWait.matched_at}
                  format="long-datetime"
                  variant="subtext"
                />
              </div>
            ) : null}
          </div>
          {eventWait?.filters?.length ? (
            <div className="flex flex-col gap-2">
              <Text variant="subtext" weight="strong">
                Filters
              </Text>
              {eventWait.filters.map((filter, index) => (
                <Code key={`${filter?.path}-${index}`}>
                  {`${filter?.from || 'payload'} ${filter?.path || '—'} ${filter?.op || '—'} ${JSON.stringify(filter?.value)}`}
                </Code>
              ))}
            </div>
          ) : null}
          {eventWait?.matched_event_id && eventWait?.trigger_id && eventHref ? (
            <Link
              href={eventHref(eventWait.matched_event_id, eventWait.trigger_id)}
            >
              View matched event
              {eventWait?.matched_event_type
                ? ` · ${eventWait.matched_event_type}`
                : ''}
              <Icon variant="CaretRightIcon" />
            </Link>
          ) : null}
        </div>
      ) : null}

      <Stack gap={2}>
        {step?.created_by?.email ? (
          <Text variant="label" theme="neutral">
            Triggered by {step.created_by.email}
          </Text>
        ) : null}

        {step?.status?.history?.length ? (
          <Expand
            className="border rounded-md"
            id={`step-history-${step.id}`}
            heading={
              <Text family="mono" variant="subtext">
                View status history
              </Text>
            }
          >
            <div className="border-t flex flex-col p-4 divide-y">
              {step.status.history.map((status, idx) => (
                <StepHistoryStatus
                  key={`${status.created_at_ts}-${idx}`}
                  status={status}
                  id={`${step.id}-history-${idx}`}
                />
              ))}
              {step.status ? (
                <StepHistoryStatus
                  status={step.status}
                  id={`${step.id}-history-current`}
                />
              ) : null}
            </div>
          </Expand>
        ) : null}

        {isLoading ? (
          <Skeleton height="120px" width="100%" />
        ) : targetData ? (
          <Expand
            className="border rounded-md"
            id={`step-data-${step.id}`}
            heading={
              <Text family="mono" variant="subtext">
                View {targetLabel} JSON
              </Text>
            }
          >
            <div className="border-t">
              <JSONViewer
                data={targetData}
                expanded={1}
                showDataTypes={false}
              />
            </div>
          </Expand>
        ) : eventWait ? null : (
          <Text variant="subtext" theme="neutral">
            No data available
          </Text>
        )}
      </Stack>
    </Card>
  )
}
