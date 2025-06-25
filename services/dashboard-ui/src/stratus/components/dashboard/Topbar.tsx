import React from 'react'
import { cn } from '@/stratus/components/helpers'
import { MobileSidebarButton, SidebarButton } from './Sidebar'

export interface ITopbar extends React.HTMLAttributes<HTMLDivElement> {}

export const Topbar = ({ className, children, ...props }: ITopbar) => {
  return (
    <header
      className={cn(
        'py-3 px-4 border-b flex shrink-0 items-center h-[60px] w-full overflow-x-auto md:overflow-visible',
        className
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
