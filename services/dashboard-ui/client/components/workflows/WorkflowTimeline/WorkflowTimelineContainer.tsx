import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallWorkflows } from '@/lib'
import { WorkflowTimeline, WorkflowTimelineSkeleton } from './WorkflowTimeline'

export { WorkflowTimelineSkeleton }

const LIMIT = 10

interface IWorkflowTimelineContainer {
  installId: string
  pollInterval?: number
  shouldPoll?: boolean
  type?: string
  planonly?: boolean
}

export const WorkflowTimelineContainer = ({
  installId,
  shouldPoll = false,
  pollInterval = 20000,
  planonly = true,
  type = '',
}: IWorkflowTimelineContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()
  const offset = Number(searchParams.get('offset') ?? 0)

  const { data: result } = useQuery({
    queryKey: ['install-workflows', org?.id, installId, offset, planonly, type],
    queryFn: () =>
      getInstallWorkflows({
        orgId: org.id,
        installId,
        limit: LIMIT,
        offset,
        planonly,
        type,
      }),
    refetchOnMount: 'always',
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org?.id && !!installId,
  })

  const workflows = result?.data ?? []
  const pagination = result?.pagination
    ? { hasNext: result.pagination.hasNext, offset, limit: LIMIT }
    : { hasNext: false, offset, limit: LIMIT }

  return (
    <WorkflowTimeline
      workflows={workflows}
      pagination={pagination}
      orgId={org?.id}
      installId={installId}
      install={install}
    />
  )
}
