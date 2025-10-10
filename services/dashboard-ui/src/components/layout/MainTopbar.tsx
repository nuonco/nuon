import React from 'react'
import { UserDropdown } from '@/components/users/UserDropdown'
import { cn } from '@/utils/classnames'
import { MainSidebarButton } from './MainSidebarButton'

export interface IMainTopbar extends React.HTMLAttributes<HTMLDivElement> {
  hideSidebarButtons?: boolean
}

export const MainTopbar = ({
  className,
  children,
  hideSidebarButtons = false,
  ...props
}: IMainTopbar) => {
  return (
    <header
      className={cn(
        'py-3 px-4 border-b flex shrink-0 items-center h-[60px] w-full overflow-x-auto md:overflow-visible',
        className
      )}
      {...props}
    >
      <div className="flex items-center gap-2 w-full">
        {hideSidebarButtons ? null : (
          <>
            <div className="md:hidden">
              <MainSidebarButton variant="mobile" />
            </div>
            <div className="hidden md:block">
              <MainSidebarButton />
            </div>
          </>
        )}
        {children}

        <div className="hidden md:flex ml-auto">
          <UserDropdown alignment="right" />
        </div>
      </div>
    </header>
  )
}
