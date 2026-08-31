import { useLocation, useNavigate } from 'react-router'
import { cn } from '@/utils/classnames'
import { Block, type IBlock } from './Block'
import type { INavItem } from './types'

export interface INavBlock extends IBlock, INavItem {
  exact?: boolean
}

export const NavBlock = ({
  path,
  label,
  className,
  exact = false,
  ...props
}: INavBlock) => {
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const isActive =
    exact || path === '/' ? pathname === path : pathname.startsWith(path)

  return (
    <Block
      title={label}
      text={label}
      onClick={() => navigate(path)}
      className={cn(
        className,
        'cursor-pointer transition-opacity hover:opacity-80',
        isActive ? 'opacity-100' : 'opacity-40'
      )}
      {...props}
    />
  )
}
