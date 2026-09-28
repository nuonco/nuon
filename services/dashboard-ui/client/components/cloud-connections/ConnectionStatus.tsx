import { Status } from '@/components/common/Status'
import { Time } from '@/components/common/Time'
import { Tooltip } from '@/components/common/Tooltip'
import type { TCloudConnection } from '@/types'

export const ConnectionStatus = ({
  connection,
  isVerifying = false,
}: {
  connection?: TCloudConnection
  isVerifying?: boolean
}) => (
  <Tooltip
    position="bottom"
    tipContentClassName="max-w-xs whitespace-normal"
    tipContent={
      isVerifying ? (
        'Verification is in progress.'
      ) : connection?.status === 'error' ? (
        connection.status_message || 'Verification failed.'
      ) : connection?.last_verified_at ? (
        <>
          Last verified{' '}
          <Time time={connection.last_verified_at} format="long-datetime" />
        </>
      ) : (
        'This connection has not been verified.'
      )
    }
  >
    <span tabIndex={0}>
      <Status
        loading={!connection}
        variant="badge"
        status={isVerifying ? 'pending' : connection?.status}
      />
    </span>
  </Tooltip>
)
