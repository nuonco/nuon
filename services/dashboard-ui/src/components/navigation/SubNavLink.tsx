'use client'

import { usePathname } from 'next/navigation'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { usePageSidebar } from '@/hooks/use-page-sidebar'
import type { TNavLink } from '@/types'
import { cn } from '@/utils/classnames'

export const SubNavLink = ({
  basePath,
  iconVariant,
  path,
  text,
}: TNavLink & { basePath: string }) => {
  const { isPageSidebarOpen } = usePageSidebar()
  const pathName = usePathname()
  const normalizePath = (path: string) =>
    path.endsWith('/') ? path.slice(0, -1) : path
  const normalizedPathName = normalizePath(pathName)
  const fullPath = normalizePath(`${basePath}${path}`)
  const isActive =
    fullPath === normalizedPathName ||
    (path !== `/` && normalizedPathName.startsWith(`${fullPath}/`))

  const link = (
    <Link
      aria-current={isActive ? 'page' : undefined}
      href={`${basePath}/${path}`}
      isActive={isActive}
      variant="nav"
    >
      <span>
        {iconVariant ? <Icon variant={iconVariant} weight="bold" /> : null}
      </span>
      <span
        className={cn(
          'whitespace-nowrap transition-all duration-fastest ease-cubic md:ml-2 w-fit',
          {
            'md:opacity-100 md:w-full': isPageSidebarOpen,
            'md:opacity-0 w-0': !isPageSidebarOpen,
          }
        )}
      >
        {text}
      </span>
    </Link>
  )

  return (
    <Tooltip
      className="w-full"
      position="right"
      tipContent={
        <Text variant="subtext" weight="stronger">
          {text
            .trim()
            .split(' ')
            .at(-1)
            ?.replace(/^./, (char) => char.toUpperCase())}
        </Text>
      }
      tipContentClassName={cn('hidden', {
        'md:flex': !isPageSidebarOpen,
      })}
    >
      {link}
    </Tooltip>
  )
}
