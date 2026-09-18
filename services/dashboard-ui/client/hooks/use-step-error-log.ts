import { useQuery } from '@tanstack/react-query'
import { getLogStreamLogs } from '@/lib/ctl-api/log-streams'
import { useOrg } from '@/hooks/use-org'
import type { TOTELLog, TWorkflowStep } from '@/types'

export type TStepErrorLog = {
  log: TOTELLog
  message: string
  detail?: string
}

const pickDetail = (log?: TOTELLog) => {
  const attributes = log?.log_attributes ?? {}
  return attributes.errorVerbose || attributes.error || undefined
}

export const useStepErrorLog = (
  step: TWorkflowStep,
  { enabled = true }: { enabled?: boolean } = {}
): TStepErrorLog | undefined => {
  const { org } = useOrg()
  const logStreamId = step?.log_stream?.id

  const { data } = useQuery({
    queryKey: ['step-error-log', logStreamId],
    queryFn: () =>
      getLogStreamLogs({
        logStreamId: logStreamId as string,
        orgId: org?.id as string,
        order: 'desc',
        filters: { severity_text: ['error'] },
      }),
    enabled: enabled && Boolean(logStreamId) && Boolean(org?.id),
    staleTime: 60_000,
  })

  const log = data?.[0]
  if (!log) return undefined

  const detail = pickDetail(log)
  const message = log.body || detail || ''
  if (!message) return undefined

  return { log, message, detail: detail === message ? undefined : detail }
}
