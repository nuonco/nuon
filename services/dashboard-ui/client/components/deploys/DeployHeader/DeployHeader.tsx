import { Button } from '@/components/common/Button'
import { CommitDetails } from '@/components/common/CommitDetails'
import { Duration } from '@/components/common/Duration'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { ComponentType } from '@/components/components/ComponentType'
import { ComponentConfigContextTooltip } from '@/components/components/ComponentConfigContextTooltip'
import { DetailHeader } from '@/components/layout/DetailHeader'
import type { TComponent, TDeploy, TInstall, TWorkflow } from '@/types'
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
  const executionRole = deploy?.runner_jobs?.at(0)?.install_role_usage

  return (
    <DetailHeader
      icon={<ComponentType type={component?.type} displayVariant="icon-only" />}
      title={`${deploy?.component_name} ${
        deploy?.install_deploy_type === 'teardown' ? 'teardown' : 'deploy'
      }`}
      id={deploy?.id}
      identity={
        <>
          <Text
            className="!flex gap-2"
            variant="subtext"
            theme="neutral"
            family="mono"
          >
            Build:
            <ID>
              <Link
                href={`/${install?.org_id}/apps/${install?.app_id}/components/${deploy?.component_id}/builds/${deploy?.build_id}`}
                variant="inline"
              >
                {deploy?.build_id}
              </Link>
            </ID>
          </Text>
          <Time
            time={deploy?.created_at}
            format="relative"
            variant="subtext"
            theme="info"
          />
        </>
      }
      actions={
        <>
          <DeploySwitcher
            componentId={deploy?.component_id}
            deployId={deploy?.id}
          />
          <ManagementDropdown
            component={component}
            currentBuildId={deploy?.build_id}
            workflow={workflow}
          />
        </>
      }
      metadata={
        <>
          <LabeledStatus
            label="Status"
            statusProps={{
              status: deploy?.status_v2?.status,
            }}
            tooltipProps={{
              tipContentClassName: 'w-fit',
              tipContent: (
                <Text nowrap variant="subtext">
                  {deploy?.status_v2?.status_human_description}
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
            <Link href={`/${install?.org_id}/installs/${install?.id}`}>
              {install?.name}
            </Link>
          </LabeledValue>
          <LabeledValue label="Config">
            <ComponentConfigContextTooltip
              componentId={component?.id}
              configId={deploy?.component_build?.component_config_connection_id}
              appId={component?.app_id}
            >
              <Link
                href={`/${install?.org_id}/installs/${install?.id}/components/${component?.id}`}
              >
                {component?.name}
              </Link>
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
          {executionRole?.role_name ? (
            <LabeledValue label="Execution role">
              <Text variant="subtext" family="mono" className="text-xs">
                <Link
                  href={`/${install?.org_id}/installs/${install?.id}/roles?panel=${executionRole.install_role_id}`}
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
      {deploy?.install_workflow_id ? (
        <Button
          href={`/${install?.org_id}/installs/${install?.id}/workflows/${workflow?.id}?panel=${stepId}`}
        >
          View workflow
        </Button>
      ) : null}

      {children}
    </DetailHeader>
  )
}
