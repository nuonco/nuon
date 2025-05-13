import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppCreateInstallButton,
  AppComponentsTable,
  AppPageSubNav,
  DashboardContent,
  ErrorFallback,
  Loading,
  NoComponents,
  Section,
} from '@/components'
import {
  getApp,
  getAppComponents,
  getComponentBuilds,
  getComponentConfig,
  getAppLatestInputConfig,
} from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const app = await getApp({ appId, orgId })

  return {
    title: `${app.name} | Components`,
  }
}

export default withPageAuthRequired(async function AppComponents({ params }) {
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
        { href: `/${orgId}/apps/${app.id}/components`, text: 'Components' },
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
              <Loading variant="page" loadingText="Loading components..." />
            }
          >
            <LoadAppComponents appId={appId} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
})

const LoadAppComponents: FC<{ appId: string; orgId: string }> = async ({
  appId,
  orgId,
}) => {
  const components = await getAppComponents({ appId, orgId })

  const hydratedComponents = await Promise.all(
    components.map(async (comp, _, arr) => {
      const [config, builds] = await Promise.all([
        getComponentConfig({ componentId: comp.id, orgId }).catch(
          console.error
        ),
        getComponentBuilds({ componentId: comp.id, orgId }).catch(
          console.error
        ),
      ])
      const deps = arr.filter((c) => comp.dependencies?.some((d) => d === c.id))

      return {
        ...comp,
        config: config || undefined,
        deps,
        latestBuild: builds[0],
      }
    })
  )

  return components.length ? (
    <AppComponentsTable
      components={hydratedComponents}
      appId={appId}
      orgId={orgId}
    />
  ) : (
    <NoComponents />
  )
}
