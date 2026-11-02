import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TCloudConnectionSummary } from '@/types'

export const PrincipalSharing = ({
  connections,
  orgId,
}: {
  connections: TCloudConnectionSummary[]
  orgId?: string
}) =>
  connections.length > 0 ? (
    <div className="flex min-w-0 flex-wrap items-center gap-2">
      <Text variant="subtext">Also used by:</Text>
      {connections.map((connection) =>
        orgId ? (
          <Link
            key={connection.id}
            href={`/${orgId}/settings/cloud-connections/${connection.id}`}
          >
            <Text family="mono" className="break-all">
              {connection.name}
            </Text>
          </Link>
        ) : (
          <Text key={connection.id} family="mono" className="break-all">
            {connection.name}
          </Text>
        )
      )}
    </div>
  ) : null
