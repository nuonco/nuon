import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import type { TCloudPlatform, TWorkflow, TSandboxRun, TInstall } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { SandboxRunSwitcher } from '../SandboxRunSwitcher'
import { ManageRunDropdown } from '@/components/sandbox/management/ManageRunDropdown'
import { SandboxConfigContextTooltip } from '@/components/sandbox/SandboxConfigContextTooltip'

interface ISandboxHeader {
  workflow: TWorkflow
  stepId: string
  sandboxRun: TSandboxRun
  install: TInstall
  orgId: string
  flush?: boolean
}

export const SandboxHeader = ({
  workflow,
  stepId,
  sandboxRun,
  install,
  orgId,
  flush,
}: ISandboxHeader) => {
  if (!flush) {
    return (
      <LegacySandboxHeader
        workflow={workflow}
        stepId={stepId}
        sandboxRun={sandboxRun}
        install={install}
        orgId={orgId}
      />
    )
  }

  return (
    <header className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center gap-4 justify-between w-full">
        <div className="flex flex-col gap-1">
          <BackLink className="mb-4" />
          <span className="flex items-center gap-2">
            <CloudPlatform
              platform={install.cloud_platform as TCloudPlatform}
              variant="subtext"
              displayVariant="icon-only"
            />
            <Text variant="base" weight="strong">
              Sandbox {sandboxRun?.run_type}
            </Text>
          </span>
          <ID>{sandboxRun?.id}</ID>
          <Time
            time={sandboxRun?.created_at}
            format="relative"
            variant="subtext"
            theme="info"
          />
        </div>

        <div className="flex items-center gap-4">
          <SandboxRunSwitcher sandboxRunId={sandboxRun?.id} />
          <ManageRunDropdown
            workflow={workflow}
            variant="primary"
          />
        </div>
      </div>

      <Card>
        <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
          <LabeledStatus
            label="Status"
            statusProps={{
              status: sandboxRun?.status_v2?.status,
            }}
            tooltipProps={{
              tipContentClassName: 'w-fit',
              tipContent: (
                <Text nowrap variant="subtext">
                  {toSentenceCase(
                    sandboxRun?.status_v2?.status_human_description
                  )}
                </Text>
              ),
              position: 'bottom',
            }}
          />
          <LabeledValue label="Duration">
            <Duration
              variant="subtext"
              beginTime={sandboxRun?.created_at}
              endTime={sandboxRun?.updated_at}
            />
          </LabeledValue>
          <LabeledValue label="Install">
            <Text variant="subtext">
              <Link href={`/${orgId}/installs/${install?.id}`}>
                {install?.name}
              </Link>
            </Text>
          </LabeledValue>
          <LabeledValue label="Config">
            <SandboxConfigContextTooltip
              appConfigId={install?.app_config_id}
              appId={install?.app_id}
            >
              <Text variant="subtext">
                <Link href={`/${orgId}/apps/${install?.app_id}`}>
                  {install?.app?.name} sandbox
                </Link>
              </Text>
            </SandboxConfigContextTooltip>
          </LabeledValue>
          {sandboxRun?.runner_jobs?.at(0)?.install_role_usage?.role_name ? (
            <LabeledValue label="Execution role">
              <Text variant="subtext" family="mono" className="text-xs">
                <Link href={`/${orgId}/installs/${install?.id}/roles?panel=${sandboxRun.runner_jobs.at(0).install_role_usage.install_role_id}`}>
                  {sandboxRun.runner_jobs.at(0).install_role_usage.role_name}
                </Link>
              </Text>
            </LabeledValue>
          ) : null}
        </div>
      </Card>

      {sandboxRun?.install_workflow_id ? (
        <Button
          href={`/${orgId}/installs/${install?.id}/workflows/${workflow?.id}?panel=${stepId}`}
        >
          View workflow
          <Icon variant="CaretRightIcon" />
        </Button>
      ) : null}
    </header>
  )
}

const LegacySandboxHeader = ({
  workflow,
  stepId,
  sandboxRun,
  install,
  orgId,
}: Omit<ISandboxHeader, 'flush'>) => {
  return (
    <header className="flex flex-col p-6 border-b gap-4">
      <div className="flex items-center justify-between">
        <BackLink />
        <div className="flex items-center gap-6">
          <LabeledStatus
            label="Status"
            statusProps={{
              status: sandboxRun?.status_v2?.status,
            }}
            tooltipProps={{
              tipContentClassName: 'w-fit',
              tipContent: (
                <Text nowrap variant="subtext">
                  {toSentenceCase(
                    sandboxRun?.status_v2?.status_human_description
                  )}
                </Text>
              ),
              position: 'bottom',
            }}
          />
          <LabeledValue label="Install">
            <Text variant="subtext">
              <Link href={`/${orgId}/installs/${install?.id}`}>
                {install?.name}
              </Link>
            </Text>
          </LabeledValue>
          <LabeledValue label="Config">
            <SandboxConfigContextTooltip
              appConfigId={install?.app_config_id}
              appId={install?.app_id}
            >
              <Text variant="subtext">
                <Link href={`/${orgId}/apps/${install?.app_id}`}>
                  {install?.app?.name} sandbox
                </Link>
              </Text>
            </SandboxConfigContextTooltip>
          </LabeledValue>
        </div>
      </div>

      <div className="flex flex-col gap-1">
        <span className="flex items-center gap-2">
          <CloudPlatform
            platform={install.cloud_platform as TCloudPlatform}
            variant="subtext"
            displayVariant="icon-only"
          />
          <Text variant="base" weight="strong">
            Sandbox {sandboxRun?.run_type}
          </Text>
        </span>
        <ID>{sandboxRun?.id}</ID>
        <div className="flex flex-wrap gap-x-8 gap-y-1 items-center mt-1">
          <Text theme="info" flex className="gap-1">
            <Icon variant="CalendarBlankIcon" />
            <Time variant="subtext" time={sandboxRun?.created_at} />
          </Text>
          <Text theme="info" flex className="gap-1">
            <Icon variant="TimerIcon" />
            <Duration
              variant="subtext"
              beginTime={sandboxRun?.created_at}
              endTime={sandboxRun?.updated_at}
            />
          </Text>
          {sandboxRun?.runner_jobs?.at(0)?.install_role_usage?.role_name ? (
            <Text theme="info" flex className="gap-1">
              <Icon variant="FileLockIcon" />
              <Text variant="subtext">{sandboxRun.runner_jobs.at(0).install_role_usage.role_name}</Text>
            </Text>
          ) : null}
        </div>
      </div>

      <div className="flex items-center justify-between">
        {sandboxRun?.install_workflow_id ? (
          <Button
            href={`/${orgId}/installs/${install?.id}/workflows/${workflow?.id}?panel=${stepId}`}
          >
            View workflow
            <Icon variant="CaretRightIcon" />
          </Button>
        ) : (
          <div />
        )}
        <div className="flex gap-4 items-center">
          <SandboxRunSwitcher sandboxRunId={sandboxRun?.id} />
          <ManageRunDropdown
            workflow={workflow}
            variant="primary"
          />
        </div>
      </div>
    </header>
  )
}
