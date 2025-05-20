'use client'

import React, { type FC, createContext, useContext, useState } from 'react'
import { setDashboardSidebarCookie } from '@/stratus/actions'

interface IDashboardContext {
  isSidebarOpen?: boolean
  closeSidebar?: () => void
  openSidebar?: () => void
  toggleSidebar?: () => void
}

const DashboardContext = createContext<IDashboardContext>({})

export const DashboardProvider: FC<{
  children: React.ReactNode
  initIsSidebarOpen?: boolean
}> = ({ children, initIsSidebarOpen = false }) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(initIsSidebarOpen)

  function closeSidebar() {
    setDashboardSidebarCookie(false)
    setIsSidebarOpen(false)
  }

  function openSidebar() {
    setDashboardSidebarCookie(true)
    setIsSidebarOpen(true)
  }

  function toggleSidebar() {
    setDashboardSidebarCookie(!isSidebarOpen)
    setIsSidebarOpen(!isSidebarOpen)
  }

  return (
    <DashboardContext.Provider
      value={{
        isSidebarOpen,
        closeSidebar,
        openSidebar,
        toggleSidebar,
      }}
    >
      {children}
    </DashboardContext.Provider>
  )
}

export const useDashboard = (): IDashboardContext =>
  useContext(DashboardContext)
