import React from 'react'
import { Logo } from '@/components/common/Logo'
import {
  BreadcrumbNav,
  type IBreadcrumbNav,
} from '@/components/navigation/Breadcrumb'
import { cn } from '@/utils/classnames'
import { MainTopbar } from './MainTopbar'

interface IPageLayout extends React.HTMLAttributes<HTMLDivElement> {
  breadcrumb?: IBreadcrumbNav
  children: React.ReactNode
  isScrollable?: boolean
  variant?: 'dashboard-page' | 'single-page'
}

export const PageLayout = ({
  breadcrumb,
  className,
  children,
  isScrollable = false,
  variant = 'dashboard-page',
  ...props
}: IPageLayout) => {
  return (
    <main className="flex flex-col h-screen w-full">
      <MainTopbar hideSidebarButtons={variant === 'single-page'}>
        {variant === 'single-page' ? <Logo /> : null}
        {breadcrumb ? <BreadcrumbNav {...breadcrumb} /> : null}
      </MainTopbar>
      <div
        className={cn(
          'flex-auto flex flex-col overflow-y-auto md:overflow-hidden',
          {
            'md:!overflow-y-auto': isScrollable,
          },
          className
        )}
        {...props}
      >
        {children}
      </div>
    </main>
  )
}
