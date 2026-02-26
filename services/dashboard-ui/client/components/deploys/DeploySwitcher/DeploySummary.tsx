import { Badge } from '@/components/common/Badge'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TDeploy } from '@/types'

export const DeploySummary = ({
  deploy,
  isLatest = false,
}: {
  deploy: TDeploy
  isLatest?: boolean
}) => {
  return (
    <span className="flex flex-col w-full">
      <Text
        className="flex items-center justify-between"
        variant="subtext"
        weight="strong"
      >
        <span className="flex items-center gap-4">
          {deploy.id}
          {isLatest ? (
            <Badge theme="info" size="sm">
              Latest
            </Badge>
          ) : null}
        </span>
        <Time
          time={deploy.created_at}
          format="relative"
          variant="label"
          theme="neutral"
        />
      </Text>
      <span className="flex items-center gap-4 w-full">
        <Status status={deploy.status_v2?.status} />
        <Text variant="label" theme="neutral">
          {deploy?.created_by?.email}
        </Text>
      </span>
    </span>
  )
}
