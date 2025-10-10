'use client'

import { usePathname } from 'next/navigation'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TNavLink } from '@/types'

export interface IBreadcrumbNav {
  baseCrumbs: TNavLink[]
}

export const BreadcrumbNav = ({ baseCrumbs }: IBreadcrumbNav) => {
  const pathname = usePathname()

  // Get the last base crumb's path as the "base" for dynamic segments
  const basePath =
    baseCrumbs.length > 0 ? baseCrumbs[baseCrumbs.length - 1].path : ''
  // Remove empty segments and the first 4 (as in original)
  const segments = pathname.split('/').filter(Boolean).slice(3)

  // Helper to render a separator
  const Separator = () => <Icon variant="CaretRight" className="muted" />

  return (
    <nav aria-label="Breadcrumb">
      <ol className="flex gap-2">
        {/* Render static/base crumbs */}
        {baseCrumbs.map((crumb, i) => (
          <li key={crumb.path} className="flex items-center gap-2">
            {i > 0 && <Separator />}
            <Text weight="strong">
              <Link
                href={crumb.path}
                isActive={i === baseCrumbs.length - 1 && segments.length === 0}
                variant="breadcrumb"
              >
                {crumb.text}
              </Link>
            </Text>
          </li>
        ))}

        {/* Render dynamic crumbs from pathname */}
        {segments.map((segment, idx) => {
          // Compose href for this crumb
          const href =
            basePath.replace(/\/$/, '') +
            '/' +
            segments.slice(0, idx + 1).join('/')

          // Skip index === 1 as in original (if this is not intentional, remove this)
          if (idx === 1) return null

          return (
            <li key={href} className="flex items-center gap-2">
              <Separator />
              <Text weight="strong">
                <Link
                  href={href}
                  isActive={idx === segments.length - 1}
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
