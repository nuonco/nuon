import { Outlet, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { ComponentType } from '@/components/components/ComponentType'
import { DriftedBanner } from '@/components/install-components/DriftedBanner'
import {
  StuckHelmReleaseBanner,
  stuckHelmReleaseStatus,
} from '@/components/install-components/StuckHelmReleaseBanner'
import { ManagementDropdown } from '@/components/install-components/management/ManagementDropdown'
import { RemovedFromAppConfigBanner } from '@/components/installs/RemovedFromAppConfig'
import { AdminDashboardLink } from '@/components/admin/AdminDashboardLink'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useInstall } from '@/hooks/use-install'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useOrg } from '@/hooks/use-org'
import { getComponentBuilds, getInstallComponent } from '@/lib'
import type { TNavLink } from '@/types'
import { groupComponentOverrideInputs } from '@/utils/install-utils'
import type { TInstallComponentOutletContext } from './install-component-tabs/types'
import { CustomerManagedSnapshotComponentDetail } from '@/components/customer-managed-support/SnapshotComponentDetail'
import { isCustomerManagedInstall } from '@/utils/install-utils'

export const InstallComponentLayout = () => {
  const { install } = useInstall()

  return isCustomerManagedInstall(install) ? (
    <CustomerManagedSnapshotComponentDetail />
  ) : (
    <ConnectedInstallComponentLayout />
  )
}

const ConnectedInstallComponentLayout = () => {
  const { componentId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: installComponent, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-component', org?.id, install?.id, componentId],
    queryFn: () =>
      getInstallComponent({
        orgId: org.id,
        installId: install.id,
        componentId: componentId!,
      }),
    enabled: !!org?.id && !!install?.id && !!componentId,
  })

  const { appConfig, isLoading: isLoadingConfig } = useInstallAppConfig()

  const { data: latestBuilds } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['component-builds', org?.id, componentId, 0],
    queryFn: () =>
      getComponentBuilds({
        orgId: org.id,
        componentId: componentId!,
        limit: 10,
        offset: 0,
      }),
    enabled: !!org?.id && !!componentId,
  })

  const component = installComponent?.component
  const latestDeploy = installComponent?.install_deploys?.[0]
  const stuckHelmStatus = stuckHelmReleaseStatus(latestDeploy)
  const config = appConfig?.component_config_connections?.find(
    (c) => c.component_id === componentId
  )
  const removed = !isLoadingConfig && !!appConfig && !config

  const installValues = install?.install_inputs?.at(0)?.values
  const overrideCard = groupComponentOverrideInputs(
    appConfig?.input?.inputs ?? []
  ).find((c) => c.component === component?.name)

  const isToggleable = config?.toggleable === true
  const isDisabled =
    installComponent?.status === 'disabled' ||
    installComponent?.enabled === false

  const dependentIds =
    appConfig?.component_config_connections
      ?.filter((c) => c.component_dependency_ids?.includes(componentId!))
      .map((c) => c.component_id!)
      .filter(Boolean) ?? []

  const basePath = `/${org?.id}/installs/${install?.id}/components/${componentId}`
  const hasState =
    component?.type === 'terraform_module' || component?.type === 'pulumi'

  const tabs: TNavLink[] = [
    { path: '/', text: 'Overview' },
    { path: '/deploys', text: 'Deploy history' },
    { path: '/config', text: 'Config' },
  ]
  if (hasState) tabs.push({ path: '/state', text: 'State' })

  const context: TInstallComponentOutletContext = {
    appConfig,
    config,
    dependentIds,
    installComponent,
    installValues,
    isDisabled,
    isLoading,
    isLoadingConfig,
    isToggleable,
    latestBuilds: latestBuilds?.data,
    latestDeploy,
    overrideCard,
    removed,
  }

  return (
    <>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/components`,
            text: 'Components',
          },
          { path: basePath, text: component?.name },
        ]}
      />

      <DetailPage
        header={
          <DetailHeader
            icon={
              <ComponentType
                type={component?.type}
                displayVariant="icon-only"
                colorVariant="color"
                iconSize="24"
              />
            }
            title={component?.name}
            loading={isLoading}
            loadingWidth={20}
            status={
              isToggleable ? (
                <Badge size="sm" theme={isDisabled ? 'neutral' : 'success'}>
                  {isDisabled ? 'Disabled' : 'Enabled'}
                </Badge>
              ) : null
            }
            id={component?.id}
            identity={
              <AdminDashboardLink
                path={`/queues?owner_id=${installComponent?.id}`}
                label="Admin panel"
              />
            }
            actions={
              component ? (
                <ManagementDropdown
                  component={component}
                  componentConfig={config}
                  currentBuildId={latestDeploy?.build_id}
                  currentDeployStatus={
                    isDisabled ? 'disabled' : latestDeploy?.status_v2?.status
                  }
                  installComponent={installComponent}
                  isConfigLoading={isLoadingConfig}
                  removed={removed}
                />
              ) : null
            }
          />
        }
        banners={
          <>
            {removed ? <RemovedFromAppConfigBanner kind="component" /> : null}
            {installComponent?.drifted_object ? (
              <DriftedBanner drifted={installComponent.drifted_object} />
            ) : null}
            {component && stuckHelmStatus ? (
              <StuckHelmReleaseBanner
                component={component}
                status={stuckHelmStatus}
              />
            ) : null}
          </>
        }
        tabNav={{ basePath, tabs }}
      >
        <Outlet context={context} />
      </DetailPage>
    </>
  )
}
