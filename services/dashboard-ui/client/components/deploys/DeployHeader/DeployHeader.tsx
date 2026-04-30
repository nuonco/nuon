import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CommitDetails } from '@/components/common/CommitDetails'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ComponentType } from '@/components/components/ComponentType'
import { ComponentConfigContextTooltip } from '@/components/components/ComponentConfigContextTooltip'
import type { TComponent, TDeploy, TInstall, TWorkflow } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { DeploySwitcher } from '@/components/deploys/DeploySwitcher'
import { OCIArtifactCard } from '@/components/deploys/OCIArtifactCard'
import { ManagementDropdown } from '@/components/deploys/management/ManagementDropdown'

interface IDeployHeader {
  children?: React.ReactNode
  component: TComponent
  workflow: TWorkflow
  stepId: string
  deploy: TDeploy
  install: TInstall
}

export const DeployHeader = ({
  children,
  component,
  workflow,
  stepId,
  deploy,
  install,
}: IDeployHeader) => {
  return (
    <header className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center gap-4 justify-between w-full">
        <div className="flex flex-col gap-1">
          <BackLink className="mb-4" />
          <span className="flex items-center gap-2">
            <ComponentType type={component?.type} displayVariant="icon-only" />
            <Text variant="base" weight="strong">
              {deploy?.component_name}{' '}
              {deploy?.install_deploy_type === 'teardown'
                ? 'teardown'
                : 'deploy'}
            </Text>
          </span>
          <span className="flex items-center gap-4">
            <ID>{deploy?.id}</ID>
            <Text
              className="!flex gap-2"
              variant="subtext"
              theme="neutral"
              family="mono"
            >
              Build:
              <ID>
                <Link
                  href={`/${deploy?.org_id}/apps/${install?.app_id}/components/${deploy?.component_id}/builds/${deploy?.build_id}`}
                >
                  {deploy?.build_id}
                </Link>
              </ID>
            </Text>
          </span>
          <Time
            time={deploy?.created_at}
            format="relative"
            variant="subtext"
            theme="info"
          />
        </div>

        <div className="flex items-center gap-4">
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

      <Card>
        <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
          <LabeledStatus
            label="Status"
            statusProps={{
              status: deploy?.status_v2?.status,
            }}
            tooltipProps={{
              tipContentClassName: 'w-fit',
              tipContent: (
                <Text nowrap variant="subtext">
                  {toSentenceCase(deploy?.status_v2?.status_human_description)}
                </Text>
              ),
              position: 'bottom',
            }}
          />
          <LabeledValue label="Duration">
            <Duration
              variant="subtext"
              beginTime={deploy?.created_at}
              endTime={deploy?.updated_at}
            />
          </LabeledValue>
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
            <LabeledValue label="OCI artifact">
              <OCIArtifactCard ociArtifact={deploy?.oci_artifact}>
                <Text
                  variant="subtext"
                  as="div"
                  className="truncate max-w-[80px]"
                  theme="neutral"
                >
                  {deploy?.oci_artifact?.tag}
                </Text>
              </OCIArtifactCard>
            </LabeledValue>
          ) : null}
          {deploy?.runner_jobs?.at(0)?.install_role_usage?.role_name ? (
            <LabeledValue label="Execution role">
              <Text variant="subtext" family="mono" className="text-xs">
                <Link href={`/${deploy?.org_id}/installs/${deploy?.install_id}/roles?panel=${deploy.runner_jobs.at(0).install_role_usage.install_role_id}`}>
                  {deploy.runner_jobs.at(0).install_role_usage.role_name}
                </Link>
              </Text>
            </LabeledValue>
          ) : null}
        </div>
      </Card>

      {deploy?.install_workflow_id ? (
        <Button
          href={`/${deploy?.org_id}/installs/${deploy?.install_id}/workflows/${workflow?.id}?panel=${stepId}`}
        >
          View workflow
          <Icon variant="CaretRightIcon" />
        </Button>
      ) : null}

      {children}
    </header>
  )
}
