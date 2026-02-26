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
      {React.Children.map(children, (c) =>
        React.isValidElement(c)
          ? c.type === Button ||
            c.type === Link ||
            (c as any)?.props?.isMenuButton
            ? React.cloneElement<IButtonAsButton | ILink>(c, {
                variant: 'ghost',
                className: cn(
                  '!p-2 text-sm !leading-none h-8 w-full flex justify-between',
                  c?.props.className,
                  {
                    '!text-red-600 dark:!text-red-400':
                      c?.props?.variant === 'danger',
                  }
                ),
              })
            : c.type === Dropdown || (c as any).isMenuDropdown
              ? React.cloneElement(c as React.ReactElement<IDropdown>, {
                  variant: 'ghost',
                  buttonClassName:
                    '!p-2 text-sm !leading-none h-8 w-full flex justify-between',
                })
              : c.type === Text
                ? React.cloneElement<IText>(c, {
                    className: 'px-1.5 py-1',
                    variant: 'label',
                    theme: 'neutral',
                  })
                : c.type === 'hr'
                  ? React.cloneElement<IText>(c, {
                      className: 'my-1',
                    })
                  : c
          : null
      )}
    </div>
  )
}
