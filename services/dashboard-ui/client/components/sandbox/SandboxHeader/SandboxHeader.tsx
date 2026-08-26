import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { DetailHeader } from '@/components/layout/DetailHeader'
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
}

export const SandboxHeader = ({
  workflow,
  stepId,
  sandboxRun,
  install,
  orgId,
}: ISandboxHeader) => {
  const executionRole = sandboxRun?.runner_jobs?.at(0)?.install_role_usage

  return (
    <DetailHeader
      icon={
        <CloudPlatform
          platform={install?.cloud_platform as TCloudPlatform}
          variant="subtext"
          displayVariant="icon-only"
        />
      }
      title={`Sandbox ${sandboxRun?.run_type}`}
      id={sandboxRun?.id}
      identity={
        <Time
          time={sandboxRun?.created_at}
          format="relative"
          variant="subtext"
          theme="info"
        />
      }
      actions={
        <>
          <SandboxRunSwitcher sandboxRunId={sandboxRun?.id} />
          <ManageRunDropdown workflow={workflow} variant="primary" />
        </>
      }
      metadata={
        <>
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
            <Link href={`/${orgId}/installs/${install?.id}`}>
              {install?.name}
            </Link>
          </LabeledValue>
          <LabeledValue label="Config">
            <SandboxConfigContextTooltip
              appConfigId={install?.app_config_id}
              appId={install?.app_id}
            >
              <Link href={`/${orgId}/apps/${install?.app_id}`}>
                {install?.app?.name} sandbox
              </Link>
            </SandboxConfigContextTooltip>
          </LabeledValue>
          {executionRole?.role_name ? (
            <LabeledValue label="Execution role">
              <Text variant="subtext" family="mono" className="text-xs">
                <Link
                  href={`/${orgId}/installs/${install?.id}/roles?panel=${executionRole.install_role_id}`}
                  variant="inline"
                >
                  {executionRole.role_name}
                </Link>
              </Text>
            </LabeledValue>
          ) : null}
        </>
      }
    >
      {sandboxRun?.install_workflow_id ? (
        <Button
          href={`/${orgId}/installs/${install?.id}/workflows/${workflow?.id}?panel=${stepId}`}
        >
          View workflow
        </Button>
      ) : null}
    </DetailHeader>
  )
}
