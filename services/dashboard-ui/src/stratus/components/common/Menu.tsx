import classNames from 'classnames'
import React, { type FC } from 'react'
import { Button, type IButton } from './Button'
import { Link, type ILink } from './Link'
import './Menu.css'

interface IMenu extends Omit<React.HTMLAttributes<HTMLDivElement>, 'role'> {}

export const Menu: FC<IMenu> = ({ className, children, ...props }) => {
  return (
    <div
      className={classNames('menu', {
        [`${className}`]: Boolean(className),
      })}
      role="menu"
      {...props}
    >
      {React.Children.map(children, (c) =>
        React.isValidElement(c) && (c.type === Button || c.type === Link)
          ? React.cloneElement<IButton | ILink>(c, {
              variant: 'ghost',
            })
          : c
      )}
    </div>
  )
}
