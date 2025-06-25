import React from 'react'
import { cn } from '@/stratus/components/helpers'
import { Button, type IButton } from './Button'
import { Dropdown, type IDropdown } from './Dropdown'
import { Link, type ILink } from './Link'
import './Menu.css'

interface IMenu extends Omit<React.HTMLAttributes<HTMLDivElement>, 'role'> {}

export const Menu = ({ className, children, ...props }: IMenu) => {
  return (
    <div className={cn('menu', className)} role="menu" {...props}>
      {React.Children.map(children, (c) =>
        React.isValidElement(c) &&
        (c.type === Button || c.type === Dropdown || c.type === Link)
          ? React.cloneElement<IButton | IDropdown | ILink>(c, {
              variant: 'ghost',
            })
          : c
      )}
    </div>
  )
}
