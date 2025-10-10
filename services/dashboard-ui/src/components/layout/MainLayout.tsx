'use client'

import React from 'react'
import { useSidebar } from '@/hooks/use-sidebar'
import { cn } from '@/utils/classnames'
import { MainSidebar } from './MainSidebar'

interface IMainLayout {
  children: React.ReactNode
}

export const MainLayout = ({ children }: IMainLayout) => {
  const { isSidebarOpen } = useSidebar()

  return (
    <div className="w-screen overflow-hidden">
      <div
        className={cn(
          'flex h-screen w-[200vw] transition-transform duration-fast ease-cubic md:w-screen md:transition-none',
          {
            'md:translate-x-0 -translate-x-[100vw]': isSidebarOpen,
          }
        )}
      >
        <MainSidebar />
        {children}
      </div>
    </div>
  )
}
