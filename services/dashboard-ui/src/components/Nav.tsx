'use client'

import classNames from 'classnames'
import NextLink from 'next/link'
import { usePathname } from 'next/navigation'
import React, { type FC, useContext } from 'react'
import {
  AppWindow,
  ArrowSquareOut,
  Books,
  CaretRight,
  Cube,
  GraduationCap,
  Signpost,
  ListDashes,
  SneakerMove,
  SquaresFour,
  UsersThree,
} from '@phosphor-icons/react'
import { useOrg } from '@/hooks/use-org'
import { UserJourneyContext } from '@/providers/user-journey-provider'
import { Link } from './Link'
import { Text } from './Typography'

export type TLink = {
  href: string
  text?: React.ReactNode
  isExternal?: boolean
  onClick?: () => void
}

export const MainNav: FC<{
  isSidebarOpen: boolean
}> = ({ isSidebarOpen }) => {
  const { org } = useOrg()
  const path = usePathname()
  const { openOnboarding } = useContext(UserJourneyContext) || {}
  const links: Array<TLink> = [
    {
      href: `/${org.id}`,
      text: (
        <>
          <span>
            <SquaresFour />
          </span>
          {isSidebarOpen ? (
            <span className="overflow-hidden">Dashboard</span>
          ) : null}
        </>
      ),
    },

    {
      href: `/${org.id}/apps`,
      text: (
        <>
          <span>
            <AppWindow weight="bold" />
          </span>
          {isSidebarOpen ? <span className="overflow-hidden">Apps</span> : null}
        </>
      ),
    },
    {
      href: `/${org.id}/installs`,
      text: (
        <>
          <span>
            <Cube weight="bold" />
          </span>
          {isSidebarOpen ? (
            <span className="overflow-hidden">Installs</span>
          ) : null}
        </>
      ),
    },
    {
      href: `/${org.id}/runner`,
      text: (
        <>
          <span>
            <SneakerMove weight="bold" />
          </span>
          {isSidebarOpen ? (
            <span className="overflow-hidden">Build runner</span>
          ) : null}
        </>
      ),
    },
  ]

  function getMainNavItems(links: Array<TLink>) {
    const l = links
    if (!org.features?.['org-dashboard']) {
      l.shift()
    }
    if (!org.features?.['org-runner']) {
      l.pop()
    }

    return l
  }

  const settingsLinks: Array<TLink> = [
    {
      href: `/${org.id}/team`,
      text: (
        <>
          <span>
            <UsersThree weight="bold" />
          </span>
          {isSidebarOpen ? <span className="overflow-hidden">Team</span> : null}
        </>
      ),
    },
  ]

  const supportLinks: Array<TLink> = [
    {
      href: `https://docs.nuon.co`,
      text: (
        <>
          <span>
            <Books weight="bold" />
          </span>
          {isSidebarOpen ? (
            <span className="overflow-hidden flex items-center gap-4">
              Developer docs <ArrowSquareOut size="14" />
            </span>
          ) : null}
        </>
      ),
      isExternal: true,
    },
    {
      href: '#',
      text: (
        <>
          <span>
            <Signpost weight="bold" />
          </span>
          {isSidebarOpen ? (
            <span className="overflow-hidden">Review Onboarding</span>
          ) : null}
        </>
      ),
      onClick: openOnboarding,
    },
  ]

  const NavLink: FC<{ link: TLink }> = ({ link }) => {
    const pathParts = path.split('/')
    const hrefParts = link.href.split('/')
    const isActive = pathParts[2] === hrefParts[2]
    const classes = classNames(
      'flex items-center font-sans font-medium gap-4 text-sm leading-normal rounded-md p-2 w-full text-nowrap overflow-hidden',
      {
        '!text-cool-grey-800 dark:!text-cool-grey-400 hover:bg-black/5 dark:hover:bg-white/10':
          !isActive,
        '!text-primary-800 dark:!text-primary-400 bg-primary-100 dark:bg-primary-600/25':
          isActive,
        'justify-center': !isSidebarOpen,
        'justify-start': isSidebarOpen,
      }
    )

    const handleClick = (e: React.MouseEvent) => {
      if (link.onClick) {
        e.preventDefault()
        link.onClick()
      }
    }

    return link.isExternal ? (
      <Link className={classes} href={link.href} target="_blank">
        {link.text}
      </Link>
    ) : link.onClick ? (
      <button className={classes} onClick={handleClick}>
        {link.text}
      </button>
    ) : (
      <NextLink key={link.href} className={classes} href={link.href}>
        {link.text}
      </NextLink>
    )
  }

  return (
    <nav className="flex-auto flex flex-col gap-2 w-full">
      {getMainNavItems(links).map((link) => (
        <NavLink key={link.href} link={link} />
      ))}

      {org?.features?.['org-settings'] ? (
        <div
          className={classNames('flex flex-col gap-2 pt-2 mt-4', {
            'border-t': !isSidebarOpen,
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

      {org?.features?.['org-support'] ? (
        <div
          className={classNames('flex flex-col gap-2 pt-2 mt-4', {
            'border-t': !isSidebarOpen,
          })}
        >
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
                'text-primary-600 dark:text-primary-400 !border-current':
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
    <div className="flex items-center gap-2 ml-8 md:ml-0 overflow-x-auto max-w-[230px] md:max-w-full">
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
