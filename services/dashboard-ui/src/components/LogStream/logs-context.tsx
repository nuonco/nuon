'use client'

import React, {
  type FC,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react'
import { parseOTELLog } from './helpers'
import type { TLogRecord } from './types'
import type { TLogStream } from '@/types'

type TLogError = Record<string | 'message', string>

interface ILogsContext {
  isLoading: boolean
  logs?: Array<TLogRecord>
  error?: TLogError
}

const LogsContext = createContext<ILogsContext>({ isLoading: true })

interface ILogsProvider {
  children: React.ReactNode
  logStream?: TLogStream
  shouldPoll?: boolean
}

export const LogsProvider: FC<ILogsProvider> = ({
  children,
  logStream,
  shouldPoll = false,
}) => {
  const [isLoading, setIsLoading] = useState(true)
  const [nextPage, setNextPage] = useState('0')
  const [logs, updateLogs] = useState([])
  const [error, setError] = useState()

  const fetchLogs = () => {
    fetch(`/api/${logStream?.org_id}/log-streams/${logStream?.id}/logs`, {
      headers: { 'X-NUON-API-Offset': nextPage },
    })
      .then((res) => {
        const next = res.headers.get('x-nuon-api-next') || '0'

        if (next !== nextPage) setNextPage(next)
        res.json().then((l) => {
          setIsLoading(false)
          updateLogs((state) =>
            [...state, ...parseOTELLog(l)].filter(
              (log, i, arr) => i === arr.findIndex((lr) => lr?.id === log?.id)
            )
          )
        })
      })
      .catch((err) => setError(err))
  }

  useEffect(() => {
    if (!logStream?.open) {
      fetchLogs()
    }
  }, [])

  useEffect(() => {
    if (!logStream?.open && nextPage !== '0') {
      fetchLogs()
    }
  }, [nextPage])

  useEffect(() => {
    if (shouldPoll) {
      const pollLogs = setInterval(fetchLogs, 1000)
      if (!logStream?.open) {
        clearInterval(pollLogs)
      }

      return () => clearInterval(pollLogs)
    }
  }, [logStream])

  return (
    <LogsContext.Provider value={{ error, isLoading, logs }}>
      {children}
    </LogsContext.Provider>
  )
}

export const useLogs = (): ILogsContext => {
  return useContext(LogsContext)
}
