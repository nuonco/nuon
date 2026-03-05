import { useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { checkVCSConnectionStatus, getVCSConnectionRepos } from '@/lib/ctl-api/vcs-connections'
import type { TVCSConnection } from '@/types'
import { GitHubAccountSection } from './GitHubAccountSection'
import { RepositoriesSection } from './RepositoriesSection'

interface IConnectionDetails {
  vcs_connection: TVCSConnection
}

export const ConnectionDetailsModal = ({
  vcs_connection,
  ...props
}: IConnectionDetails & IModal) => {
  const { org } = useOrg()

  const { data: status, isLoading: isLoadingStatus } = useQuery({
    queryKey: ['vcs-connection-status', org?.id, vcs_connection?.id],
    queryFn: () =>
      checkVCSConnectionStatus({ orgId: org!.id, connectionId: vcs_connection.id }),
    enabled: !!org?.id && !!vcs_connection?.id,
  })

  const {
    data: repos,
    error: reposError,
    isLoading: isLoadingRepos,
  } = useQuery({
    queryKey: ['vcs-connection-repos', org?.id, vcs_connection?.id],
    queryFn: () =>
      getVCSConnectionRepos({ orgId: org!.id, connectionId: vcs_connection.id }),
    enabled: !!org?.id && !!vcs_connection?.id,
  })

  return (
    <Modal
      heading={
        <div className="flex items-center gap-4">
          <div className="flex flex-col gap-2">
            <Text
              variant="h3"
              weight="strong"
              className="!flex gap-2 items-center"
            >
              <Icon variant="GitHub" size="24" />
              {vcs_connection?.github_account_name ||
                vcs_connection?.github_account_id ||
                'GitHub'}{' '}
              connection
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
                  <Time
                    time={status?.checked_at}
                    format="relative"
                    variant="subtext"
                  />
                </Text>
              </div>
            ) : null}
          </div>
        </div>
      }
      size="half"
      className="!max-h-[80vh]"
      childrenClassName="flex-auto overflow-y-auto"
      {...props}
    >
      <div className="flex flex-col gap-6">
        <GitHubAccountSection vcs_connection={vcs_connection} />
        <RepositoriesSection
          repos={repos}
          error={reposError}
          isLoading={isLoadingRepos}
        />
      </div>
    </Modal>
  )
}

export const ConnectionDetailsButton = ({
  vcs_connection,
  ...props
}: IConnectionDetails & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ConnectionDetailsModal vcs_connection={vcs_connection} />

  return (
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      View connection
      <Icon variant="Info" />
    </Button>
  )
}
