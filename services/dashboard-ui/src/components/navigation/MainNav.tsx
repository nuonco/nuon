import { Text } from '@/components/common/Text'
import { useOrg } from '@/hooks/use-org'
import { useSidebar } from '@/hooks/use-sidebar'
import { cn } from '@/utils/classnames'
import { MainNavLink } from './MainNavLink'
import { MAIN_LINKS, SETTINGS_LINKS, SUPPORT_LINKS } from './main-nav-links'

const NavLabel = ({ children }: { children: string }) => {
  const { isSidebarOpen } = useSidebar()
  return (
    <Text
      variant="subtext"
      className={cn(
        'px-2 overflow-hidden whitespace-nowrap duration-fast transition-all ease-cubic',
        {
          'md:h-[17px] md:opacity-100': isSidebarOpen,
          'md:h-[0px] md:opacity-0': !isSidebarOpen,
        }
      )}
    >
      {children}
    </Text>
  )
}

const Divider = () => {
  const { isSidebarOpen } = useSidebar()
  return (
    <hr
      className={cn('transition-opacity duration-fast ease-cubic', {
        'md:opacity-100': !isSidebarOpen,
        'md:opacity-0': isSidebarOpen,
      })}
    />
  )
}

export const MainNav = () => {
  const { org } = useOrg()
  const basePath = `/${org.id}`
  return (
    <nav className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        {MAIN_LINKS.map((link) => (
          <MainNavLink key={link.text} basePath={basePath} {...link} />
        ))}
      </div>

      <Divider />

      {org?.features?.['org-settings'] ? (
        <div className="flex flex-col gap-1">
          <NavLabel>Settings</NavLabel>

          {SETTINGS_LINKS.map((link) => (
            <MainNavLink key={link.text} basePath={basePath} {...link} />
          ))}
        </div>
      ) : null}

      <Divider />

      {org?.features?.['org-support'] ? (
        <div className="flex flex-col gap-1">
          <NavLabel>Resources</NavLabel>

          {SUPPORT_LINKS.map((link) => (
            <MainNavLink key={link.text} basePath={basePath} {...link} />
          ))}
        </div>
      ) : null}
    </nav>
  )
}
