import type { ReactNode } from 'react'
import { Logo } from '@/components/common/Logo'
import { BreadcrumbNav } from '@/components/navigation/Breadcrumb'
import { cn } from '@/utils/classnames'
import { MainTopbar } from './MainTopbar'

interface IPageLayout extends React.HTMLAttributes<HTMLDivElement> {
  children: ReactNode
  hideBreadcrumbs?: boolean
  isScrollable?: boolean
  variant?: 'dashboard-page' | 'single-page'
}

export const PageLayout = ({
  className,
  children,
  hideBreadcrumbs = false,
  isScrollable = false,
  variant = 'dashboard-page',
  ...props
}: IPageLayout) => {
  return (
    <main className="flex flex-col h-screen w-full">
      <MainTopbar hideSidebarButtons={variant === 'single-page'}>
        {variant === 'single-page' ? <Logo /> : null}
        {hideBreadcrumbs ? null : <BreadcrumbNav />}
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
