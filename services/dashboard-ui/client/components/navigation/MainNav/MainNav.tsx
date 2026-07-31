import { Text } from '@/components/common/Text'
import { cn } from '@/utils/classnames'
import { MainNavLink } from '../MainNavLink'
import {
  MAIN_LINKS,
  SETTINGS_LINKS,
  SLACK_LINK,
  SUPPORT_LINKS,
} from '../main-nav-links'
import type { TOrg } from '@/types'

interface IMainNav {
  org: TOrg
  isSidebarOpen: boolean
  hasServiceAccountsAndTokens: boolean
  hasSlack: boolean
  hasTriggers: boolean
  hasCustomerPortal: boolean
  customerPortalUrl: string
}

const NavLabel = ({
  children,
  isSidebarOpen,
}: {
  children: string
  isSidebarOpen: boolean
}) => (
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
  hasServiceAccountsAndTokens,
  hasSlack,
  hasTriggers,
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
  const settingsLinks = SETTINGS_LINKS.filter(
    (link) =>
      (hasTriggers || link.path !== '/triggers') &&
      (hasServiceAccountsAndTokens ||
        (link.path !== '/api-tokens' && link.path !== '/service-accounts'))
  )

  return (
    <nav className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        {mainLinks.map((link) => (
          <MainNavLink key={link.text} basePath={basePath} {...link} />
        ))}
      </div>

      <Divider isSidebarOpen={isSidebarOpen} />

      <div className="flex flex-col gap-1">
        <NavLabel isSidebarOpen={isSidebarOpen}>Settings</NavLabel>

        {settingsLinks.map((link) => (
          <MainNavLink key={link.text} basePath={basePath} {...link} />
        ))}

        {hasSlack ? (
          <MainNavLink basePath={basePath} {...SLACK_LINK} />
        ) : null}
      </div>

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
