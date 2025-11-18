import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getAppById, getOrgById } from '@/lib'
import { AppRoles, AppRolesSkeleton, AppRolesError } from './roles'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getAppById({ appId, orgId })

  return {
    title: `Roles | ${app.name} | Nuon`,
  }
}

export default async function AppRolesPage({ params }) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const [{ data: app, error }, { data: org }] = await Promise.all([
    getAppById({ appId, orgId }),
    getOrgById({ orgId }),
  ])

  if (error) {
    notFound()
  }

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/apps`,
            text: 'Apps',
          },
          {
            path: `/${orgId}/apps/${appId}`,
            text: app?.name,
          },
          {
            path: `/${orgId}/apps/${appId}/roles`,
            text: 'Roles',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          IAM roles
        </Text>
        <Text variant="subtext" theme="neutral">
          View the IAM roles that your app uses to access customer AWS resources.
        </Text>
      </HeadingGroup>
      
      <ErrorBoundary fallback={<AppRolesError />}>
        <Suspense fallback={<AppRolesSkeleton />}>
          <AppRoles appId={appId} orgId={orgId} />
        </Suspense>
      </ErrorBoundary>
    </PageSection>
  )
}
