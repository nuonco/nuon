import type { Metadata } from 'next'
import { Suspense } from 'react'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getInstall, getOrg } from '@/lib'
import { TPageProps } from '@/types'
import { InstallStacksTable, InstallStacksTableSkeleton } from './stacks-table'

// NOTE: old layout stuff
import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  InstallStatuses,
  InstallPageSubNav,
  InstallManagementDropdown,
  Link as OldLink,
  Loading,
  Text as OldText,
  Time,
} from '@/components'
import { Stacks } from './stacks'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install }: any = await getInstall({ installId, orgId })

  return {
    title: `Stacks | ${install.name} | Nuon`,
  }
}

export default async function InstallStack({ params }: TInstallPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const [{ data: install }, { data: org }] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({
      orgId,
    }),
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
            path: `/${orgId}/installs/${installId}/stacks`,
            text: 'Stacks',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install stacks
        </Text>
        <Text variant="subtext" theme="neutral">
          View your install stack versions below.
        </Text>
      </HeadingGroup>

      <OldErrorBoundary fallbackRender={ErrorFallback}>
        <Suspense fallback={<InstallStacksTableSkeleton />}>
          <InstallStacksTable installId={install?.id} orgId={orgId} />
        </Suspense>
      </OldErrorBoundary>
    </PageSection>
  )
}
