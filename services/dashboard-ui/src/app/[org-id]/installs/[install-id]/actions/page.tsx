import type { Metadata } from 'next'
import { Suspense } from 'react'
import { InstallActionsTableSkeleton } from '@/components/actions/InstallActionsTable'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getInstall, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { InstallActionsTable } from './actions-table'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstall({ installId, orgId })

  return {
    title: `Actions | ${install.name} | Nuon`,
  }
}

export default async function InstallActionsPage({
  params,
  searchParams,
}: TInstallPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const [{ data: install }, { data: org }] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
  ])

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
            path: `/${orgId}/installs/${installId}/actions`,
            text: 'Actions',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Actions
        </Text>
        <Text theme="neutral">
          View and manage all actions for this install.
        </Text>
      </HeadingGroup>
      <ErrorBoundary
        fallback={
          <Text>
            An error loading your install components, please refresh the page
            and try again.
          </Text>
        }
      >
        <Suspense fallback={<InstallActionsTableSkeleton />}>
          <InstallActionsTable
            installId={installId}
            orgId={orgId}
            offset={sp['offset'] || '0'}
            q={sp['q'] || ''}
            trigger_types={sp['trigger_types'] || ''}
          />
        </Suspense>
      </ErrorBoundary>
    </PageSection>
  )
}
