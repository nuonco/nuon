import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  AppCreateInstallButton,
  AppPageSubNav,
  AppWorkflowsTable,
  DashboardContent,
  ErrorFallback,
  Loading,
  NoActions,
  Section,
} from '@/components'
import { getApp, getAppActionWorkflows, getAppLatestInputConfig } from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const app = await getApp({ appId, orgId })

  return {
    title: `${app.name} | Actions`,
  }
}

export default async function AppWorkflows({ params }) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const [app, inputCfg] = await Promise.all([
    getApp({ appId, orgId }),
    getAppLatestInputConfig({ appId, orgId }).catch(console.error),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}`, text: app.name },
        { href: `/${orgId}/apps/${app.id}/actions`, text: 'Actions' },
      ]}
      heading={app.name}
      headingUnderline={app.id}
      statues={
        inputCfg ? (
          <AppCreateInstallButton
            platform={app?.cloud_platform}
            inputConfig={inputCfg}
            appId={appId}
            orgId={orgId}
          />
        ) : null
      }
      meta={<AppPageSubNav appId={appId} orgId={orgId} />}
    >
      <Section childrenClassName="flex flex-auto">
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading variant="page" loadingText="Loading actions..." />
            }
          >
            <LoadAppActions appId={appId} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
}

const LoadAppActions: FC<{ appId: string; orgId: string }> = async ({
  appId,
  orgId,
}) => {
  const actions = await getAppActionWorkflows({ appId, orgId })
  return actions && actions?.length ? (
    <AppWorkflowsTable appId={appId} orgId={orgId} workflows={actions} />
  ) : (
    <NoActions />
  )
}
