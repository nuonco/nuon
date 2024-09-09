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
        'flex flex-initial items-center w-fit gap-1 text-primary-700 hover:text-primary-600 focus:text-primary-600 active:text-primary-800 dark:text-primary-500 dark:hover:text-primary-400 dark:focus:text-primary-400 dark:active:text-primary-600',
        { [`${className}`]: Boolean(className) }
      )}
      href={href}
      {...props}
    >
      {children}
    </NextLink>
  )
}
