import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getLogStreamSpans } from '@/lib'

export function usePrefetchLogStreamSpans(logStreamId?: string) {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const orgId = org?.id

  useEffect(() => {
    if (!orgId || !logStreamId) return
    queryClient.prefetchQuery({
      queryKey: ['log-stream-spans', orgId, logStreamId],
      queryFn: () => getLogStreamSpans({ orgId, logStreamId }),
    })
  }, [orgId, logStreamId, queryClient])
}
