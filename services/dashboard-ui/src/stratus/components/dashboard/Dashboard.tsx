'use client'

import React from 'react'
import { cn } from '@/stratus/components/helpers'
import { useDashboard } from '@/stratus/context'
import { Sidebar } from './Sidebar'
import './Dashboard.css'

interface IDashboard {
  children: React.ReactNode
}

export const Dashboard = ({ children }: IDashboard) => {
  const { isSidebarOpen } = useDashboard()

  return (
    <div
      className={cn('dashboard', {
        'is-open': isSidebarOpen,
      })}
    >
      <Sidebar />
      {children}
      <div id="surface-root" />
    </div>
  )
}
