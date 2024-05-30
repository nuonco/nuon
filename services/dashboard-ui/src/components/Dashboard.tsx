import React, { type FC } from 'react'
import { Logo, Link, ProfileDropdown } from '@/components'

export const DashboardHeader: FC = () => {
  return (
    <header className="flex flex-wrap items-center justify-between gap-6 pb-6 border-b">
      <div className="flex items-center gap-6">
        <Logo />
      </div>

      <div className="flex gap-4 items-center">
        <Link className="text-sm" href="https://docs.nuon.co" target="_blank">
          Documentation
        </Link>

        <ProfileDropdown />
      </div>
    </header>
  )
}

export const Dashboard: FC<{ children: React.ReactElement }> = ({
  children,
}) => {
  return (
    <div className="flex flex-col gap-6 p-6 xl:px-24 w-full h-dvh overflow-auto">
      <DashboardHeader />
      {children}
    </div>
  )
}
