import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { ComponentDependencyGraphButton } from '@/components/components/ComponentDependencyGraph'
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
import { Panel } from '@/components/surfaces/Panel'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getAppConfig, getInstallComponent } from '@/lib'

export const InstallComponentDetail = () => {
  const { componentId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addPanel } = useSurfaces()

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

  const dependentIds = appConfig?.component_config_connections
    ?.filter((c) => c.component_dependency_ids?.includes(componentId!))
    .map((c) => c.component_id!)
    .filter(Boolean) ?? []

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

      <div className="@container flex flex-col flex-auto gap-6">
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
            <AdminDashboardLink
              path={`/queues?owner_id=${installComponent?.id}`}
              label="Admin panel"
            />
          </HeadingGroup>

          {component && (
            <div className="flex items-center gap-4">
              <div className="@5xl:hidden">
                <Button
                  variant="secondary"
                  onClick={() =>
                    addPanel(
                      <Panel heading="Deploy history">
                        <DeployTimeline
                          componentId={componentId!}
                          componentName={component.name}
                          shouldPoll
                        />
                      </Panel>
                    )
                  }
                >
                  <Icon variant="ClockCounterClockwiseIcon" size={16} />
                  Deploy history
                </Button>
              </div>
              <ManagementDropdown
                component={component}
                currentBuildId={latestDeploy?.build_id}
                currentDeployStatus={latestDeploy?.status_v2?.status}
                installComponent={installComponent}
              />
            </div>
          )}
        </div>

        <div className="grid grid-cols-1 @5xl:grid-cols-12 gap-6">
          <div className="@5xl:col-span-8 flex flex-col gap-6">
            {installComponent?.drifted_object ? (
              <DriftedBanner drifted={installComponent.drifted_object} />
            ) : null}

            {isLoadingConfig ? (
              <ComponentConfigCardSkeleton />
            ) : config ? (
              <ComponentConfigCard
                config={config}
                headerActions={
                  appConfig && componentId && component?.name ? (
                    <ComponentDependencyGraphButton
                      componentId={componentId}
                      componentName={component.name}
                      componentType={component.type}
                      appConfig={appConfig}
                      basePath={`/${org?.id}/installs/${install?.id}/components`}
                      size="sm"
                    />
                  ) : null
                }
                footer={
                  (config.component_dependency_ids?.length || dependentIds.length > 0) ? (
                    <>
                      {config.component_dependency_ids?.length ? (
                        <div className="flex flex-col gap-2">
                          <Text variant="body" weight="strong" level={5}>Dependencies</Text>
                          <InstallComponentDependencies
                            deps={config.component_dependency_ids}
                            variant="inline"
                          />
                        </div>
                      ) : null}
                      {dependentIds.length > 0 ? (
                        <div className="flex flex-col gap-2">
                          <Text variant="body" weight="strong" level={5}>Dependents</Text>
                          <InstallComponentDependencies
                            deps={dependentIds}
                            variant="inline"
                            tooltipTitle="More dependents"
                          />
                        </div>
                      ) : null}
                    </>
                  ) : undefined
                }
              />
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

          <div className="hidden @5xl:flex flex-col @5xl:col-span-4 gap-4">
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
      </div>

    </PageSection>
  )
}
