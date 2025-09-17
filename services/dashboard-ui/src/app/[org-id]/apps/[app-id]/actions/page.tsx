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
  Notice,
  Pagination,
  Section,
} from '@/components'
import { getAppById, getAppLatestInputConfig, getActions } from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getAppById({ appId, orgId })

  return {
    title: `Actions | ${app.name} | Nuon`,
  }
}

export default async function AppWorkflows({ params, searchParams }) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const sp = await searchParams
  const [{ data: app }, inputCfg] = await Promise.all([
    getAppById({ appId, orgId }),
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
              q={sp['q'] || ''}
            />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
}

// TODO(nnnnat): move to server component file
const LoadAppActions: FC<{
  appId: string
  orgId: string
  limit?: number
  offset?: string
  q?: string
}> = async ({ appId, orgId, limit = 10, offset, q }) => {
  const {
    data: actions,
    error,
    headers,
  } = await getActions({
    appId,
    limit,
    offset,
    orgId,
    q,
  })

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }
  return error ? (
    <Notice>Can&apos;t load actions: {error?.error}</Notice>
  ) : actions ? (
    <div className="flex flex-col gap-4 w-full">
      <AppWorkflowsTable appId={appId} orgId={orgId} workflows={actions} />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={limit}
      />
    </div>
  ) : (
    <NoActions />
  )
}
