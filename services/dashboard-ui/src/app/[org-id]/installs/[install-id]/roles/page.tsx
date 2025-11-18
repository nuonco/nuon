import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getInstallById, getOrgById } from '@/lib'
import { InstallRoles, InstallRolesSkeleton, InstallRolesError } from './roles'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

  return {
    title: `Roles | ${install.name} | Nuon`,
  }
}

export default async function InstallRolesPage({ params }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const [{ data: install, error }, { data: org }] = await Promise.all([
    getInstallById({ installId, orgId }),
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
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name,
          },
          {
            path: `/${orgId}/installs/${installId}/roles`,
            text: 'Roles',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          IAM roles
        </Text>
        <Text variant="subtext" theme="neutral">
          View the IAM roles that your install uses to access customer AWS
          resources.
        </Text>
      </HeadingGroup>

      <ErrorBoundary fallback={<InstallRolesError />}>
        <Suspense fallback={<InstallRolesSkeleton />}>
          <InstallRoles
            appConfigId={install?.app_config_id}
            appId={install?.app_id}
            orgId={orgId}
          />
        </Suspense>
      </ErrorBoundary>
    </PageSection>
  )
}
