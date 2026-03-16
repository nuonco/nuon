import React from 'react'
import { useHashScroll } from '@/hooks/use-hash-scroll'
import { useSidebar } from '@/hooks/use-sidebar'
import type { TNuonVersion } from '@/types'
import { cn } from '@/utils/classnames'
import { MainSidebar } from './MainSidebar'

interface IMainLayout {
  children: React.ReactNode
  versions: TNuonVersion
  hideOrgContent?: boolean
}

export const MainLayout = ({
  children,
  versions,
  hideOrgContent,
}: IMainLayout) => {
  useHashScroll()
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
        <MainSidebar versions={versions} hideOrgContent={hideOrgContent} />
        <div className="w-full">{children}</div>
      </div>
    </div>
  )
}
