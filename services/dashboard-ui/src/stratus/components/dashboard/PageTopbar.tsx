import classNames from 'classnames'
import React, { type FC } from 'react'
import { MobileSidebarButton, SidebarButton } from './Sidebar'

export interface IPageTopbar extends React.HTMLAttributes<HTMLDivElement> {}

export const PageTopbar: FC<IPageTopbar> = ({
  className,
  children,
  ...props
}) => {
  return (
    <header
      className={classNames(
        'py-3 px-4 border-b flex items-center h-[60px] w-full',
        {
          [`${className}`]: Boolean(className),
        }
      )}
      {...props}
    >
      <div className="flex items-center gap-2 w-full">
        <div className="md:hidden">
          <MobileSidebarButton />
        </div>
        <div className="hidden md:block">
          <SidebarButton />
        </div>
        {children}
      </div>
    </header>
  )
}
