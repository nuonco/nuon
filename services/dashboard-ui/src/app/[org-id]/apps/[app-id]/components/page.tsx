import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  AppCreateInstallButton,
  AppComponentsTable,
  AppPageSubNav,
  DashboardContent,
  ErrorFallback,
  Loading,
  NoComponents,
  Notice,
  Pagination,
  Section,
} from '@/components'
import { getApp, getAppLatestInputConfig, getAppLatestConfig } from '@/lib'
import type { TAppConfig, TBuild, TComponent } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const app = await getApp({ appId, orgId })

  return {
    title: `${app.name} | Components`,
  }
}

export default async function AppComponents({ params, searchParams }) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const sp = await searchParams
  const [app, appConfig, inputCfg] = await Promise.all([
    getApp({ appId, orgId }),
    getAppLatestConfig({ appId, orgId }).catch(console.error),
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
              configId={(appConfig as TAppConfig)?.id}
              orgId={orgId}
              offset={sp['offset'] || '0'}
              q={sp['q'] || ''}
              types={sp['types'] || ''}
            />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
}

const LoadAppComponents: FC<{
  appId: string
  configId: string
  orgId: string
  limit?: string
  offset?: string
  q?: string
  types?: string
}> = async ({ appId, configId, orgId, limit = '10', offset, q, types }) => {
  const params = new URLSearchParams({ offset, limit, q, types }).toString()
  const {
    data: components,
    error,
    headers,
  } = await nueQueryData<TComponent[]>({
    orgId,
    path: `apps/${appId}/components${params ? '?' + params : params}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })
  const hydratedComponents =
    components &&
    !error &&
    (await Promise.all(
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
    ))
  

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return error ? (
    <Notice>Can&apos;t load components: {error?.error}</Notice>
  ) : components ? (
    <div className="flex flex-col gap-4 w-full">
      <AppComponentsTable
        initComponents={hydratedComponents}
        appId={appId}
        configId={configId}
        orgId={orgId}
      />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={parseInt(limit)}
      />
    </div>
  ) : (
    <NoComponents />
  )
}
