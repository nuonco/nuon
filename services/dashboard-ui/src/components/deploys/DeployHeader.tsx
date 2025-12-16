'use client'

import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { CommitDetails } from '@/components/common/CommitDetails'
import { Duration } from '@/components/common/Duration'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ComponentType } from '@/components/components/ComponentType'
import { ComponentConfigContextTooltip } from '@/components/components/ComponentConfigContextTooltip'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'
import type { TComponent, TWorkflow } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { DeploySwitcher } from '@/components/deploys/DeploySwitcher'
import { OCIArtifactCard } from '@/components/deploys/OCIArtifactCard'
import { ManagementDropdown } from '@/components/deploys/management/ManagementDropdown'

export const DeployHeader = ({
  component,
  workflow,
  stepId,
}: {
  component: TComponent
  workflow: TWorkflow
  stepId: string
}) => {
  const { deploy } = useDeploy()
  const { install } = useInstall()
  return (
    <header className="p-6 border-b flex justify-between">
      <HeadingGroup>
        <BackLink className="mb-6" />
        <div className="flex flex-col gap-1">
          <span className="flex items-cenert gap-2">
            <ComponentType type={component?.type} displayVariant="icon-only" />
            <Text variant="base" weight="strong">
              {deploy?.component_name} deploy
            </Text>
          </span>
          <ID>{deploy?.id}</ID>
        </div>
        <div className="flex gap-8 items-center justify-start my-2">
          <Text theme="info" className="!flex items-center gap-1">
            <Icon variant="CalendarBlankIcon" />
            <Time variant="subtext" time={deploy?.created_at} />
          </Text>
          <Text theme="info" className="!flex items-center gap-1">
            <Icon variant="TimerIcon" />
            <Duration
              variant="subtext"
              beginTime={deploy?.created_at}
              endTime={deploy?.updated_at}
            />
          </Text>
        </div>
        {deploy?.install_workflow_id ? (
          <Button
            href={`/${deploy?.org_id}/installs/${deploy?.install_id}/workflows/${workflow?.id}?panel=${stepId}`}
          >
            View workflow
            <Icon variant="CaretRightIcon" />
          </Button>
        ) : null}
      </HeadingGroup>

      <div className="flex flex-col gap-6">
        <div className="flex gap-6 items-start justify-start">
          <LabeledStatus
            label="Status"
            statusProps={{
              status: deploy?.status_v2?.status,
            }}
            tooltipProps={{
              tipContentClassName: 'w-fit',
              tipContent: (
                <Text className="!text-nowrap" variant="subtext">
                  {toSentenceCase(deploy?.status_v2?.status_human_description)}
                </Text>
              ),
              position: 'left',
            }}
          />

          <LabeledValue label="Install">
            <Text variant="subtext">
              <Link href={`/${deploy?.org_id}/installs/${deploy?.install_id}`}>
                {install?.name}
              </Link>
            </Text>
          </LabeledValue>
          <LabeledValue label="Config">
            <ComponentConfigContextTooltip
              componentId={component?.id}
              configId={deploy?.component_build?.component_config_connection_id}
              appId={component?.app_id}
            >
              <Text variant="subtext">
                <Link
                  href={`/${deploy.org_id}/installs/${install.id}/components/${component?.id}`}
                >
                  {component?.name}
                </Link>
              </Text>
            </ComponentConfigContextTooltip>
          </LabeledValue>
          {deploy?.component_build?.vcs_connection_commit ? (
            <LabeledValue label="Commit">
              <CommitDetails
                commit={deploy?.component_build?.vcs_connection_commit}
              />
            </LabeledValue>
          ) : null}

          {deploy?.oci_artifact ? (
            <LabeledValue label="OCI Artifact">
              <OCIArtifactCard ociArtifact={deploy?.oci_artifact}>
                <Text variant="subtext">{deploy?.oci_artifact?.tag}</Text>
              </OCIArtifactCard>
            </LabeledValue>
          ) : null}

          <DeploySwitcher
            componentId={deploy?.component_id}
            deployId={deploy?.id}
          />

          <ManagementDropdown
            component={component}
            currentBuildId={deploy?.build_id}
            workflow={workflow}
          />
        </div>
      </div>
    </header>
  )
}
