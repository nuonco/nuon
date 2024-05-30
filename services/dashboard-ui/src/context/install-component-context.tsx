'use client'

import React, {
  FunctionComponent,
  ReactElement,
  createContext,
  useContext,
  useEffect,
  useState,
} from 'react'
import type { TInstallComponent } from '@/types'
import { POLL_DURATION } from '@/utils'

type TInstallComponentContext = {
  isFetching: boolean
  installComponent: TInstallComponent
}

export const InstallComponentContext =
  createContext<TInstallComponentContext | null>(null)

export const InstallComponentProvider: FunctionComponent<{
  children?: ReactElement
  initInstallComponent: TInstallComponent
  shouldPoll?: boolean
}> = ({ children, initInstallComponent, shouldPoll = false }) => {
  const [isFetching, setIsFetching] = useState<boolean>(false)
  const [installComponent, setInstallComponent] =
    useState<TInstallComponent>(initInstallComponent)

  useEffect(() => {
    const fetchInstallComponent = () => {
      setIsFetching(true)
      fetch(
        `/api/${installComponent?.org_id}/installs/${installComponent?.install_id}/components/${installComponent.id}`
      )
        .then((res) =>
          res.json().then((comp) => {
            console.log('install component?', comp)
            setInstallComponent(comp)
            setIsFetching(false)
          })
        )
        .catch(console.error)
    }

    if (shouldPoll) {
      const pollInstallComponent = setInterval(
        fetchInstallComponent,
        POLL_DURATION
      )
      return () => clearInterval(pollInstallComponent)
    }
  }, [installComponent, shouldPoll])

  return (
    <InstallComponentContext.Provider
      value={{
        installComponent,
        isFetching,
      }}
    >
      {children}
    </InstallComponentContext.Provider>
  )
}

export const useInstallComponentContext = () => {
  const context = useContext(InstallComponentContext)

  if (!context) {
    throw new Error(
      'useInstallComponentContext() may only be used in the context of a <InstallComponentProvider> component.'
    )
  }

  return context
}
