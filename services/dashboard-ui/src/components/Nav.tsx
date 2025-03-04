'use client'

import classNames from 'classnames'
import NextLink from 'next/link'
import { usePathname } from 'next/navigation'
import React, { type FC } from 'react'
import {
  AppWindow,
  Books,
  CaretRight,
  Cube,
  ListDashes,
  SneakerMove,
  SquaresFour,
  UsersThree,
} from '@phosphor-icons/react'
import { Link } from './Link'
import { Text } from './Typography'

export type TLink = {
  href: string
  text?: React.ReactNode
  isExternal?: boolean
}

export const MainNav: FC<{
  orgId: string
  isSidebarOpen: boolean
  featureFlags?: Record<string, boolean>
}> = ({ orgId, isSidebarOpen, featureFlags }) => {
  const path = usePathname()
  const links: Array<TLink> = [
    {
      href: `/${orgId}`,
      text: (
        <>
          <SquaresFour />
          {isSidebarOpen ? 'Dashboard' : null}
        </>
      ),
    },

    {
      href: `/${orgId}/apps`,
      text: (
        <>
          <AppWindow weight="bold" />
          {isSidebarOpen ? 'Apps' : null}
        </>
      ),
    },
    {
      href: `/${orgId}/installs`,
      text: (
        <>
          <Cube weight="bold" />
          {isSidebarOpen ? 'Installs' : null}
        </>
      ),
    },
    {
      href: `/${orgId}/runner`,
      text: (
        <>
          <SneakerMove weight="bold" />
          {isSidebarOpen ? 'Runner' : null}
        </>
      ),
    },
  ]

  function getMainNavItems(links: Array<TLink>) {
    const l = links
    if (!featureFlags['ORG_DASHBOARD']) {
      l.shift()
    }
    if (!featureFlags['ORG_RUNNER']) {
      l.pop()
    }

    return l
  }

  const settingsLinks: Array<TLink> = [
    {
      href: `/${orgId}/team`,
      text: (
        <>
          <UsersThree weight="bold" />
          {isSidebarOpen ? 'Team' : null}
        </>
      ),
    },
  ]

  const supportLinks: Array<TLink> = [
    {
      href: `https://docs.nuon.co`,
      text: (
        <>
          <Books weight="bold" />
          {isSidebarOpen ? 'Developer docs' : null}
        </>
      ),
      isExternal: true,
    },

    {
      href: `/releases`,
      text: (
        <>
          <ListDashes weight="bold" />
          {isSidebarOpen ? 'Releases' : null}
        </>
      ),
    },
  ]

  const NavLink: FC<{ link: TLink }> = ({ link }) => {
    const isActive = path.split('/')[2] === link.href.split('/')[2]
    const classes = classNames(
      'flex items-center font-sans font-medium gap-4 text-lg leading-normal rounded-md p-2.5 w-full',
      {
        '!text-cool-grey-800 dark:!text-cool-grey-400 hover:bg-black/5 dark:hover:bg-white/10':
          !isActive,
        '!text-primary-800 dark:!text-primary-400 bg-primary-100 dark:bg-primary-600/25':
          isActive,
        'justify-center': !isSidebarOpen,
        'justify-start': isSidebarOpen,
      }
    )

    return link.isExternal ? (
      <Link className={classes} href={link.href} target="_blank">
        {link.text}
      </Link>
    ) : (
      <NextLink key={link.href} className={classes} href={link.href}>
        {link.text}
      </NextLink>
    )
  }

  return (
    <nav className="flex-auto flex flex-col gap-2">
      {getMainNavItems(links).map((link) => (
        <NavLink key={link.href} link={link} />
      ))}

      {featureFlags['ORG_SETTINGS'] ? (
        <div
          className={classNames('flex flex-col gap-2 py-2 my-4', {
            'border-y': !isSidebarOpen,
          })}
        >
          <Text
            className={classNames('text-cool-grey-600 dark:text-white/70', {
              hidden: !isSidebarOpen,
            })}
            variant="med-14"
          >
            Settings
          </Text>

          {settingsLinks.map((link) => (
            <NavLink key={link.href} link={link} />
          ))}
        </div>
      ) : null}

      {featureFlags['ORG_SUPPORT'] ? (
        <div className={classNames('flex flex-col gap-2', {})}>
          <Text
            className={classNames('text-cool-grey-600 dark:text-white/70', {
              hidden: !isSidebarOpen,
            })}
            variant="med-14"
          >
            Support
          </Text>

          {supportLinks.map((link) => (
            <NavLink key={link.href} link={link} />
          ))}
        </div>
      ) : null}
    </nav>
  )
}

export const SubNav: FC<{ links: Array<TLink> }> = ({ links }) => {
  const path = usePathname()
  return (
    <nav className="flex items-center gap-6">
      {links.map((link) => {
        const isActive =
          path.split('/')?.at(-1) === link.href.split('/')?.at(-1)

        return (
          <NextLink
            className={classNames(
              'px-4 py-3 border-b text-base font-sans font-medium leading-normal',
              {
                'text-cool-grey-600 dark:text-cool-grey-400 border-transparent':
                  !isActive,
                'text-primary-600 dark:text-primary-400 border-current':
                  isActive,
              }
            )}
            key={link.href}
            href={link.href}
          >
            {link.text}
          </NextLink>
        )
      })}
    </nav>
  )
}

export const BreadcrumbNav: FC<{ links: Array<TLink> }> = ({ links }) => {
  return (
    <div className="flex items-center gap-2">
      {links.map((link, i) => (
        <span
          key={`${link.href}-${i}`}
          className="flex items-center gap-2 font-sans font-semibold leading-normal tracking-wide text-base"
        >
          {i !== 0 ? (
            <CaretRight className="text-cool-grey-600 dark:text-cool-grey-500" />
          ) : null}
          <Link
            className="!inline max-w-60 truncate"
            href={link.href}
            variant="breadcrumb"
            isActive={links.length === i + 1}
          >
            {link.text}
          </Link>
        </span>
      ))}
    </div>
  )
}
