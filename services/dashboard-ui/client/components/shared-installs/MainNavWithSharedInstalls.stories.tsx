export default {
  title: 'Playground / Customer portal / Main nav',
}

import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { MainNavLink } from '@/components/navigation/MainNavLink'
import {
  MAIN_LINKS,
  SETTINGS_LINKS,
  SUPPORT_LINKS,
} from '@/components/navigation/main-nav-links'
import type { TNavLink } from '@/types'

const mockOrgId = 'org-1'

const SHARED_INSTALLS_LINK: TNavLink = {
  iconVariant: 'ShareIcon',
  path: `/shared-installs`,
  text: 'Shared installs',
  shortcut: 'g h',
}

const NavLabel = ({ children }: { children: string }) => (
  <Text
    variant="subtext"
    className="px-2 overflow-hidden whitespace-nowrap"
  >
    {children}
  </Text>
)

/** Playground-only MainNav preview — does not change production MAIN_LINKS. */
const PlaygroundMainNav = ({ isSidebarOpen = true }: { isSidebarOpen?: boolean }) => {
  const basePath = `/${mockOrgId}`
  const mainLinks = [...MAIN_LINKS, SHARED_INSTALLS_LINK]

  return (
    <nav className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        {mainLinks.map((link) => (
          <MainNavLink key={link.text} basePath={basePath} {...link} />
        ))}
      </div>

      <hr
        className={cn('transition-opacity', {
          'opacity-100': !isSidebarOpen,
          'opacity-0': isSidebarOpen,
        })}
      />

      <div className="flex flex-col gap-1">
        <NavLabel>Manage</NavLabel>
        {SETTINGS_LINKS.map((link) => (
          <MainNavLink key={link.text} basePath={basePath} {...link} />
        ))}
      </div>

      <hr
        className={cn('transition-opacity', {
          'opacity-100': !isSidebarOpen,
          'opacity-0': isSidebarOpen,
        })}
      />

      <div className="flex flex-col gap-1">
        <NavLabel>Resources</NavLabel>
        {SUPPORT_LINKS.map((link) => (
          <MainNavLink key={link.text} basePath={basePath} {...link} />
        ))}
      </div>
    </nav>
  )
}

export const WithSharedInstalls = () => (
  <div className="w-[248px] p-4">
    <PlaygroundMainNav />
  </div>
)
