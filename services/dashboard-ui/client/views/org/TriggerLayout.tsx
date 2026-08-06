import { Outlet, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { TabNav } from '@/components/navigation/TabNav'
import { useOrg } from '@/hooks/use-org'
import { getTrigger } from '@/lib'
export const TriggerLayout = () => {
  const { triggerId } = useParams()
  const { org } = useOrg()
  const query = useQuery({
    queryKey: ['trigger', org?.id, triggerId],
    queryFn: () => getTrigger({ orgId: org!.id, triggerId: triggerId! }),
    enabled: !!org?.id && !!triggerId,
  })
  const trigger = query.data
  const base = `/${org?.id}/settings/triggers/${triggerId}`
  return (
    <>
      <PageTitle title={`${trigger?.name || 'Trigger'} | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/settings`, text: 'Settings' },
          { path: `/${org?.id}/settings/triggers`, text: 'Triggers' },
          { path: base, text: trigger?.name || triggerId },
        ]}
      />
      <PageHeader>
        <div className="flex flex-col gap-2">
          <div className="flex items-center gap-3">
            <Text variant="h3" weight="stronger" level={1}>
              {trigger?.name || 'Trigger'}
            </Text>
            <Status
              status={trigger?.status === 'active' ? 'success' : 'warn'}
              variant="badge"
            >
              {trigger?.status || 'Unknown'}
            </Status>
          </div>
          {triggerId ? <ID>{triggerId}</ID> : null}
          {trigger?.description ? (
            <Text theme="neutral">{trigger.description}</Text>
          ) : null}
        </div>
      </PageHeader>
      <PageContent>
        <PageSection>
          <TabNav
            basePath={base}
            tabs={[
              { path: '/', text: 'Overview' },
              { path: '/rules', text: 'Rules' },
              { path: '/events', text: 'Events' },
            ]}
          />
          {query.isLoading ? (
            <Text theme="neutral">Loading trigger...</Text>
          ) : query.error || !trigger ? (
            <div className="flex flex-col items-start gap-3">
              <Text theme="error">Trigger loading failed.</Text>
              <Button variant="secondary" onClick={() => void query.refetch()}>
                <Icon variant="ArrowClockwiseIcon" />
                Retry loading trigger
              </Button>
            </div>
          ) : (
            <Outlet context={{ trigger }} />
          )}
        </PageSection>
      </PageContent>
    </>
  )
}
