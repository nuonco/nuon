import { ActionTriggerType } from '@/components/actions/ActionTriggerType'
import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Code } from '@/components/common/Code'
import { Duration } from '@/components/common/Duration'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RunnerJobPlanButton } from '@/components/runners/RunnerJobPlan'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import { useInstall } from '@/hooks/use-install'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import { useOrg } from '@/hooks/use-org'
import { useAuth } from '@/hooks/use-auth'
import type { TWorkflow } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { getWorkflowStep } from '@/utils/workflow-utils'
import type { TActionConfigTriggerType } from '@/types'

interface IInstallActionRunHeader {
  actionId: string
  actionName: string
  workflow: TWorkflow
}

export const InstallActionRunHeader = ({
  actionId,
  actionName,
  workflow,
}: IInstallActionRunHeader) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { installActionRun } = useInstallActionRun()
  const { isAdmin } = useAuth()
  const step = getWorkflowStep({
    workflow,
    stepTargetId: installActionRun?.id,
  })
  const basePath = `/${org?.id}/installs/${install?.id}`

  return (
    <header className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center gap-4 justify-between w-full">
        <HeadingGroup>
          <BackLink className="mb-4" />
          <Text
            className="inline-flex items-center gap-4 mb-2"
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
          <CancelWorkflowButton workflow={workflow} />
          {installActionRun?.runner_job?.id ? (
            <RunnerJobPlanButton
              runnerJobId={installActionRun?.runner_job?.id}
            />
          ) : null}
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

          <LabeledValue label="Execution role">
            {installActionRun?.runner_job?.json?.composite_plan?.plan_auth
              ?.aws_auth?.assume_role?.role_arn ? (
              <Code variant="inline" className="text-xs">
                {
                  installActionRun.runner_job.json.composite_plan.plan_auth
                    .aws_auth.assume_role.role_arn
                }
              </Code>
            ) : (
              <Text variant="subtext" theme="neutral">
                —
              </Text>
            )}
          </LabeledValue>
        </div>
      </Card>

      {workflow ? (
        <Button
          href={`/${org.id}/installs/${install.id}/workflows/${workflow.id}?panel=${step?.id}`}
        >
          View workflow
          <Icon variant="CaretRightIcon" />
        </Button>
      ) : null}
    </header>
  )
}
