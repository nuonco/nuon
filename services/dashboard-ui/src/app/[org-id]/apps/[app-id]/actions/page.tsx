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
  Pagination,
  Section,
} from '@/components'
import { getApp, getAppLatestInputConfig } from '@/lib'
import type { TActionWorkflow } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const app = await getApp({ appId, orgId })

  return {
    title: `${app.name} | Actions`,
  }
}

export default async function AppWorkflows({ params, searchParams }) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const sp = await searchParams
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
            <LoadAppActions
              appId={appId}
              orgId={orgId}
              offset={sp['offset'] || '0'}
            />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
}

const LoadAppActions: FC<{
  appId: string
  orgId: string
  limit?: string
  offset?: string
}> = async ({ appId, orgId, limit = '10', offset }) => {
  const params = new URLSearchParams({ offset, limit }).toString()
  const {
    data: actions,
    error,
    headers,
  } = await nueQueryData<TActionWorkflow[]>({
    orgId,
    path: `apps/${appId}/action-workflows${params ? '?' + params : params}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }
  return actions && actions?.length && !error ? (
    <div className="flex flex-col gap-4 w-full">
      <AppWorkflowsTable appId={appId} orgId={orgId} workflows={actions} />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={parseInt(limit)}
      />
    </div>
  ) : (
    <NoActions />
  )
}
