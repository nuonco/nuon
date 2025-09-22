import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  AppCreateInstallButton,
  AppPageSubNav,
  DashboardContent,
  ErrorFallback,
  Loading,
  Section,
} from '@/components'
import { getAppById, getAppConfigs } from '@/lib'
import { AppComponents } from './components'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getAppById({ appId, orgId })

  return {
    title: `Components | ${app.name} | Nuon`,
  }
}

export default async function AppComponentsPage({ params, searchParams }) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const sp = await searchParams
  const [{ data: app }, { data: configs }] = await Promise.all([
    getAppById({ appId, orgId }),
    getAppConfigs({ appId, orgId }),
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
        app?.cloud_platform === 'aws' || app.cloud_platform === 'azure' ? (
          <AppCreateInstallButton platform={app?.cloud_platform} />
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
            <AppComponents
              appId={appId}
              configId={configs?.at(0)?.id}
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
