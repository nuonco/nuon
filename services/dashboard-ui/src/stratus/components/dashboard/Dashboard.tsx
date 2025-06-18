'use client'

import classNames from 'classnames'
import React, { type FC } from 'react'
import { useDashboard } from '@/stratus/context'
import { Sidebar } from './Sidebar'
import './Dashboard.css'

interface IDashboard {
  children: React.ReactNode
}

export const Dashboard: FC<IDashboard> = ({ children }) => {
  const { isSidebarOpen } = useDashboard()

  return (
    <div
      className={classNames('dashboard', {
        'is-open': isSidebarOpen,
      })}
    >
      <Sidebar />
      {children}     
      <div id="surface-root" />
    </div>
  )
}
