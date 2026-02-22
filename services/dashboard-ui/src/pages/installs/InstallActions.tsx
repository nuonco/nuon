import { useParams, useSearchParams } from 'react-router-dom'
import { usePolling } from '@/hooks/use-polling'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { InstallActionsTable } from '@/components/actions/InstallActionsTable'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import type { TInstallActionWithLatestRun } from '@/types'

const LIMIT = 10

const InstallActionsTableWrapper = ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) => {
  const [searchParams] = useSearchParams()

  const offset = searchParams.get('offset') || '0'
  const q = searchParams.get('q') || ''
  const trigger_types = searchParams.get('trigger_types') || ''

  const {
    data: actionsWithRuns,
    error,
    isLoading,
    headers,
  } = usePolling<TInstallActionWithLatestRun[]>({
    path: `/api/ctl-api/v1/installs/${installId}/actions/latest-runs?limit=${LIMIT}&offset=${offset}${q ? `&q=${q}` : ''}${trigger_types ? `&trigger_types=${trigger_types}` : ''}`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  const pagination = {
    limit: Number(headers?.['x-nuon-page-limit'] ?? LIMIT),
    hasNext: headers?.['x-nuon-page-next'] === 'true',
    offset: Number(headers?.['x-nuon-page-offset'] ?? '0'),
  }

  if (error && !actionsWithRuns) {
    return (
      <div>
        <p>Could not load your actions.</p>
        <p>{error.message || 'Unknown error'}</p>
      </div>
    )
  }

  if (isLoading && !actionsWithRuns) {
    return <div>Loading actions...</div>
  }

  return (
    <InstallActionsTable
      actionsWithRuns={actionsWithRuns || []}
      pagination={pagination}
      shouldPoll
    />
  )
}

export default function InstallActions() {
  const { installId, orgId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  if (!installId || !orgId) {
    return null
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

      <InstallActionsTableWrapper installId={installId} orgId={orgId} />

      <BackToTop />
    </PageSection>
  )
}