import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  AppCreateInstallButton,
  AppInstallsTable,
  AppPageSubNav,
  DashboardContent,
  ErrorFallback,
  Loading,
  Notice,
  NoInstalls,
  Pagination,
  Section,
} from '@/components'
import { getApp, getAppLatestInputConfig } from '@/lib'
import type { TApp, TInstall } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const app = await getApp({ appId, orgId })

  return {
    title: `${app.name} | Installs`,
  }
}

export default async function AppInstalls({ params, searchParams }) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const sp = await searchParams
  const [app, inputCfg] = await Promise.all([
    getApp({ appId, orgId }).catch((error) => {
      console.error(error)
      notFound()
    }),
    getAppLatestInputConfig({ appId, orgId }).catch((error) => {
      console.error(error)
    }),
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
            <LoadAppInstalls
              app={app}
              appId={appId}
              orgId={orgId}
              offset={sp['offset'] || '0'}
              q={sp['q'] || ''}
            />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
}

const LoadAppInstalls: FC<{
  app: TApp
  appId: string
  orgId: string
  limit?: string
  offset?: string
  q?: string
}> = async ({ app, appId, orgId, limit = '10', offset, q }) => {
  const params = new URLSearchParams({ offset, limit, q }).toString()
  const {
    data: installs,
    error,
    headers,
  } = await nueQueryData<TInstall[]>({
    orgId,
    path: `apps/${appId}/installs${params ? '?' + params : params}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return error ? (
    <Notice>Can&apos;t load installs: {error?.error}</Notice>
  ) : installs ? (
    <div className="flex flex-col gap-8 w-full">
      <AppInstallsTable
        installs={installs.map((install) => ({ ...install, app }))}
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
    <NoInstalls />
  )
}
