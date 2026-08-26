import type { ReactNode } from 'react'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DetailHeader } from '@/components/layout/DetailHeader'
import type {
  TActionConfigTriggerType,
  TInstallActionRun,
  TWorkflow,
  TWorkflowStep,
} from '@/types'

interface IInstallActionRunHeader {
  actionId: string
  actionName: string
  workflow: TWorkflow
  installActionRun: TInstallActionRun
  basePath: string
  isAdmin: boolean
  step?: TWorkflowStep
  cancelWorkflowButton: ReactNode
  runActionButton?: ReactNode
  runnerJobPlanButton?: ReactNode
}

export const InstallActionRunHeader = ({
  actionId,
  actionName,
  workflow,
  installActionRun,
  basePath,
  isAdmin,
  step,
  cancelWorkflowButton,
  runActionButton,
  runnerJobPlanButton,
}: IInstallActionRunHeader) => {
  return (
    <DetailHeader
      title={actionName}
      status={
        <ActionTriggerType
          componentName={installActionRun?.run_env_vars?.COMPONENT_NAME}
          componentPath={`${basePath}/components/${installActionRun?.run_env_vars?.COMPONENT_ID}`}
          triggerType={
            installActionRun?.triggered_by_type as TActionConfigTriggerType
          }
          size="sm"
        />
      }
      id={actionId}
      identity={
        <>
          {isAdmin && installActionRun?.install_action_workflow_id ? (
            <ID>{String(installActionRun.install_action_workflow_id)}</ID>
          ) : null}
          <Time
            time={installActionRun?.updated_at}
            format="relative"
            variant="subtext"
            theme="info"
          />
        </>
      }
      actions={
        <>
          {cancelWorkflowButton}
          {runnerJobPlanButton}
          {runActionButton}
        </>
      }
      metadata={
        <>
          <LabeledStatus
            label="Status"
            statusProps={{
              status: installActionRun?.status_v2?.status,
            }}
            tooltipProps={{
              tipContent: installActionRun?.status_v2?.status_human_description,
            }}
          />
          <LabeledValue label="Total duration">
            <Duration nanoseconds={installActionRun?.execution_time} />
          </LabeledValue>
          <LabeledValue label="Timeout">
            <Duration nanoseconds={installActionRun?.config?.timeout} />
          </LabeledValue>
          <LabeledValue
            label={`Triggered via ${installActionRun?.triggered_by_type}`}
          >
            {installActionRun?.created_by?.email || (
              <ID theme="default">{installActionRun?.created_by_id}</ID>
            )}
          </LabeledValue>
          <LabeledValue label="Execution role">
            {installActionRun?.runner_job?.install_role_usage?.role_name ? (
              <Text variant="subtext" family="mono" className="text-xs">
                {installActionRun.runner_job.install_role_usage.role_name}
              </Text>
            ) : (
              <Text variant="subtext" theme="neutral">
                —
              </Text>
            )}
          </LabeledValue>
          {installActionRun?.config?.image ? (
            <LabeledValue label="Container image">
              <Text
                variant="subtext"
                family="mono"
                className="text-xs break-all"
              >
                {installActionRun.config.image}
              </Text>
            </LabeledValue>
          ) : null}
        </>
      }
    >
      {workflow ? (
        <Button href={`${basePath}/workflows/${workflow.id}?panel=${step?.id}`}>
          View workflow
        </Button>
      ) : null}
    </DetailHeader>
  )
}
