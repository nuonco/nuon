import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/utils/classnames'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import {
  DisclosureGroupContext,
  useDisclosureGroup,
  useDisclosureGroupState,
} from './use-disclosure'

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

export interface IExpandAllButton
  extends Omit<IButtonAsButton, 'children' | 'variant'> {
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
  const label = expanded ? collapseLabel : expandLabel

  return (
    <Button
      variant="icon"
      size="sm"
      aria-pressed={expanded}
      aria-label={label}
      tooltipProps={{ tipContent: label }}
      onClick={() => (expanded ? group.closeAll() : group.openAll())}
      {...props}
    >
      <Icon
        variant={
          expanded ? 'ArrowsInLineVerticalIcon' : 'ArrowsOutLineVerticalIcon'
        }
        size={14}
        aria-hidden
      />
    </Button>
  )
}
