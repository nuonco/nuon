'use client'

import { usePathname } from 'next/navigation'
import React from 'react'
import { useDashboard, useOrg } from '@/stratus/context'
import { Icon, Link, Text, Tooltip } from '@/stratus/components/common'
import type { TNavLink } from '@/types'
import './MainNav.css'

const MAIN_LINKS: Array<TNavLink> = [
  {
    icon: <Icon variant="House" weight="bold" />,
    path: `/`,
    text: 'Dashboard',
  },
  {
    icon: <Icon variant="AppWindow" weight="bold" />,
    path: `/apps`,
    text: 'Apps',
  },
  {
    icon: <Icon variant="Cube" weight="bold" />,
    path: `/installs`,
    text: 'Installs',
  },
  {
    icon: <Icon variant="SneakerMove" weight="bold" />,
    path: `/runner`,
    text: 'Build runner',
  },
]

const SETTINGS_LINKS: Array<TNavLink> = [
  {
    icon: <Icon variant="UsersThree" weight="bold" />,
    path: `/team`,
    text: 'Team',
  },
]

const SUPPORT_LINKS: Array<TNavLink> = [
  {
    icon: <Icon variant="BookOpen" weight="bold" />,
    path: `https://docs.nuon.co/get-started/introduction`,
    text: 'Devloper docs',
    isExternal: true,
  },
  {
    icon: <Icon variant="ListBullets" weight="bold" />,
    path: `/releases`,
    text: 'Releases',
  },
]

export const MainNav = () => {
  const { org } = useOrg()
  const basePath = `/stratus/${org.id}`
  return (
    <nav className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        {MAIN_LINKS.map((link) => (
          <MainNavLink key={link.text} basePath={basePath} {...link} />
        ))}
      </div>

      <hr />

      {org?.features?.['org-settings'] ? (
        <div className="flex flex-col gap-1">
          <Text variant="subtext" className="nav-label px-2">
            Settings
          </Text>

          {SETTINGS_LINKS.map((link) => (
            <MainNavLink key={link.text} basePath={basePath} {...link} />
          ))}
        </div>
      ) : null}

      <hr />

      {org?.features?.['org-support'] ? (
        <div className="flex flex-col gap-1">
          <Text variant="subtext" className="nav-label px-2">
            Resources
          </Text>

          {SUPPORT_LINKS.map((link) => (
            <MainNavLink key={link.text} basePath={basePath} {...link} />
          ))}
        </div>
      ) : null}
    </nav>
  )
}

interface IMainNavLink extends TNavLink {
  basePath: string
}

const MainNavLink = ({
  basePath,
  text,
  icon,
  path,
  isExternal,
}: IMainNavLink) => {
  const { isSidebarOpen } = useDashboard()
  const pathName = usePathname()
  const normalizePath = (path: string) =>
    path.endsWith('/') ? path.slice(0, -1) : path
  const normalizedPathName = normalizePath(pathName)
  const fullPath = normalizePath(`${basePath}${path}`)
  const isActive =
    fullPath === normalizedPathName ||
    (path !== `/` && normalizedPathName.startsWith(`${fullPath}/`))

  const link = (
    <Link
      aria-current={isActive ? 'page' : undefined}
      href={isExternal ? path : `${basePath}${path}`}
      isExternal={isExternal}
      variant="nav"
      isActive={isActive}
    >
      <span>{icon}</span>
      <span className="link-text">{text}</span>
    </Link>
  )

  return isSidebarOpen ? (
    link
  ) : (
    <Tooltip
      position="right"
      tipContent={
        <Text variant="subtext" weight="stronger">
          {text
            .trim()
            .split(' ')
            .at(-1)
            ?.replace(/^./, (char) => char.toUpperCase())}
        </Text>
      }
    >
      {link}
    </Tooltip>
  )
}
