import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useOrg } from '@/hooks/use-org'
import { getLogStream } from '@/lib'
import { Loading } from '@/components/common/Loading'
import type { TLogStream } from '@/types'

type LogStreamContextValue = {
  logStream: TLogStream
  refresh: () => void
}

export const LogStreamContext = createContext<
  LogStreamContextValue | undefined
>(undefined)

export function LogStreamProvider({
  children,
  logStreamId,
  pollInterval = 20000,
  shouldPoll = false,
}: {
  children: ReactNode
  logStreamId: string
  pollInterval?: number
  shouldPoll?: boolean
}) {
  const { org } = useOrg()
  const { data: logStream, isLoading, refetch } = useQuery({
    queryKey: ['log-stream', org.id!, logStreamId],
    queryFn: () => getLogStream({ orgId: org.id!, logStreamId }),
    refetchInterval: shouldPoll ? pollInterval : false,
    enabled: !!org.id && !!logStreamId,
  })

  if (isLoading || !logStream) return <Loading />

  return (
    <LogStreamContext.Provider value={{ logStream, refresh: refetch }}>
      {children}
    </LogStreamContext.Provider>
  )
}
