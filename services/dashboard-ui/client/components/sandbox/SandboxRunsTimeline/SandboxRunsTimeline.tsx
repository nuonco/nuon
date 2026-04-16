import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import type { TSandboxRun } from '@/types'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'

interface ISandboxRunsTimeline {
  runs: TSandboxRun[]
  pagination: { hasNext: boolean; offset: number; limit: number }
  orgId: string
  installId: string
}

export const SandboxRunsTimeline = ({
  runs,
  pagination,
  orgId,
  installId,
}: ISandboxRunsTimeline) => {
  return (
    <Timeline<TSandboxRun>
      events={runs}
      pagination={pagination}
      renderEvent={(run) => {
        return (
          <TimelineEvent
            key={run.id}
            caption={<ID>{run?.id}</ID>}
            createdAt={run?.created_at}
            status={run?.status}
            title={
              <span className="flex items-center gap-2">
                <Link
                  href={`/${orgId}/installs/${installId}/sandbox/runs/${run?.id}`}
                >
                  {toSentenceCase(snakeToWords(run?.run_type))}
                </Link>
                {run?.status_v2?.status === 'drifted' ? (
                  <Badge variant="code" size="sm">
                    drift scan
                  </Badge>
                ) : null}
              </span>
            }
            underline={<>Run by: {run?.created_by?.email}</>}
          />
        )
      }}
    />
  )
}
