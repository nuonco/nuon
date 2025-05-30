'use client'

import React, { type FC, createContext, useContext, useState } from 'react'
import { setPageNavCookie } from '@/stratus/actions'

interface IPageContext {
  isPageNavOpen?: boolean
  closePageNav?: () => void
  openPageNav?: () => void
  togglePageNav?: () => void
}

const PageContext = createContext<IPageContext>({})

export const PageProvider: FC<{
  children: React.ReactNode
  initIsPageNavOpen?: boolean
}> = ({ children, initIsPageNavOpen = true }) => {
  const [isPageNavOpen, setIsPageNavOpen] = useState(initIsPageNavOpen)

  function closePageNav() {
    setPageNavCookie(false)
    setIsPageNavOpen(false)
  }

  function openPageNav() {
    setPageNavCookie(true)
    setIsPageNavOpen(true)
  }

  function togglePageNav() {
    setPageNavCookie(!isPageNavOpen)
    setIsPageNavOpen(!isPageNavOpen)
  }

  return (
    <PageContext.Provider
      value={{
        isPageNavOpen,
        closePageNav,
        openPageNav,
        togglePageNav,
      }}
    >
      {children}
    </PageContext.Provider>
  )
}

export const usePage = (): IPageContext => useContext(PageContext)
