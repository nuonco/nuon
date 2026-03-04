import { Badge } from '@/components/common/Badge'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TSandboxRun } from '@/types'

export const SandboxRunSummary = ({
  sandboxRun,
  isLatest = false,
}: {
  sandboxRun: TSandboxRun
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
          {sandboxRun.id}
          {isLatest ? (
            <Badge theme="info" size="sm">
              Latest
            </Badge>
          ) : null}
        </span>
        <Time
          time={sandboxRun.created_at}
          format="relative"
          variant="label"
          theme="neutral"
        />
      </Text>
      <span className="flex items-center gap-4 w-full">
        <Status status={sandboxRun.status_v2?.status} />
        <Text variant="label" theme="neutral">
          {sandboxRun?.created_by?.email}
        </Text>
      </span>
    </span>
  )
}
