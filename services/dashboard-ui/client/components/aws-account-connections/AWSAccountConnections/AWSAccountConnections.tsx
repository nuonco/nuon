import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TAWSAccountConnection } from '@/types'

export const AWSAccountConnections = ({
  connections,
  onSelect,
  onVerify,
  verifyingId,
}: {
  connections: TAWSAccountConnection[]
  onSelect: (connection: TAWSAccountConnection) => void
  onVerify: (connection: TAWSAccountConnection) => void
  verifyingId?: string
}) => (
  <div className="flex flex-col gap-2">
    {connections.length ? (
      connections.map((connection) => (
        <div className="flex items-center gap-1" key={connection.id}>
          <Button
            className="!flex !h-auto !justify-start !p-1 min-w-0 flex-1"
            onClick={() => onSelect(connection)}
            variant="ghost"
          >
            <Status status={connection.verification_status} isWithoutText />
            <Icon variant="AWS" />
            <div className="min-w-0 text-left">
              <Text variant="subtext" weight="strong" className="truncate">
                {connection.name}
              </Text>
              <Text variant="subtext" className="truncate">
                {connection.account_id}
              </Text>
              {connection.last_checked_at ? (
                <Text variant="subtext" theme="neutral">
                  Last checked{' '}
                  <Time
                    time={connection.last_checked_at}
                    format="relative"
                    variant="subtext"
                    shouldTick
                  />
                </Text>
              ) : null}
              {connection.verification_status === 'error' ? (
                <Text variant="subtext" theme="error" className="truncate">
                  {connection.verification_message || 'Verification failed'}
                </Text>
              ) : null}
            </div>
          </Button>
          {connection.role_arn ? (
            <Button
              aria-label={`Re-check ${connection.name}`}
              disabled={verifyingId === connection.id}
              onClick={() => onVerify(connection)}
              size="sm"
              variant="ghost"
            >
              <Icon
                variant={
                  verifyingId === connection.id
                    ? 'Loading'
                    : 'ArrowClockwiseIcon'
                }
              />
              {verifyingId === connection.id ? 'Checking...' : 'Re-check'}
            </Button>
          ) : null}
        </div>
      ))
    ) : (
      <Text variant="subtext">No AWS accounts connected yet.</Text>
    )}
  </div>
)
