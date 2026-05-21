import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { MainNavLink } from '../MainNavLink'
import { MAIN_LINKS, SETTINGS_LINKS, SLACK_LINK, SUPPORT_LINKS } from '../main-nav-links'
import type { TOrg } from '@/types'

interface IMainNav {
  org: TOrg
  isSidebarOpen: boolean
  hasOrgDashboard: boolean
  hasOrgSettings: boolean
  hasSlack: boolean
  hasCustomerPortal: boolean
  customerPortalUrl: string
}

const NavLabel = ({ children, isSidebarOpen }: { children: string; isSidebarOpen: boolean }) => (
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

const Divider = ({ isSidebarOpen }: { isSidebarOpen: boolean }) => (
  <hr
    className={cn('transition-opacity duration-fast ease-cubic', {
      'md:opacity-100': !isSidebarOpen,
      'md:opacity-0': isSidebarOpen,
    })}
  />
)

export const MainNav = ({
  org,
  isSidebarOpen,
  hasOrgDashboard,
  hasOrgSettings,
  hasSlack,
  hasCustomerPortal,
  customerPortalUrl,
}: IMainNav) => {
  const basePath = `/${org.id}`
  const mainLinks = hasCustomerPortal
    ? [
        ...MAIN_LINKS,
        {
          iconVariant: 'UsersIcon' as const,
          path: customerPortalUrl,
          text: 'Customer Portal',
          isExternal: true,
        },
      ]
    : MAIN_LINKS

  return (
    <nav className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        {mainLinks.map((link, idx) =>
          !hasOrgDashboard && idx === 0 ? null : (
            <MainNavLink key={link.text} basePath={basePath} {...link} />
          )
        )}
      </div>

      <Divider isSidebarOpen={isSidebarOpen} />

      {hasOrgSettings ? (
        <div className="flex flex-col gap-1">
          <NavLabel isSidebarOpen={isSidebarOpen}>Settings</NavLabel>

          {SETTINGS_LINKS.map((link) => (
            <MainNavLink key={link.text} basePath={basePath} {...link} />
          ))}

          {hasSlack ? (
            <MainNavLink basePath={basePath} {...SLACK_LINK} />
          ) : null}
        </div>
      ) : null}

      <Divider isSidebarOpen={isSidebarOpen} />

      <div className="flex flex-col gap-1">
        <NavLabel isSidebarOpen={isSidebarOpen}>Resources</NavLabel>

        {SUPPORT_LINKS.map((link) => (
          <MainNavLink key={link.text} basePath={basePath} {...link} />
        ))}
      </div>
    </nav>
  )
}
