import type { HTMLAttributes } from 'react'
import {
  Button,
  type IButtonAsButton,
  type TButtonVariant,
  type TButtonSize,
} from '@/components/common/Button'
import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { cn } from '@/utils/classnames'

export interface ISplitButton
  extends HTMLAttributes<Omit<HTMLSpanElement, 'children'>> {
  buttonProps?: Omit<IButtonAsButton, 'variant' | 'size'>
  dropdownProps?: Omit<IDropdown, 'buttonText' | 'icon' | 'variant' | 'size'>
  size?: TButtonSize
  variant?: Exclude<TButtonVariant, 'ghost' | 'tab'>
}

export const SplitButton = ({
  buttonProps: { className: buttonClassName, ...buttonProps } = {},
  className,
  dropdownProps: { buttonClassName: dropdownBtnClassName, ...dropdownProps } = {
    id: 'split-btn',
    children: <>No dropdown children</>,
  },
  size = 'md',
  variant,
}: ISplitButton) => {
  return (
    <span className={cn('flex items-center', className)}>
      <Button
        className={cn('!rounded-e-none focus:z-10', buttonClassName)}
        size={size}
        variant={variant}
        {...buttonProps}
      />
      <Dropdown
        buttonText=""
        buttonClassName={cn(
          '!px-1.5 !rounded-s-none !border-l-0 focus:z-10',
          dropdownBtnClassName
        )}
        icon={<Icon variant="DotsThreeVerticalIcon" size={getIconSize(size)} />}
        size={size}
        variant={variant}
        {...dropdownProps}
      />
    </span>
  )
}

function getIconSize(size: TButtonSize) {
  return size === 'lg' ? '26' : size === 'sm' ? '18' : '22'
}
