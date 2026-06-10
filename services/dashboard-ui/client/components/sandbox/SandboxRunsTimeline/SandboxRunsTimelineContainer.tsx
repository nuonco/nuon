import { useSearchParams } from 'react-router'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSSETimelineQuery } from '@/hooks/use-sse-timeline-query'
import { getInstallSandboxRuns } from '@/lib'
import { SandboxRunsTimeline } from './SandboxRunsTimeline'

const LIMIT = 10

interface ISandboxRunsTimelineContainer {
  pollInterval?: number
  shouldPoll?: boolean
}

export const SandboxRunsTimelineContainer = ({
  shouldPoll = false,
  pollInterval = 20000,
}: ISandboxRunsTimelineContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result } = useSSETimelineQuery({
    sseUrl:
      org?.id && install?.id
        ? `/api/orgs/${org.id}/installs/${install.id}/sandbox-runs/sse?limit=${LIMIT}&offset=${offset}`
        : undefined,
    queryKey: ['install-sandbox-runs', org?.id, install?.id, offset],
    queryFn: () =>
      getInstallSandboxRuns({
        orgId: org.id,
        installId: install.id,
        limit: LIMIT,
        offset,
      }),
    enabled: !!org?.id && !!install?.id,
    shouldPoll,
    pollInterval,
    eventName: 'sandbox-runs',
  })

  const runs = result?.data ?? []
  const pagination = result?.pagination
    ? { hasNext: result.pagination.hasNext, offset, limit: LIMIT }
    : { hasNext: false, offset, limit: LIMIT }

  return (
    <SandboxRunsTimeline
      runs={runs}
      pagination={pagination}
      orgId={org?.id}
      installId={install?.id}
    />
  )
}
