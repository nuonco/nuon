import { useQuery } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { checkVCSConnectionStatus, getVCSConnectionRepos } from '@/lib/ctl-api/vcs-connections'
import type { TVCSConnection } from '@/types'
import { ConnectionDetailsModal } from './ConnectionDetails'

interface IConnectionDetails {
  vcs_connection: TVCSConnection
}

export const ConnectionDetailsModalContainer = ({
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
    <ConnectionDetailsModal
      vcs_connection={vcs_connection}
      status={status}
      isLoadingStatus={isLoadingStatus}
      repos={repos}
      reposError={reposError}
      isLoadingRepos={isLoadingRepos}
      {...props}
    />
  )
}

export const ConnectionDetailsButton = ({
  vcs_connection,
  ...props
}: IConnectionDetails & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ConnectionDetailsModalContainer vcs_connection={vcs_connection} />

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
