import { Outlet, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useOrg } from '@/hooks/use-org'
import { getTrigger } from '@/lib'
export const TriggerLayout = () => {
  const { triggerId } = useParams()
  const { org } = useOrg()
  const query = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['trigger', org?.id, triggerId],
    queryFn: () => getTrigger({ orgId: org!.id, triggerId: triggerId! }),
    enabled: !!org?.id && !!triggerId,
  })
  const trigger = query.data
  const base = `/${org?.id}/settings/triggers/${triggerId}`
  return (
    <>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/settings`, text: 'Settings' },
          { path: `/${org?.id}/settings/triggers`, text: 'Triggers' },
          { path: base, text: trigger?.name || triggerId },
        ]}
      />
      <DetailPage
        header={
          <DetailHeader
            backLink={false}
            title={trigger?.name || 'Trigger'}
            description={trigger?.description}
            status={
              <Status
                status={trigger?.status === 'active' ? 'success' : 'warn'}
                variant="badge"
              >
                {trigger?.status || 'Unknown'}
              </Status>
            }
            id={triggerId}
          />
        }
        tabNav={{
          basePath: base,
          tabs: [
            { path: '/', text: 'Overview' },
            { path: '/rules', text: 'Rules' },
            { path: '/events', text: 'Events' },
          ],
        }}
      >
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
      </DetailPage>
    </>
  )
}
