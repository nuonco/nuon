'use client'

import React, {
  FunctionComponent,
  ReactElement,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react'
import type { TSandboxRun } from '@/types'
import { POLL_DURATION } from '@/utils'

type TSandboxRunContext = {
  isFetching: boolean
  run: TSandboxRun
}

export const SandboxRunContext = createContext<TSandboxRunContext | null>(null)

export const SandboxRunProvider: FunctionComponent<{
  children?: ReactElement
  initRun: TSandboxRun
  shouldPoll?: boolean
}> = ({ children, initRun, shouldPoll = false }) => {
  const [isFetching, setIsFetching] = useState<boolean>(false)
  const [run, setSandboxRun] = useState<TSandboxRun>(initRun)

  useEffect(() => {
    const fetchSandboxRun = () => {
      setIsFetching(true)
      fetch(`/api/${run.org_id}/installs/${run.install_id}/runs/${run.id}`)
        .then((res) =>
          res.json().then((ins) => {
            setSandboxRun(ins)
            setIsFetching(false)
          })
        )
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollSandboxRun = setInterval(fetchSandboxRun, POLL_DURATION)
      return () => clearInterval(pollSandboxRun)
    }
  }, [run, shouldPoll])

  return (
    <SandboxRunContext.Provider
      value={{
        run,
        isFetching,
      }}
    >
      {children}
    </SandboxRunContext.Provider>
  )
}

export const useSandboxRunContext = () => {
  const context = useContext(SandboxRunContext)

  if (!context) {
    throw new Error(
      'useSandboxRunContext() may only be used in the context of a <SandboxRunProvider> component.'
    )
  }

  return context
}
