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
  getAppLatestInputConfig,
  getAppLatestConfig,
} from '@/lib'
import type { TBuild } from '@/types'
import { nueQueryData } from '@/utils'

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
  const [app, appConfig, inputCfg] = await Promise.all([
    getApp({ appId, orgId }),
    getAppLatestConfig({ appId, orgId }),
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
            <LoadAppComponents
              appId={appId}
              configId={appConfig?.id}
              orgId={orgId}
            />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
})

const LoadAppComponents: FC<{
  appId: string
  configId: string
  orgId: string
}> = async ({ appId, configId, orgId }) => {
  const components = await getAppComponents({ appId, orgId })
  const hydratedComponents = await Promise.all(
    components
      //.filter((c) => c?.type === 'helm_chart' || c?.type === 'terraform_module')
      .sort((a, b) => a?.id?.localeCompare(b?.id))
      .map(async (comp, _) => {
        const { data: build } = await nueQueryData<TBuild>({
          orgId,
          path: `components/${comp?.id}/builds/latest`,
        })

        const deps = components.filter((c) =>
          comp.dependencies?.some((d) => d === c.id)
        )

        return {
          ...comp,
          latestBuild: build,
          deps,
        }
      })
  )

  return components.length ? (
    <AppComponentsTable
      components={hydratedComponents}
      appId={appId}
      configId={configId}
      orgId={orgId}
    />
  ) : (
    <NoComponents />
  )
}
