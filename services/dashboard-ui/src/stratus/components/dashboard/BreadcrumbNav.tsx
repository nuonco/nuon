'use client'

import { usePathname } from 'next/navigation'
import React from 'react'
import { Icon, Link, Text } from '@/stratus/components/common'
import type { TNavLink } from '@/types'

export interface IBreadcrumbNav {
  baseCrumbs: Array<TNavLink>
}

export const BreadcrumbNav = ({ baseCrumbs }: IBreadcrumbNav) => {
  const pathname = usePathname()
  const basePath = baseCrumbs.map((c) => c.path).pop()
  const segments = pathname.split('/').filter(Boolean).slice(4)

  return (
    <nav aria-label="Breadcrumb">
      <ol className="flex gap-2">
        {baseCrumbs.map((crumb, i) => (
          <li key={crumb.path} className="flex items-center gap-2">
            {i > 0 ? <Icon variant="CaretRight" className="muted" /> : null}
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
              <Icon variant="CaretRight" className="muted" />
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
