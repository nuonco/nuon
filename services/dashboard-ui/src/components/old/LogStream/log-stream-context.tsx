'use client'

import React, {
  type FC,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react'
import { LogsProvider } from './logs-context'
import type { TLogStream } from '@/types'

type TLogError = Record<string | 'message', string>

interface ILogStreamContext {
  logStream?: TLogStream
  error?: TLogError
}

const LogStreamContext = createContext<ILogStreamContext>({})

interface ILogStreamProvider {
  children: React.ReactNode
  initLogStream: TLogStream
}

export const LogStreamProvider: FC<ILogStreamProvider> = ({
  children,
  initLogStream,
}) => {
  const [error, setError] = useState<TLogError>()
  const [logStream, updateStream] = useState<TLogStream>(initLogStream)

  useEffect(() => {
    if (initLogStream?.id === '') {
      setError({
        message: 'Log stream not created yet.',
      })
    }
  }, [])

  useEffect(() => {
    const refreshLogStream = () => {
      fetch(
        `/api/orgs/${initLogStream?.org_id}/log-streams/${initLogStream?.id}`
      )
        .then((res) =>
          res.json().then(({ data, error }) => {
            if (error) {
              setError(error?.error)
            } else {
              updateStream(data)
            }
          })
        )
        .catch((err) => setError(err))
    }

    if (logStream?.open || logStream?.id === '') {
      const pollStream = setInterval(refreshLogStream, 5000)

      return () => {
        clearInterval(pollStream)
      }
    }
  }, [logStream])

  return (
    <LogStreamContext.Provider
      value={{
        error,
        logStream,
      }}
    >
      <LogsProvider
        logStream={logStream}
        shouldPoll={logStream?.open}
        logStreamError={error}
      >
        {children}
      </LogsProvider>
    </LogStreamContext.Provider>
  )
}

export const useLogStream = (): ILogStreamContext => {
  return useContext(LogStreamContext)
}
