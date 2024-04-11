import React, { type FC } from 'react'
import { Link } from './Link'

export type TLink = {
  href: string
  text?: string
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
            <span className="text-gray-500"> / </span>
            <Link href={path}>{l?.text ? l?.text : l.href}</Link>
          </span>
        )
      })}
    </nav>
  )
}
