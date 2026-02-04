import type { Metadata } from 'next'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getInstall, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import {
  InstallComponentsTable,
  InstallComponentsTableSkeleton,
} from './components-table'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstall({ installId, orgId })

  return {
    title: `Components | ${install.name} | Nuon`,
  }
}

export default async function InstallComponentsPage({
  params,
  searchParams,
}: TInstallPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const [{ data: install }, { data: org }] = await Promise.all([
    getInstall({ orgId, installId }),
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
            path: `/${orgId}/installs/${installId}/components`,
            text: 'Components',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install components
        </Text>
        <Text theme="neutral">
          View and manage all components for this install.
        </Text>
      </HeadingGroup>

      <AsyncBoundary
        loadingFallback={<InstallComponentsTableSkeleton />}
        errorFallback={
          <span className="text-md">Unable to load install components</span>
        }
      >
        <InstallComponentsTable
          install={install}
          installId={install?.id}
          orgId={orgId}
          offset={sp['offset'] || '0'}
          q={sp['q'] || ''}
          types={sp['types'] || ''}
        />
      </AsyncBoundary>
    </PageSection>
  )
}
