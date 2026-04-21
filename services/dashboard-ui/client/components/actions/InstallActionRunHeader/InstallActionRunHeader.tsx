import type { ReactNode } from 'react'
import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Code } from '@/components/common/Code'
import type { TActionConfigTriggerType, TInstallActionRun, TWorkflow, TWorkflowStep } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'

interface IInstallActionRunHeader {
  actionId: string
  actionName: string
  workflow: TWorkflow
  installActionRun: TInstallActionRun
  basePath: string
  isAdmin: boolean
  step?: TWorkflowStep
  cancelWorkflowButton: ReactNode
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
  runnerJobPlanButton,
}: IInstallActionRunHeader) => {
  return (
    <header className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center gap-4 justify-between w-full">
        <HeadingGroup>
          <BackLink className="mb-4" />
          <Text
            flex
            className="gap-4 mb-2"
            variant="h3"
            weight="strong"
          >
            {actionName}{' '}
            <ActionTriggerType
              componentName={installActionRun?.run_env_vars?.COMPONENT_NAME}
              componentPath={`${basePath}/components/${installActionRun?.run_env_vars?.COMPONENT_ID}`}
              triggerType={
                installActionRun?.triggered_by_type as TActionConfigTriggerType
              }
              size="sm"
            />
          </Text>
          <span className="flex items-center gap-4">
            <ID>{actionId}</ID>{' '}
            {isAdmin && installActionRun?.install_action_workflow_id ? (
              <ID>{String(installActionRun.install_action_workflow_id)}</ID>
            ) : null}
          </span>
          <Time
            time={installActionRun?.updated_at}
            format="relative"
            variant="subtext"
            theme="info"
          />
        </HeadingGroup>

        <div className="flex items-center gap-4">
          {cancelWorkflowButton}
          {runnerJobPlanButton}
        </div>
      </div>

      <Card>
        <div className="grid grid-cols-5">
          <LabeledValue
            label={`Triggered via ${installActionRun?.triggered_by_type}`}
          >
            {installActionRun?.created_by?.email || (
              <ID theme="default">{installActionRun?.created_by_id}</ID>
            )}
          </LabeledValue>

          <LabeledStatus
            label="Status"
            statusProps={{
              status: installActionRun?.status_v2?.status,
            }}
            tooltipProps={{
              tipContent: toSentenceCase(
                installActionRun?.status_v2?.status_human_description
              ),
            }}
          />

          <LabeledValue label={`Total duration`}>
            <Duration nanoseconds={installActionRun?.execution_time} />
          </LabeledValue>

          <LabeledValue label={`Timeout`}>
            <Duration nanoseconds={installActionRun?.config?.timeout} />
          </LabeledValue>

          <LabeledValue label="Execution role">
            {installActionRun?.runner_job?.install_role_usage?.role_name ? (
              <Text variant="subtext" family='mono' className="text-xs">
                {installActionRun.runner_job.install_role_usage.role_name}
              </Text>
            ) : (
              <Text variant="subtext" theme="neutral">
                -
              </Text>
            )}
          </LabeledValue>
        </div>
      </Card>

      {workflow ? (
        <Button
          href={`${basePath}/workflows/${workflow.id}?panel=${step?.id}`}
        >
          View workflow
          <Icon variant="CaretRightIcon" />
        </Button>
      ) : null}
    </header>
  )
}
