'use client'

import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { Duration } from '@/components/common/Duration'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSandboxRun } from '@/hooks/use-sandbox-run'
import type { TCloudPlatform, TWorkflow } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { SandboxRunSwitcher } from './SandboxRunSwitcher'
import { ManageRunDropdown } from '@/components/sandbox/management/ManageRunDropdown'
import { SandboxConfigContextTooltip } from '@/components/sandbox/SandboxConfigContextTooltip'

interface ISandboxHeader {
  workflow: TWorkflow
  stepId: string
}

export const SandboxHeader = ({
  workflow,
  stepId,
}: ISandboxHeader) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { sandboxRun } = useSandboxRun()

  return (
    <header className="flex p-6 border-b justify-between w-full">
      <HeadingGroup>
        <BackLink className="mb-6" />
        <div className="flex flex-col gap-1">
          <span className="flex items-cenert gap-2">
            <CloudPlatform
              platform={install.cloud_platform as TCloudPlatform}
              variant="subtext"
              displayVariant="icon-only"
            />
            <Text
              className="inline-flex items-center gap-4"
              variant="h3"
              weight="strong"
            >
              Sandbox {sandboxRun?.run_type}
            </Text>
          </span>
          <ID>{sandboxRun?.id}</ID>
        </div>

        <div className="flex gap-8 items-center justify-start my-2">
          <Text theme="info" className="!flex items-center gap-1">
            <Icon variant="CalendarBlankIcon" />
            <Time variant="subtext" time={sandboxRun?.created_at} />
          </Text>
          <Text theme="info" className="!flex items-center gap-1">
            <Icon variant="TimerIcon" />
            <Duration
              variant="subtext"
              beginTime={sandboxRun?.created_at}
              endTime={sandboxRun?.updated_at}
            />
          </Text>
        </div>

        {sandboxRun?.install_workflow_id ? (
          <Button
            href={`/${org?.id}/installs/${install?.id}/workflows/${workflow?.id}?panel=${stepId}`}
          >
            View workflow
            <Icon variant="CaretRightIcon" />
          </Button>
        ) : null}
      </HeadingGroup>

      <div className="flex flex-col gap-6">
        <div className="flex items-start justify-start gap-6">
          <LabeledStatus
            label="Status"
            statusProps={{
              status: sandboxRun?.status_v2?.status,
            }}
            tooltipProps={{
              tipContentClassName: 'w-fit',
              tipContent: (
                <Text className="!text-nowrap" variant="subtext">
                  {toSentenceCase(
                    sandboxRun?.status_v2?.status_human_description
                  )}
                </Text>
              ),
              position: 'left',
            }}
          />

          <LabeledValue label="Install">
            <Text variant="subtext">
              <Link href={`/${org?.id}/installs/${install?.id}`}>
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
                <Link href={`/${org?.id}/apps/${install?.app_id}`}>
                  {install?.app?.name} sandbox
                </Link>
              </Text>
            </SandboxConfigContextTooltip>
          </LabeledValue>
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
