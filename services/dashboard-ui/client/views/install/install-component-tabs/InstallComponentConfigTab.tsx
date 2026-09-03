import { useOutletContext, useParams } from 'react-router'
import { EmptyState } from '@/components/common/EmptyState/EmptyState'
import { Text } from '@/components/common/Text'
import { ComponentConfigCard } from '@/components/components/ComponentConfigCard'
import { ComponentDependencyGraphButton } from '@/components/components/ComponentDependencyGraph'
import { InstallComponentDependencies } from '@/components/install-components/InstallComponentDependencies'
import { ComponentOverrideCard } from '@/components/install-overrides/ComponentOverrideCard'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type { TInstallComponentOutletContext } from './types'

export const InstallComponentConfigTab = () => {
  const { componentId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const {
    appConfig,
    config,
    dependentIds,
    installComponent,
    installValues,
    isLoadingConfig,
    latestBuilds,
    latestDeploy,
    overrideCard,
  } = useOutletContext<TInstallComponentOutletContext>()

  const component = installComponent?.component
  const latestResolvedBuild = latestBuilds?.find((b) => !!b.source_digest)
  const deployCommit = latestDeploy?.component_build?.vcs_connection_commit
  const latestCommit = deployCommit
    ? {
        status: latestDeploy?.status_v2?.status,
        href: `/${org?.id}/installs/${install?.id}/components/${componentId}/deploys/${latestDeploy?.id}`,
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
      <PageTitle segments={['Component config', install?.name]} />

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
                basePath={`/${org?.id}/installs/${install?.id}/components`}
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
