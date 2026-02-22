import { useParams, useSearchParams } from 'react-router-dom'
import { useOrg } from '@/hooks/use-org'
import { usePolling } from '@/hooks/use-polling'
import { AppsTable } from '@/components/apps/AppsTable'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import type { TApp } from '@/types'

const LIMIT = 10

export default function AppsPage() {
  const { orgId } = useParams()
  const { org } = useOrg()
  const [searchParams] = useSearchParams()

  const offset = searchParams.get('offset') || '0'
  const q = searchParams.get('q') || ''

  const { data: response, error, isLoading, headers } = usePolling<TApp[]>({
    path: `/api/ctl-api/v1/apps?limit=${LIMIT}&offset=${offset}${q ? `&q=${q}` : ''}`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? LIMIT),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  if (error && !response && !isLoading) {
    return (
      <PageLayout>
        <PageContent>
          <div>
            <p>Could not load your apps.</p>
            <p>{error.error}</p>
            <Link href="/api/auth/logout">Log out</Link>
          </div>
        </PageContent>
      </PageLayout>
    )
  }

  return (
    <PageLayout isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name || '' },
          { path: `/${org?.id}/apps`, text: 'Apps' },
        ]}
      />
      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Apps
          </Text>
          <Text theme="neutral">Manage your applications here.</Text>
        </HeadingGroup>
      </PageHeader>
      <PageContent>
        <PageSection>
          <AppsTable
            apps={response || []}
            pagination={pagination}
            shouldPoll
          />
        </PageSection>
      </PageContent>
    </PageLayout>
  )
}