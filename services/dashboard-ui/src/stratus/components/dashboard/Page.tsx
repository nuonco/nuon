import React, { type FC } from 'react'
import { UserDropdown } from '@/stratus/components/user'
import { BreadcrumbNav, type IBreadcrumbNav } from './BreadcrumbNav'
import { PageTopbar } from './PageTopbar'

interface IPage {
  breadcrumb: IBreadcrumbNav
  children: React.ReactNode
}

export const Page: FC<IPage> = ({ breadcrumb, children }) => {
  return (
    <main className="flex flex-col h-screen">
      <PageTopbar>
        <div className="flex items-center justify-between w-full">
          <BreadcrumbNav {...breadcrumb} />

          <div className="hidden md:block">
            <UserDropdown alignment="right" />
          </div>
        </div>
      </PageTopbar>
      {children}
    </main>
  )
}
