import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { checkVCSConnectionStatus } from '@/lib/ctl-api/vcs-connections'
import type { TVCSConnection } from '@/types'
import { getStatusTheme } from '@/utils/vcs-connection-utils'
import { VCSConnectionItem } from './VCSConnections'

const VCSConnectionWithStatus = ({
  vcs_connection,
}: {
  vcs_connection: TVCSConnection
}) => {
  const { org } = useOrg()
  const { data, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['vcs-connection-status', org?.id, vcs_connection?.id],
    queryFn: () => checkVCSConnectionStatus({ orgId: org!.id, connectionId: vcs_connection.id }),
    enabled: !!org?.id && !!vcs_connection?.id,
  })

  const connectionHref = `/${org?.id}/settings/vcs/${vcs_connection.id}`

  return (
    <Text
      key={vcs_connection?.id}
      className="!flex gap-2 justify-between w-full"
      variant="subtext"
    >
      <VCSConnectionItem
        vcs_connection={vcs_connection}
        statusTheme={getStatusTheme(data?.status)}
        isLoadingStatus={isLoading}
        href={connectionHref}
      />
    </Text>
  )
}

export const VCSConnections = ({
  vcsConnections,
}: {
  vcsConnections: TVCSConnection[]
}) => {
  return (
    <>
      {vcsConnections?.length &&
        vcsConnections?.map((vcs) => (
          <VCSConnectionWithStatus key={vcs?.id} vcs_connection={vcs} />
        ))}
    </>
  )
}
