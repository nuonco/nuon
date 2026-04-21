import { useQuery } from '@tanstack/react-query'
import { useNavigate, useParams } from 'react-router'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { ConnectGithubButton } from '@/components/vcs-connections/ConnectGithub'
import { ConnectionDetail } from '@/components/vcs-connections/ConnectionDetail'
import { RemoveConnectionButton } from '@/components/vcs-connections/RemoveConnection'
import { useOrg } from '@/hooks/use-org'
import { checkVCSConnectionStatus, getVCSConnection } from '@/lib'

export const VCSConnectionDetail = () => {
  const { connectionId } = useParams<{ connectionId: string }>()
  const { org } = useOrg()
  const navigate = useNavigate()

  const { data: vcs_connection } = useQuery({
    queryKey: ['vcs-connection', org?.id, connectionId],
    queryFn: () => getVCSConnection({ orgId: org!.id, connectionId: connectionId! }),
    enabled: !!org?.id && !!connectionId,
  })

  const { data: status, isLoading: isLoadingStatus } = useQuery({
    queryKey: ['vcs-connection-status', org?.id, connectionId],
    queryFn: () => checkVCSConnectionStatus({ orgId: org!.id, connectionId: connectionId! }),
    enabled: !!org?.id && !!connectionId,
    refetchInterval: 60_000,
  })

  const accountName =
    vcs_connection?.github_account_name ||
    vcs_connection?.github_account_id ||
    'GitHub'

  return (
    <PageLayout className="pb-6">
      <PageTitle title={`${accountName} connection | ${org?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/connections/vcs/${connectionId}`, text: `${accountName} connection` },
        ]}
      />
      <PageHeader className="flex items-start justify-between">
        <div className="flex flex-col gap-2">
          <Text variant="h3" weight="stronger" level={1} className="!flex gap-2 items-center">
            <Icon variant="GitHub" size="24" />
            {accountName} connection
          </Text>
          {isLoadingStatus ? (
            <div className="flex items-center gap-2">
              <Skeleton height="24px" width="65px" />
              <Skeleton height="17px" width="160px" />
            </div>
          ) : status ? (
            <div className="flex items-center gap-2">
              <Status status={status.status} variant="badge" />
              <Text variant="subtext" theme="neutral">
                Last checked{' '}
                <Time time={status?.checked_at} format="relative" variant="subtext" shouldTick />
              </Text>
            </div>
          ) : null}
        </div>
        <div className="flex items-center gap-2">
          <ConnectGithubButton size="md">Add connection</ConnectGithubButton>
          {vcs_connection && (
            <RemoveConnectionButton
              vcs_connection={vcs_connection}
              onRemoveSuccess={() => navigate(`/${org?.id}`)}
            />
          )}
        </div>
      </PageHeader>
      <PageContent>
        <PageSection>
          {vcs_connection && <ConnectionDetail vcs_connection={vcs_connection} />}
        </PageSection>
      </PageContent>
    </PageLayout>
  )
}
