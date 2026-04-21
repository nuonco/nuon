import { Status } from '@/components/common/Status'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { TVCSConnection } from '@/types'
import type { TTheme } from '@/types'
import { cn } from '@/utils/classnames'
import { VCSAccountLink } from '../VCSAccountLink'

interface IVCSConnectionItem {
  vcs_connection: TVCSConnection
  statusTheme?: TTheme
  isLoadingStatus?: boolean
}

export const VCSConnectionItem = ({
  vcs_connection,
  statusTheme,
  isLoadingStatus,
}: IVCSConnectionItem) => {
  return (
    <span className="!flex gap-2 items-center w-full">
      <Status
        className={cn({ 'animate-pulse': isLoadingStatus })}
        status={statusTheme}
        isWithoutText
      />
      <Text theme="neutral">
        <Icon variant="GitHub" />
      </Text>
      <VCSAccountLink vcs_connection={vcs_connection} />
    </span>
  )
}

interface IVCSConnections {
  vcsConnections: TVCSConnection[]
  statusMap?: Record<string, { theme?: TTheme; isLoading?: boolean }>
}

export const VCSConnections = ({
  vcsConnections,
  statusMap = {},
}: IVCSConnections) => {
  return (
    <>
      {vcsConnections?.length &&
        vcsConnections?.map((vcs) => (
          <Text
            key={vcs?.id}
            className="!flex gap-2 justify-between w-full"
            variant="subtext"
          >
            <VCSConnectionItem
              vcs_connection={vcs}
              statusTheme={statusMap[vcs?.id]?.theme}
              isLoadingStatus={statusMap[vcs?.id]?.isLoading}
            />
          </Text>
        ))}
    </>
  )
}
