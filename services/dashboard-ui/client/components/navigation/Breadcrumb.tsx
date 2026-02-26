'use client'

import { useEffect } from 'react'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useBreadcrumb } from '@/hooks/use-breadcrumb'
import type { TNavLink } from '@/types'

export const BreadcrumbNav = () => {
  const { breadcrumbLinks, isLoading } = useBreadcrumb()
  const Separator = () => <Icon variant="CaretRight" className="muted" />

  return (
    <nav aria-label="Breadcrumb" className="max-w-[1000px] truncate">
      <ol className="flex gap-2">
        {breadcrumbLinks.map((crumb, idx) => (
          <li key={`${crumb.path}-${idx}`} className="flex items-center gap-2">
            {idx > 0 && <Separator />}
            <Text weight="strong">
              {isLoading ? (
                <Skeleton
                  height="17px"
                  width={`${crumb?.text?.length * 16 * 0.6}px`}
                  maxWidth="200px"
                />
              ) : (
                <Link
                  href={crumb.path}
                  isActive={idx === breadcrumbLinks?.length - 1}
                  variant="breadcrumb"
                >
                  {crumb.text}
                </Link>
              )}
            </Text>
          </li>
        ))}
      </ol>
    </nav>
  )
}

export const Breadcrumbs = ({ breadcrumbs }: { breadcrumbs: TNavLink[] }) => {
  const { updateBreadcrumb } = useBreadcrumb()

  useEffect(() => {
    updateBreadcrumb(breadcrumbs)
  }, [])

  return <></>
}
