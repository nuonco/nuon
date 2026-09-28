import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import type { TSandboxRun } from '@/types'
import { humanize } from '@/utils/string-utils'
import { useInstallLink } from '@/hooks/use-install-path'

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
  const installLink = useInstallLink()
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
                  href={installLink({ orgId: orgId, installId: installId, suffix: `/sandbox/runs/${run?.id}` })}
                  variant="inline"
                >
                  {humanize(run?.run_type)}
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
