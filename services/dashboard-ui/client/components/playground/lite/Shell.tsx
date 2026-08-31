import { useMemo, useState, type ReactNode } from 'react'
import { Sidebar } from './Sidebar'
import { StatusBar } from './StatusBar'
import { ShellContext } from './shell-context'
import type { INavItem } from './types'

export interface IShell {
  primaryNav: INavItem[]
  secondaryNav?: INavItem[]
  children: ReactNode
}

export const Shell = ({ primaryNav, secondaryNav, children }: IShell) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)
  const [showText, setShowText] = useState(false)

  const shell = useMemo(
    () => ({
      isSidebarOpen,
      toggleSidebar: () => setIsSidebarOpen((open) => !open),
      showText,
      toggleText: () => setShowText((show) => !show),
    }),
    [isSidebarOpen, showText]
  )

  return (
    <ShellContext.Provider value={shell}>
      <div className="flex flex-col h-screen w-full overflow-hidden">
        <div className="flex flex-1 min-h-0">
          <Sidebar primaryNav={primaryNav} secondaryNav={secondaryNav} />

          <main className="flex flex-1 min-w-0 flex-col gap-8 p-4 overflow-y-auto">
            {children}
          </main>
        </div>

        <StatusBar />
      </div>
    </ShellContext.Provider>
  )
}
