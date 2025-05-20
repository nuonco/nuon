'use client'

import classNames from 'classnames'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import React, { useMemo, type FC } from 'react'
import {
  AppWindow,
  BookOpen,
  House,
  Cube,
  SneakerMove,
  UsersThree,
} from '@phosphor-icons/react'
import { useDashboard, useOrg } from '@/stratus/context'
import { Text } from '@/stratus/components/common'
import './DashboardSidebarNav.css'

export type TNavLink = {
  icon: React.ReactNode
  path: string
  text: string
  isExternal?: boolean
}

const DashboardLinks: Array<TNavLink> = [
  {
    icon: <House weight="bold" />,
    path: `/`,
    text: 'Dashboard',
  },
  {
    icon: <AppWindow weight="bold" />,
    path: `/apps`,
    text: 'Apps',
  },
  {
    icon: <Cube weight="bold" />,
    path: `/installs`,
    text: 'Installs',
  },
  {
    icon: <SneakerMove weight="bold" />,
    path: `/runner`,
    text: 'Build runner',
  },
]

const SettingsLinks: Array<TNavLink> = [
  {
    icon: <UsersThree weight="bold" />,
    path: `/team`,
    text: 'Team',
  },
]

const SupportLinks: Array<TNavLink> = [
  {
    icon: <BookOpen weight="bold" />,
    path: `https://docs.nuon.co/get-started/introduction`,
    text: 'Devloper docs',
    isExternal: true,
  },
]

export const DashboardSidebarNav: FC = () => {
  const { org } = useOrg()
  const { isSidebarOpen } = useDashboard()
  const basePath = `/stratus/${org.id}`
  return (
    <nav className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        {DashboardLinks.map((link) => (
          <SidebarNavLink key={link.text} basePath={basePath} {...link} />
        ))}
      </div>

      <hr />

      {org?.features?.['org-settings'] ? (
        <div className="flex flex-col gap-1">
          <Text variant="subtext" className="nav-label px-2">
            Settings
          </Text>

          {SettingsLinks.map((link) => (
            <SidebarNavLink key={link.text} basePath={basePath} {...link} />
          ))}
        </div>
      ) : null}

      <hr />

      {org?.features?.['org-support'] ? (
        <div className="flex flex-col gap-1">
          <Text variant="subtext" className="nav-label px-2">
            Resources
          </Text>

          {SupportLinks.map((link) => (
            <SidebarNavLink key={link.text} basePath={basePath} {...link} />
          ))}
        </div>
      ) : null}
    </nav>
  )
}

interface ISidebarNavLink extends TNavLink {
  basePath: string
}

const SidebarNavLink: FC<ISidebarNavLink> = ({
  basePath,
  text,
  icon,
  path,
  isExternal,
}) => {
  const pathName = usePathname()
  const normalizePath = (path: string) =>
    path.endsWith('/') ? path.slice(0, -1) : path
  const normalizedPathName = normalizePath(pathName)
  const fullPath = normalizePath(`${basePath}${path}`)
  const isActive =
    fullPath === normalizedPathName ||
    (path !== `/` && normalizedPathName.startsWith(`${fullPath}/`))

  return isExternal ? (
    <a
      key={text}
      href={path}
      target="_blank"
      className={classNames('link', {
        'link-active': isActive,
        'link-inactive': !isActive,
      })}
      aria-current={isActive ? 'page' : undefined}
    >
      <span>{icon}</span>
      <span className="link-text">{text}</span>
    </a>
  ) : (
    <Link
      key={text}
      href={`${basePath}${path}`}
      className={classNames('link', {
        'link-active': isActive,
        'link-inactive': !isActive,
      })}
      aria-current={isActive ? 'page' : undefined}
    >
      <span>{icon}</span>
      <span className="link-text">{text}</span>
    </Link>
  )
}
