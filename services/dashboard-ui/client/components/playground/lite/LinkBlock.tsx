import { useNavigate } from 'react-router'
import { cn } from '@/utils/classnames'
import { Block, type IBlock } from './Block'
import type { INavItem } from './types'
import { labelWidth } from './utils'

export interface ILinkBlock extends IBlock, INavItem {}

export const LinkBlock = ({ path, label, className, ...props }: ILinkBlock) => {
  const navigate = useNavigate()

  return (
    <Block
      title={label}
      text={label}
      onClick={() => navigate(path)}
      className={cn(
        'cursor-pointer transition-opacity hover:opacity-60',
        className
      )}
      style={{ width: labelWidth(label) }}
      {...props}
    />
  )
}
