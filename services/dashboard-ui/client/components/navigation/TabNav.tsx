import { useLocation } from 'react-router'
import { Button } from '@/components/common/Button'
import type { TNavLink } from '@/types'

export interface ITabNav {
  basePath: string
  tabs: TNavLink[]
}

export const TabNav = ({ basePath, tabs }: ITabNav) => {
  const { pathname } = useLocation()

  return (
    <nav
      aria-label="tab navigation"
      className="flex items-center gap-6 border-b w-full"
    >
      {tabs.map((tab) => {
        const href = `${basePath}${tab.path === '/' ? '' : tab.path}`
        const isActive = pathname === href

        return (
          <Button key={tab.path} href={href} isActive={isActive} variant="tab">
            {tab.text}
          </Button>
        )
      })}
    </nav>
  )
}
