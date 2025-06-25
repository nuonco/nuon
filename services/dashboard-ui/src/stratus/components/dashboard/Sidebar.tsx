'use client'

import React from 'react'
import { useDashboard } from '@/stratus/context'
import { Button, Icon } from '@/stratus/components/common'
import { OrgSwitcher } from '@/stratus/components/orgs'
import { UserDropdown } from '@/stratus/components/user'
import { MainNav } from './MainNav'
import { Logo } from './Logo'

export const Sidebar = () => {
  return (
    <aside className="bg-cool-grey-50 dark:bg-dark-grey-200 flex flex-col border-r">
      <header className="flex items-center justify-between">
        <Logo />
        <div className="md:hidden">
          <MobileSidebarButton />
        </div>
      </header>
      <div className="p-4 flex flex-col gap-4 flex-auto">
        <div className="flex h-14">
          <OrgSwitcher />
        </div>

        <MainNav />

        <div className="flex flex-auto items-end md:hidden">
          <UserDropdown
            alignment="left"
            className="!w-full"
            buttonClassName="!w-full"
            icon={<Icon variant="CaretUp" />}
            position="above"
          />
        </div>
      </div>
    </aside>
  )
}

export const SidebarButton = () => {
  const { toggleSidebar } = useDashboard()

  return (
    <Button variant="ghost" className="!py-1 !px-1.5" onClick={toggleSidebar}>
      <Icon variant="SidebarSimple" size="20" />
    </Button>
  )
}

export const MobileSidebarButton = () => {
  const { isSidebarOpen, toggleSidebar } = useDashboard()

  return (
    <Button variant="ghost" className="!px-2" onClick={toggleSidebar}>
      {isSidebarOpen ? (
        <Icon variant="ArrowLineLeft" />
      ) : (
        <Icon variant="ArrowLineRight" />
      )}
    </Button>
  )
}
