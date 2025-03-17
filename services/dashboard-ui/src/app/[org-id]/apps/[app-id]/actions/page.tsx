import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
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
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const app = await getApp({ appId, orgId })

  return {
    title: `${app.name} | Actions`,
  }
}

export default withPageAuthRequired(async function AppWorkflows({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [app, inputCfg] = await Promise.all([
    getApp({ appId, orgId }),
    getAppLatestInputConfig({ appId, orgId }).catch(console.error),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}`, text: app.name },
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
      <Section>
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
})

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
