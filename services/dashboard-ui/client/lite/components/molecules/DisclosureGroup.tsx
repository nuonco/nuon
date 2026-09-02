import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import {
  DisclosureGroupContext,
  useDisclosureGroup,
  useDisclosureGroupState,
} from '../../hooks/use-disclosure'
import { Button, type IButton } from '../atoms/Button'
import { Icon } from '../atoms/Icon'

export interface IDisclosureGroup extends HTMLAttributes<HTMLDivElement> {
  defaultOpen?: boolean
  children: ReactNode
}

export const DisclosureGroup = ({
  defaultOpen = false,
  className,
  children,
  ...props
}: IDisclosureGroup) => {
  const group = useDisclosureGroupState(defaultOpen)

  return (
    <DisclosureGroupContext.Provider value={group}>
      <div className={cn('flex flex-col', className)} {...props}>
        {children}
      </div>
    </DisclosureGroupContext.Provider>
  )
}

export interface IExpandAllButton extends Omit<IButton, 'children'> {
  expandLabel?: string
  collapseLabel?: string
}

export const ExpandAllButton = ({
  expandLabel = 'Expand all',
  collapseLabel = 'Collapse all',
  ...props
}: IExpandAllButton) => {
  const group = useDisclosureGroup()
  if (!group) return null

  const expanded = group.allOpen

  return (
    <Button
      variant="ghost"
      size="sm"
      aria-pressed={expanded}
      onClick={() => (expanded ? group.closeAll() : group.openAll())}
      icon={
        <Icon
          variant={expanded ? 'CaretUpIcon' : 'CaretDownIcon'}
          size={14}
          aria-hidden
        />
      }
      {...props}
    >
      {expanded ? collapseLabel : expandLabel}
    </Button>
  )
}
