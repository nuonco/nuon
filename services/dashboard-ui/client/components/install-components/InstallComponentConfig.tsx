import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Text } from '@/components/common/Text'
import { ComponentConfigCard } from '@/components/components/ComponentConfigCard'
import { ComponentDependencyGraphButton } from '@/components/components/ComponentDependencyGraph'
import { InstallComponentDependencies } from '@/components/install-components/InstallComponentDependencies'
import { ComponentOverrideCard } from '@/components/install-overrides/ComponentOverrideCard'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { useInstallAppConfig } from '@/hooks/use-install-app-config'
import { useInstallPage } from '@/hooks/use-install-path'
import { useOrg } from '@/hooks/use-org'
import { getComponentBuilds, getInstallComponent } from '@/lib'
import type {
  TAppConfig,
  TBuild,
  TComponentConfig,
  TDeploy,
  TInstallComponent,
} from '@/types'
import {
  groupComponentOverrideInputs,
  type TComponentOverrideCard,
} from '@/utils/install-utils'

export const InstallComponentConfigBody = ({
  appConfig,
  componentId,
  config,
  dependentIds,
  installComponent,
  installValues,
  isLoadingConfig,
  latestBuilds,
  latestDeploy,
  overrideCard,
}: {
  appConfig?: TAppConfig
  componentId?: string
  config?: TComponentConfig
  dependentIds: string[]
  installComponent?: TInstallComponent
  installValues?: Record<string, string>
  isLoadingConfig: boolean
  latestBuilds?: TBuild[]
  latestDeploy?: TDeploy
  overrideCard?: TComponentOverrideCard<{ name?: string; index?: number }>
}) => {
  const { href } = useInstallPage()
  const component = installComponent?.component
  const latestResolvedBuild = latestBuilds?.find((b) => !!b.source_digest)
  const deployCommit = latestDeploy?.component_build?.vcs_connection_commit
  const latestCommit = deployCommit
    ? {
        status: latestDeploy?.status_v2?.status,
        href: href(`/components/${componentId}/deploys/${latestDeploy?.id}`),
        message: deployCommit.message?.split('\n')[0],
        author: deployCommit.author_name,
        avatarUrl: deployCommit.author_avatar_url,
        sha: deployCommit.sha,
        createdAt: deployCommit.created_at,
      }
    : undefined

  const hasConfigOverride =
    !!overrideCard?.configInput?.name &&
    !!installValues?.[overrideCard.configInput.name]

  return (
    <>
      {isLoadingConfig ? (
        <ComponentConfigCard loading />
      ) : config ? (
        <ComponentConfigCard
          config={config}
          latestBuild={latestResolvedBuild}
          latestCommit={latestCommit}
          headerActions={
            appConfig && componentId && component?.name ? (
              <ComponentDependencyGraphButton
                componentId={componentId}
                componentName={component.name}
                componentType={component.type}
                appConfig={appConfig}
                basePath={href('/components')}
                size="sm"
              />
            ) : null
          }
          footer={
            config.component_dependency_ids?.length ||
            dependentIds.length > 0 ? (
              <>
                {config.component_dependency_ids?.length ? (
                  <div className="flex flex-col gap-2">
                    <Text variant="body" weight="strong" level={5}>
                      Dependencies
                    </Text>
                    <InstallComponentDependencies
                      deps={config.component_dependency_ids}
                      variant="inline"
                    />
                  </div>
                ) : null}
                {dependentIds.length > 0 ? (
                  <div className="flex flex-col gap-2">
                    <Text variant="body" weight="strong" level={5}>
                      Dependents
                    </Text>
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

      {overrideCard && hasConfigOverride ? (
        <div className="flex flex-col gap-4">
          <SectionHeader
            title="Install overrides"
            description="Configuration applied to this component on this install."
          />
          <ComponentOverrideCard
            card={overrideCard}
            values={installValues}
            readOnly
            showEnabled={false}
          />
        </div>
      ) : null}
    </>
  )
}

export const InstallComponentConfig = ({
  componentId,
}: {
  componentId: string
}) => {
  const { org } = useOrg()
  const { install } = useInstallPage()
  const { appConfig, isLoading: isLoadingConfig } = useInstallAppConfig()

  const { data: installComponent } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-component', org?.id, install?.id, componentId],
    queryFn: () =>
      getInstallComponent({
        orgId: org.id,
        installId: install.id,
        componentId,
      }),
    enabled: !!org?.id && !!install?.id && !!componentId,
  })

  const { data: latestBuilds } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['component-builds', org?.id, componentId, 0],
    queryFn: () =>
      getComponentBuilds({
        orgId: org.id,
        componentId,
        limit: 10,
        offset: 0,
      }),
    enabled: !!org?.id && !!componentId,
  })

  const component = installComponent?.component
  const config = appConfig?.component_config_connections?.find(
    (connection) => connection.component_id === componentId
  )
  const dependentIds =
    appConfig?.component_config_connections
      ?.filter((connection) =>
        connection.component_dependency_ids?.includes(componentId)
      )
      .map((connection) => connection.component_id!)
      .filter(Boolean) ?? []
  const overrideCard = groupComponentOverrideInputs(
    appConfig?.input?.inputs ?? []
  ).find((card) => card.component === component?.name)

  return (
    <InstallComponentConfigBody
      appConfig={appConfig}
      componentId={componentId}
      config={config}
      dependentIds={dependentIds}
      installComponent={installComponent}
      installValues={install?.install_inputs?.at(0)?.values}
      isLoadingConfig={isLoadingConfig}
      latestBuilds={latestBuilds?.data}
      latestDeploy={installComponent?.install_deploys?.[0]}
      overrideCard={overrideCard}
    />
  )
}
