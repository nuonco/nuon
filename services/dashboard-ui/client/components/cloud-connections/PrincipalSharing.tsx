import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TCloudConnectionSummary } from '@/types'

export const PrincipalSharing = ({
  connections,
  orgId,
  sharedPermissions = false,
}: {
  connections: TCloudConnectionSummary[]
  orgId?: string
  sharedPermissions?: boolean
}) =>
  connections.length > 0 ? (
    <div className="flex min-w-0 flex-col gap-1">
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        <Text variant="subtext">Also used by:</Text>
        {connections.map((connection, index) => (
          <span key={connection.id} className="inline-flex items-center">
            {orgId ? (
              <Link
                href={`/${orgId}/settings/cloud-connections/${connection.id}`}
                className="!font-mono break-all"
              >
                {connection.name}
              </Link>
            ) : (
              <Text family="mono" className="break-all">
                {connection.name}
              </Text>
            )}
            {index < connections.length - 1 ? (
              <Text variant="subtext">,</Text>
            ) : null}
          </span>
        ))}
      </div>
      {sharedPermissions && (
        <Text variant="subtext" theme="neutral">
          These connections share this principal's permissions. Revoking the
          principal cuts off all of them.
        </Text>
      )}
    </div>
  ) : null
