import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppCreateInstallButton,
  AppInstallsTable,
  AppPageSubNav,
  DashboardContent,
  ErrorFallback,
  Loading,
  NoInstalls,
  Section,
} from '@/components'
import { getApp, getAppInstalls, getAppLatestInputConfig } from '@/lib'
import type { TApp } from '@/types'

export async function generateMetadata({ params }): Promise<Metadata> {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const app = await getApp({ appId, orgId })

  return {
    title: `${app.name} | Installs`,
  }
}

export default withPageAuthRequired(async function AppInstalls({ params }) {
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
        { href: `/${orgId}/apps/${app.id}/installs`, text: 'Installs' },
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
              <Loading variant="page" loadingText="Loading installs..." />
            }
          >
            <LoadAppInstalls app={app} appId={appId} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
})

const LoadAppInstalls: FC<{
  app: TApp
  appId: string
  orgId: string
}> = async ({ app, appId, orgId }) => {
  const installs = await getAppInstalls({ appId, orgId })

  return installs.length ? (
    <AppInstallsTable
      installs={installs.map((install) => ({ ...install, app }))}
      orgId={orgId}
    />
  ) : (
    <NoInstalls />
  )
}
