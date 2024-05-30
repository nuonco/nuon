'use client'

import React, {
  FunctionComponent,
  ReactElement,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react'
import type { TBuild } from '@/types'
import { POLL_DURATION } from '@/utils'

type TBuildContext = {
  isFetching: boolean
  build: TBuild
}

export const BuildContext = createContext<TBuildContext | null>(null)

export const BuildProvider: FunctionComponent<{
  children?: ReactElement
  initBuild: TBuild
  shouldPoll?: boolean
}> = ({ children, initBuild, shouldPoll = false }) => {
  const [isFetching, setIsFetching] = useState<boolean>(false)
  const [build, setBuild] = useState<TBuild>(initBuild)

  useEffect(() => {
    const fetchBuild = () => {
      setIsFetching(true)
      fetch(
        `/api/${build?.org_id}/components/${build.component_id}/builds/${build.id}`
      )
        .then((res) =>
          res.json().then((ins) => {
            setBuild(ins)
            setIsFetching(false)
          })
        )
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollBuild = setInterval(fetchBuild, POLL_DURATION)
      return () => clearInterval(pollBuild)
    }
  }, [build, shouldPoll])

  return (
    <BuildContext.Provider
      value={{
        build,
        isFetching,
      }}
    >
      {children}
    </BuildContext.Provider>
  )
}

export const useBuildContext = () => {
  const context = useContext(BuildContext)

  if (!context) {
    throw new Error(
      'useBuildContext() may only be used in the context of a <BuildProvider> component.'
    )
  }

  return context
}
