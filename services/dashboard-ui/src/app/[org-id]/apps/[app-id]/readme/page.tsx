import type { Metadata } from 'next'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { BackToTop } from '@/components/common/BackToTop'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getAppConfigs, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { Readme, ReadmeError, ReadmeSkeleton } from './readme'

type TAppPageProps = TPageProps<'org-id' | 'app-id'>

export async function generateMetadata({
  params,
}: TAppPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getApp({ appId, orgId })

  return {
    title: `README | ${app.name} | Nuon`,
  }
}

export default async function AppReadmePage({ params }: TAppPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const [{ data: app }, { data: configs }, { data: org }] = await Promise.all([
    getApp({ appId, orgId }),
    getAppConfigs({ appId, orgId, limit: 1 }),
    getOrg({ orgId }),
  ])

  const containerId = 'app-readme-page'
  return (
    <PageSection id={containerId} className="!pb-6" isScrollable>
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
        ]}
      />
      <AsyncBoundary
        errorFallback={<ReadmeError />}
        loadingFallback={<ReadmeSkeleton />}
      >
        <Readme appConfigId={configs?.at(0)?.id} appId={appId} orgId={orgId} />
      </AsyncBoundary>
      <BackToTop containerId={containerId} />
    </PageSection>
  )
}
