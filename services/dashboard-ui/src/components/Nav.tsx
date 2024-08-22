'use client'

import classNames from 'classnames'
import NextLink from 'next/link'
import { usePathname } from 'next/navigation'
import React, { type FC } from 'react'
import { GrAppsRounded } from 'react-icons/gr'
import { IoIosCloudOutline } from 'react-icons/io'
import { Link } from './Link'

export type TLink = {
  href: string
  text?: React.ReactNode
}

export const Nav: FC<{ links?: Array<TLink> }> = ({ links = [] }) => {
  let path = '/dashboard'

  return (
    <nav className="flex gap-2 text-xs items-center overflow-y-auto">
      <Link key={path} href={path}>
        Dashboard
      </Link>
      {links.map((l) => {
        path = `${path}/${l.href}`
        return (
          <span className="flex items-center gap-2" key={l.href}>
            <span className="text-slate-500"> / </span>
            <Link href={path}>{l?.text ? l?.text : l.href}</Link>
          </span>
        )
      })}
    </nav>
  )
}

export const MainNav: FC<{ orgId: string }> = ({ orgId }) => {
  const path = usePathname()
  const links: Array<TLink> = [
    {
      href: `/beta/${orgId}/apps`,
      text: (
        <>
          <GrAppsRounded className="text-lg" />
          Apps
        </>
      ),
    },
    {
      href: `/beta/${orgId}/installs`,
      text: (
        <>
          <IoIosCloudOutline className="text-lg" />
          Installs
        </>
      ),
    },
  ]

  return (
    <nav className="flex-auto flex flex-col gap-4">
      {links.map((link) => (
        <NextLink
          key={link.href}
          className={classNames(
            'flex items-center justify-start font-medium gap-4 text-base rounded p-2 hover:bg-gray-900/5 dark:hover:bg-gray-100/5 w-full',
            {
              'text-fuchsia-500 dark:text-fuchsia-300 bg-fuchsia-500/10 dark:bg-gray-100/5':
                path.split('/')[3] === link.href.split('/')[3],
            }
          )}
          href={link.href}
        >
          {link.text}
        </NextLink>
      ))}
    </nav>
  )
}

export const SubNav: FC<{ links: Array<TLink> }> = ({ links }) => {
  const path = usePathname()
  return (
    <nav className="flex items-center gap-6">
      {links.map((link) => (
        <NextLink
          className={classNames('px-4 py-3 border-b-2 text-sm font-medium', {
            'text-gray-600 dark:text-gray-400 border-transparent':
              path.split('/')?.[5] !== link.href.split('/')?.[5],
            'text-active dark:text-active border-current':
              path.split('/')?.[5] === link.href.split('/')?.[5],
          })}
          key={link.href}
          href={link.href}
        >
          {link.text}
        </NextLink>
      ))}
    </nav>
  )
}
