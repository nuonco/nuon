'use client'

import { Status } from '@/components/common/Status'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { useOrg } from '@/hooks/use-org'
import { useQuery } from '@/hooks/use-query'
import type { TVCSConnection, TVCSConnectionStatus } from '@/types'
import { cn } from '@/utils/classnames'
import { getStatusTheme } from '@/utils/vcs-connection-utils'
import { VCSManagementDropdown } from './management/VCSManagementDropdown'
import { VCSAccountLink } from './VCSAccountLink'

export const VCSConnections = ({
  vcsConnections,
}: {
  vcsConnections: TVCSConnection[]
}) => {
  return (
    <>
      {vcsConnections?.length &&
        vcsConnections?.map((vcs) => (
          <Text
            key={vcs?.id}
            className="!flex gap-2 justify-between w-full"
            variant="subtext"
          >
            <VCSConnection vcs_connection={vcs} />
            <span className="self-end">
              <Tooltip tipContent="More" position="left">
                <VCSManagementDropdown vcs_connection={vcs} />
              </Tooltip>
            </span>
          </Text>
        ))}
    </>
  )
}

const VCSConnection = ({
  vcs_connection,
}: {
  vcs_connection: TVCSConnection
}) => {
  const { org } = useOrg()
  const { data, isLoading } = useQuery<TVCSConnectionStatus>({
    path: `/api/orgs/${org?.id}/vcs-connections/${vcs_connection?.id}/check-status`,
  })

  return (
    <span className="!flex gap-2 items-center w-full">
      <Status
        className={cn({ 'animate-pulse': isLoading })}
        status={getStatusTheme(data?.status)}
        isWithoutText
      />
      <Text theme="neutral">
        <Icon variant="GitHub" />
      </Text>
      <VCSAccountLink vcs_connection={vcs_connection} />
    </span>
  )
}
