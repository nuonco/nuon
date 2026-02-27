import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Timeline, type ITimeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallSandboxRuns } from '@/lib'
import type { TSandboxRun } from '@/types'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'

interface ISandboxRunsTimeline
  extends Omit<ITimeline<TSandboxRun>, 'events' | 'renderEvent'> {
  initRuns: Array<TSandboxRun>
  pollInterval?: number
  shouldPoll?: boolean
}

export const SandboxRunsTimeline = ({
  initRuns,
  pagination,
  shouldPoll = false,
  pollInterval = 20000,
}: ISandboxRunsTimeline) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: runs } = useQuery<TSandboxRun[]>({
    queryKey: ['install-sandbox-runs', org?.id, install?.id, pagination?.offset],
    queryFn: () =>
      getInstallSandboxRuns({
        orgId: org.id,
        installId: install.id,
        limit: 10,
        offset: pagination?.offset,
      }),
    initialData: initRuns,
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!install?.id,
  })

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
                  href={`/${org.id}/installs/${install?.id}/sandbox/${run?.id}`}
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
