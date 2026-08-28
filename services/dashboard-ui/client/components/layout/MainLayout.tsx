import React, { useEffect, useRef } from 'react'
import { useLocation } from 'react-router'
import { useHashScroll } from '@/hooks/use-hash-scroll'
import { useIsDesktop } from '@/hooks/use-is-desktop'
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
  const { isSidebarOpen, openSidebar, toggleSidebar } = useSidebar()
  const isDesktop = useIsDesktop()
  const { pathname } = useLocation()
  const prevPathname = useRef(pathname)

  useEffect(() => {
    if (prevPathname.current === pathname) return
    prevPathname.current = pathname
    if (!isDesktop && !isSidebarOpen) openSidebar?.()
  }, [pathname, isDesktop, isSidebarOpen, openSidebar])

  return (
    <div className="flex h-screen w-full overflow-hidden">
      <MainSidebar versions={versions} hideOrgContent={hideOrgContent} />
      <div
        className={cn(
          'fixed inset-0 z-40 bg-black/40 transition-opacity duration-fast md:hidden',
          !isSidebarOpen ? 'opacity-100' : 'opacity-0 pointer-events-none'
        )}
        onClick={toggleSidebar}
        aria-hidden="true"
      />
      <div className="flex-1 min-w-0 flex flex-col overflow-hidden">
        {children}
      </div>
    </div>
  )
}
