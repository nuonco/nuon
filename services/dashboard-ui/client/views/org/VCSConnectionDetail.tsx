import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useNavigate, useParams } from 'react-router'
import { Icon } from '@/components/common/Icon'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Time } from '@/components/common/Time'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
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
    placeholderData: keepPreviousData,
    queryKey: ['vcs-connection', org?.id, connectionId],
    queryFn: () =>
      getVCSConnection({ orgId: org!.id, connectionId: connectionId! }),
    enabled: !!org?.id && !!connectionId,
  })

  const { data: status, isLoading: isLoadingStatus } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['vcs-connection-status', org?.id, connectionId],
    queryFn: () =>
      checkVCSConnectionStatus({ orgId: org!.id, connectionId: connectionId! }),
    enabled: !!org?.id && !!connectionId,
    refetchInterval: 60_000,
  })

  const accountName =
    vcs_connection?.github_account_name ||
    vcs_connection?.github_account_id ||
    'GitHub'

  return (
    <>
      <PageTitle title={`${accountName} connection`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/settings`, text: 'Settings' },
          { path: `/${org?.id}/settings/vcs`, text: 'VCS connections' },
          {
            path: `/${org?.id}/settings/vcs/${connectionId}`,
            text: `${accountName} connection`,
          },
        ]}
      />
      <DetailPage
        header={
          <DetailHeader
            backLink={false}
            icon={<Icon variant="GitHub" size="24" />}
            title={`${accountName} connection`}
            id={connectionId}
            actions={
              <>
                <ConnectGithubButton size="md">
                  Add connection
                </ConnectGithubButton>
                {vcs_connection && (
                  <RemoveConnectionButton
                    vcs_connection={vcs_connection}
                    onRemoveSuccess={() => navigate(`/${org?.id}/settings/vcs`)}
                  />
                )}
              </>
            }
            metadata={
              <>
                <LabeledStatus
                  label="Status"
                  loading={isLoadingStatus}
                  statusProps={{ status: status?.status }}
                />
                <LabeledValue label="Last checked" loading={isLoadingStatus}>
                  <Time
                    time={status?.checked_at}
                    format="relative"
                    variant="subtext"
                    shouldTick
                  />
                </LabeledValue>
              </>
            }
          />
        }
      >
        {vcs_connection && <ConnectionDetail vcs_connection={vcs_connection} />}
      </DetailPage>
    </>
  )
}
