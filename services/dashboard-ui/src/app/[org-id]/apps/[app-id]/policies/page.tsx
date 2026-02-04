import type { Metadata } from 'next'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { PoliciesTable, PoliciesTableSkeleton } from './policies-table'

type TAppPageProps = TPageProps<'org-id' | 'app-id'>

export async function generateMetadata({
  params,
}: TAppPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getApp({ appId, orgId })

  return {
    title: `Policies | ${app?.name} | Nuon`,
  }
}

export default async function AppPoliciesPage({ params }: TAppPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const [{ data: app }, { data: org }] = await Promise.all([
    getApp({ appId, orgId }),
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
            path: `/${orgId}/apps`,
            text: 'Apps',
          },
          {
            path: `/${orgId}/apps/${appId}`,
            text: app?.name,
          },
          {
            path: `/${orgId}/apps/${appId}/policies`,
            text: 'Policies',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          App policies
        </Text>
      </HeadingGroup>

      <div className="flex flex-auto">
        <AsyncBoundary
          errorFallback={
            <span className="text-md">Unable to load app policies</span>
          }
          loadingFallback={<PoliciesTableSkeleton />}
        >
          <PoliciesTable appId={appId} orgId={orgId} />
        </AsyncBoundary>
      </div>
    </PageSection>
  )
}
