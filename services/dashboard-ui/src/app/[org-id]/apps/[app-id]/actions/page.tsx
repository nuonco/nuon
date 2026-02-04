import type { Metadata } from 'next'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { ActionsTable, ActionsTableSkeleton } from './actions-table'

type TAppPageProps = TPageProps<'org-id' | 'app-id'>

export async function generateMetadata({
  params,
}: TAppPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getApp({ appId, orgId })

  return {
    title: `Actions | ${app.name} | Nuon`,
  }
}

export default async function AppActionsPage({
  params,
  searchParams,
}: TAppPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const sp = await searchParams
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
            path: `/${orgId}/apps/${appId}/actions`,
            text: 'Actions',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          App actions
        </Text>
      </HeadingGroup>

      <AsyncBoundary
        errorFallback={
          <span className="text-md">Unable to load app actions</span>
        }
        loadingFallback={<ActionsTableSkeleton />}
      >
        <ActionsTable
          appId={appId}
          orgId={orgId}
          offset={sp['offset'] || '0'}
          q={sp['q'] || ''}
        />
      </AsyncBoundary>
    </PageSection>
  )
}
