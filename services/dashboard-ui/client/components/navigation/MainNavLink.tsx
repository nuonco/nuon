import { useLocation } from 'react-router'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { useSidebar } from '@/hooks/use-sidebar'
import type { TNavLink } from '@/types'
import { cn } from '@/utils/classnames'

interface IMainNavLink extends TNavLink {
  basePath: string
}

export const MainNavLink = ({
  basePath,
  text,
  iconVariant,
  path,
  isExternal,
}: IMainNavLink) => {
  const { isSidebarOpen } = useSidebar()
  const { pathname: pathName } = useLocation()
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
      className={cn({})}
      href={isExternal ? path : `${basePath}${path}`}
      target={isExternal ? '_blank' : undefined}
      rel={isExternal ? 'noopener noreferrer' : undefined}
      isActive={isActive}
      tabIndex={0}
      variant="nav"
      data-active={isActive ? 'true' : undefined}
    >
      <span>
        {iconVariant ? <Icon variant={iconVariant} weight="bold" /> : null}
      </span>
      <span
        className={cn('transition-all duration-fast whitespace-nowrap flex items-center gap-1', {
          'md:opacity-100 w-full': isSidebarOpen,
          'md:opacity-0 md:w-0': !isSidebarOpen,
        })}
      >
        {text}
        {isExternal ? <Icon variant="ArrowSquareOut" size={12} /> : null}
      </span>
    </Link>
  )

  return (
    <Tooltip
      className="w-full"
      position="right"
      tipContentClassName={cn('hidden', {
        'md:flex': !isSidebarOpen,
      })}
      tipContent={
        <Text variant="subtext" weight="stronger">
          {text
            .trim()
            .split(' ')
            .at(-1)
            ?.replace(/^./, (char) => char.toUpperCase())}
        </Text>
      }
    >
      {link}
    </Tooltip>
  )
}
