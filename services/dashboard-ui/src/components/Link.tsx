import classNames from 'classnames'
import NextLink from 'next/link'
import  { type FC, HTMLAttributes } from 'react'

export interface ILink extends HTMLAttributes<HTMLAnchorElement> {
  href: string
  target?: string
  isExternal?: boolean
}

export const Link: FC<ILink> = ({ className, children, href, ...props }) => {
  return (
    <NextLink
      className={classNames(
        'flex flex-initial items-center w-fit gap-1 text-fuchsia-700 hover:text-fuchsia-600 focus:text-fuchsia-600 active:text-fuchsia-800 dark:text-fuchsia-500 dark:hover:text-fuchsia-400 dark:focus:text-fuchsia-400 dark:active:text-fuchsia-600',
        { [`${className}`]: Boolean(className) }
      )}
      href={href}
      {...props}
    >
      {children}
    </NextLink>
  )
}
