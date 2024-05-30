'use client'

import React, {
  FunctionComponent,
  ReactElement,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react'
import type { TInstallDeploy } from '@/types'
import { POLL_DURATION } from '@/utils'

type TInstallDeployContext = {
  isFetching: boolean
  deploy: TInstallDeploy
}

export const InstallDeployContext = createContext<TInstallDeployContext | null>(
  null
)

export const InstallDeployProvider: FunctionComponent<{
  children?: ReactElement
  initDeploy: TInstallDeploy
  shouldPoll?: boolean
}> = ({ children, initDeploy, shouldPoll = false }) => {
  const [isFetching, setIsFetching] = useState<boolean>(false)
  const [deploy, setInstallDeploy] = useState<TInstallDeploy>(initDeploy)

  useEffect(() => {
    const fetchInstallDeploy = () => {
      setIsFetching(true)
      fetch(
        `/api/${deploy?.org_id}/installs/${deploy?.install_id}/deploys/${deploy?.id}`
      )
        .then((res) =>
          res.json().then((ins) => {
            setInstallDeploy(ins)
            setIsFetching(false)
          })
        )
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollInstallDeploy = setInterval(fetchInstallDeploy, POLL_DURATION)
      return () => clearInterval(pollInstallDeploy)
    }
  }, [deploy, shouldPoll])

  return (
    <InstallDeployContext.Provider
      value={{
        deploy,
        isFetching,
      }}
    >
      {children}
    </InstallDeployContext.Provider>
  )
}

export const useInstallDeployContext = () => {
  const context = useContext(InstallDeployContext)

  if (!context) {
    throw new Error(
      'useInstallDeployContext() may only be used in the context of a <InstallDeployProvider> component.'
    )
  }

  return context
}
