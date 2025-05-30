'use client'

import classNames from 'classnames'
import { usePathname } from 'next/navigation'
import React, { type FC } from 'react'
import { CaretRight } from '@phosphor-icons/react'
import { Link, Text } from '@/stratus/components/common'
import type { TNavLink } from '@/types'

export interface IBreadcrumbNav {
  baseCrumbs: Array<TNavLink>
}

export const BreadcrumbNav: FC<IBreadcrumbNav> = ({ baseCrumbs }) => {
  const pathname = usePathname()
  const basePath = baseCrumbs.map((c) => c.path).pop()
  const segments = pathname.split('/').filter(Boolean).slice(4)

  return (
    <nav aria-label="Breadcrumb">
      <ol className="flex gap-2">
        {baseCrumbs.map((crumb, i) => (
          <li key={crumb.path} className="flex items-center gap-2">
            {i > 0 ? (
              <CaretRight className="text-cool-grey-600 dark:text-white/70" />
            ) : null}
            <Text weight="strong">
              <Link
                href={crumb?.path}
                isActive={
                  baseCrumbs?.length === i + 1 && segments?.length === 0
                }
                variant="breadcrumb"
              >
                {crumb?.text}
              </Link>
            </Text>
          </li>
        ))}

        {segments.map((segment, index) => {
          const href = basePath + '/' + segments.slice(0, index + 1).join('/')

          return index === 1 ? null : (
            <li key={href} className="flex items-center gap-2">
              <CaretRight className="text-cool-grey-600 dark:text-white/70" />
              <Text weight="strong">
                <Link
                  href={href}
                  isActive={segments?.length - 1 === index}
                  variant="breadcrumb"
                >
                  {decodeURIComponent(segment)}
                </Link>
              </Text>
            </li>
          )
        })}
      </ol>
    </nav>
  )
}
