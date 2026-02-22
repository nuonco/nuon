import { useParams, useSearchParams } from 'react-router-dom'
import { useOrg } from '@/hooks/use-org'
import { usePolling } from '@/hooks/use-polling'
import { InstallsTable } from '@/components/installs/InstallsTable'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import type { TInstall } from '@/types'

const LIMIT = 10

export default function InstallsPage() {
  const { orgId } = useParams()
  const { org } = useOrg()
  const [searchParams] = useSearchParams()

  const offset = searchParams.get('offset') || '0'
  const q = searchParams.get('q') || ''

  const { data: response, error, isLoading, headers } = usePolling<TInstall[]>({
    path: `/api/ctl-api/v1/installs?limit=${LIMIT}&offset=${offset}${q ? `&q=${q}` : ''}`,
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
            <p>Could not load your installs.</p>
            <p>{error.error}</p>
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
          { path: `/${org?.id}/installs`, text: 'Installs' },
        ]}
      />
      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Installs
          </Text>
          <Text theme="neutral">View and manage all deployed installs here.</Text>
        </HeadingGroup>
      </PageHeader>
      <PageContent>
        <PageSection>
          <InstallsTable
            installs={response || []}
            pagination={pagination}
            shouldPoll
          />
        </PageSection>
      </PageContent>
    </PageLayout>
  )
}