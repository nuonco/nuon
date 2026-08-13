import React from 'react'
import { cn } from '@/utils/classnames'
import { Button, type IButtonAsButton } from './Button'
import { Dropdown, type IDropdown } from './Dropdown'
import { Link, type ILink } from './Link'
import { Text, type IText } from './Text'

export interface IMenu
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'role'> {}

export const Menu = ({ className, children, ...props }: IMenu) => {
  return (
    <div
      className={cn('flex flex-col p-2 gap-0.5 w-56', className)}
      role="menu"
      {...props}
    >
      {React.Children.map(children, (c) => {
        if (!React.isValidElement(c)) return null

        const childProps = (c as any).props ?? {}
        const menuItemClass =
          '!p-2 text-sm !leading-none h-8 w-full flex justify-between !rounded-md'
        const menuLinkItem =
          '!bg-transparent !text-cool-grey-800 dark:!text-white/70 hover:!bg-cool-grey-500/8 dark:hover:!bg-cool-grey-500/8 focus:!bg-cool-grey-500/8 dark:focus:!bg-cool-grey-500/8 active:!bg-cool-grey-500/16 dark:active:!bg-cool-grey-500/16'
        const menuLinkDanger =
          '!bg-transparent !text-red-800 dark:!text-red-500 hover:!bg-red-50 dark:hover:!bg-[#1D0D10] focus:!bg-red-50 dark:focus:!bg-[#1D0D10] active:!bg-red-100 dark:active:!bg-[#2E1013]'

        if (c.type === Button || c.type === Link || childProps.isMenuButton) {
          const isDanger = childProps.variant === 'danger'
          const isLink = c.type === Link
          return React.cloneElement<IButtonAsButton | ILink>(c, {
            variant: isDanger ? 'danger' : 'ghost',
            isMenuButton: isLink ? childProps.isMenuButton : true,
            className: cn(
              menuItemClass,
              isLink && (isDanger ? menuLinkDanger : menuLinkItem),
              childProps.className
            ),
          })
        }

        if (c.type === Dropdown || childProps.isMenuDropdown) {
          return React.cloneElement(c as React.ReactElement<IDropdown>, {
            variant: 'ghost',
            buttonClassName: menuItemClass,
          })
        }

        if (c.type === Text) {
          return React.cloneElement<IText>(c, {
            className: 'px-1.5 py-1',
            variant: 'label',
            theme: 'neutral',
          })
        }

        if (c.type === 'hr') {
          return React.cloneElement(c, { className: 'my-1' })
        }

        return c
      })}
    </div>
  )
}
