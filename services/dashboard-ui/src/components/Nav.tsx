'use client'

import classNames from 'classnames'
import NextLink from 'next/link'
import { usePathname } from 'next/navigation'
import React, { type FC } from 'react'
import { CaretRight, SquaresFour, Wrench, Cube } from '@phosphor-icons/react'
import { Link } from './Link'

export type TLink = {
  href: string
  text?: React.ReactNode
}

// Old UI Nav
export const Nav: FC<{ links?: Array<TLink> }> = ({ links = [] }) => {
  let path = '/dashboard'

  return (
    <nav className="flex gap-2 text-base items-center overflow-y-auto">
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

export const MainNav: FC<{ orgId: string, isSidebarOpen: boolean }> = ({ orgId, isSidebarOpen }) => {
  const path = usePathname()
  const links: Array<TLink> = [
    /* {
     *   href: `/${orgId}`,
     *   text: (
     *     <>
     *       <SquaresFour />
     *       {isSidebarOpen ? "Dashboard" : null}
     *     </>
     *   ),
     * }, */
    {
      href: `/${orgId}/apps`,
      text: (
        <>
          <SquaresFour weight="bold" />
          {isSidebarOpen ? "Apps" : null}
        </>
      ),
    },
    {
      href: `/${orgId}/installs`,
      text: (
        <>
          <Cube weight="bold" />
          {isSidebarOpen ? "Installs" : null}
        </>
      ),
    },
  ]

  return (
    <nav className="flex-auto flex flex-col gap-2">
      {links.map((link) => {
        const isActive = path.split('/')[2] === link.href.split('/')[2]
        return (
          <NextLink
            key={link.href}
            className={classNames(
              'flex items-center font-sans font-medium gap-4 text-lg leading-normal rounded-md p-2.5 w-full',
              {
                'text-cool-grey-600 dark:text-cool-grey-400 hover:bg-black/5 dark:hover:bg-white/10':
                  !isActive,
                'text-primary-600 dark:text-primary-400 bg-primary-100 dark:bg-primary-600/25':
                isActive,
                'justify-center': !isSidebarOpen,
                'justify-start': isSidebarOpen,
              }
            )}
            href={link.href}
          >
            {link.text}
          </NextLink>
        )
      })}
    </nav>
  )
}

export const SubNav: FC<{ links: Array<TLink> }> = ({ links }) => {
  const path = usePathname()
  return (
    <nav className="flex items-center gap-6">
      {links.map((link) => {
        const isActive = path.split('/')?.at(-1) === link.href.split('/')?.at(-1)

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
