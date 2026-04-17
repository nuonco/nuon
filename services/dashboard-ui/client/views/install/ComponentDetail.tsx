import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { ComponentType } from '@/components/components/ComponentType'
import {
  ComponentConfigCard,
  ComponentConfigCardSkeleton,
} from '@/components/components/ComponentConfigCard'
import { DeployTimeline } from '@/components/deploys/DeployTimeline'
import { DriftedBanner } from '@/components/install-components/DriftedBanner'
import { InstallComponentDependencies } from '@/components/install-components/InstallComponentDependencies'
import { ManagementDropdown } from '@/components/install-components/management/ManagementDropdown'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { TerraformWorkspaceCard } from '@/components/terraform-workspace/TerraformWorkspaceCard'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getAppConfig, getInstallComponent } from '@/lib'

export const InstallComponentDetail = () => {
  const { componentId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: installComponent, isLoading } = useQuery({
    queryKey: ['install-component', org?.id, install?.id, componentId],
    queryFn: () =>
      getInstallComponent({
        orgId: org.id,
        installId: install.id,
        componentId: componentId!,
      }),
    enabled: !!org?.id && !!install?.id && !!componentId,
  })

  const { data: appConfig, isLoading: isLoadingConfig } = useQuery({
    queryKey: [
      'app-config',
      org?.id,
      install?.app_id,
      install?.app_config_id,
      'recurse',
    ],
    queryFn: () =>
      getAppConfig({
        orgId: org.id,
        appId: install.app_id,
        appConfigId: install.app_config_id,
        recurse: true,
      }),
    enabled: !!org?.id && !!install?.app_config_id,
  })

  const component = installComponent?.component
  const latestDeploy = installComponent?.install_deploys?.[0]
  const config = appConfig?.component_config_connections?.find(
    (c) => c.component_id === componentId
  )

  return (
    <PageSection>
      <PageTitle
        title={`${component?.name ?? 'Component'} | ${install?.name}`}
      />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/components`,
            text: 'Components',
          },
          {
            path: `/${org?.id}/installs/${install?.id}/components/${componentId}`,
            text: component?.name,
          },
        ]}
      />

      <div className="flex items-start justify-between">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <span className="flex items-center gap-2">
            <ComponentType
              type={component?.type}
              displayVariant="icon-only"
              colorVariant="color"
              iconSize="24"
            />
            <Text variant="base" weight="strong">
              {component?.name}
            </Text>
          </span>
          {component?.id ? <ID>{component.id}</ID> : null}
        </HeadingGroup>

        {component && (
          <div className="flex items-center gap-4">
            <AdminDashboardLink
              path={`/queues?owner_id=${installComponent?.id}`}
              label="View in admin panel"
            />
            <ManagementDropdown
              component={component}
              currentBuildId={latestDeploy?.build_id}
              currentDeployStatus={latestDeploy?.status_v2?.status}
              installComponent={installComponent}
            />
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto gap-6">
        <div className="md:col-span-8 flex flex-col gap-6">
          {installComponent?.drifted_object ? (
            <DriftedBanner drifted={installComponent.drifted_object} />
          ) : null}

          {config?.component_dependency_ids?.length ? (
            <Card>
              <Text weight="strong">Dependencies</Text>
              <InstallComponentDependencies
                deps={config.component_dependency_ids}
                variant="inline"
              />
            </Card>
          ) : null}

          {isLoadingConfig ? (
            <ComponentConfigCardSkeleton />
          ) : config ? (
            <ComponentConfigCard config={config} />
          ) : (
            <EmptyState
              variant="table"
              emptyTitle="No configuration"
              emptyMessage="This component has no configuration yet."
            />
          )}

          {component?.type === 'terraform_module' || component?.type === 'pulumi' ? (
            <TerraformWorkspaceCard
              workspaceId={installComponent?.terraform_workspace?.id}
              componentType={component?.type}
            />
          ) : null}
        </div>

        <div className="md:col-span-4 flex flex-col gap-4">
          <Text variant="base" weight="strong">
            Deploy history
          </Text>
          {component ? (
            <DeployTimeline
              componentId={componentId!}
              componentName={component.name}
              shouldPoll
            />
          ) : null}
        </div>
      </div>

    </PageSection>
  )
}
