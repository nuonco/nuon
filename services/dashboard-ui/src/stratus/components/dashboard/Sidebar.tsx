'use client'

import classNames from 'classnames'
import Image from 'next/image'
import React, { type FC } from 'react'
import {
  CaretUp,
  ArrowLineLeft,
  ArrowLineRight,
  Sidebar as SidebarIcon,
} from '@phosphor-icons/react'
import { useDashboard, useOrg } from '@/stratus/context'
import { Avatar, Button, Text } from '@/stratus/components/common'
import { OrgSwitcher } from '@/stratus/components/orgs'
import { UserDropdown } from '@/stratus/components/user'
import { initialsFromString } from '@/utils'
import { MainNav } from './MainNav'
import { Logo } from './Logo'

interface ISidebar {}

export const Sidebar: FC<ISidebar> = () => {
  const { isSidebarOpen } = useDashboard()
  const { org } = useOrg()

  return (
    <aside className="bg-cool-grey-50 dark:bg-dark-grey-200 flex flex-col">
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
            icon={<CaretUp />}
            position="above"
          />
        </div>
      </div>
    </aside>
  )
}

export const SidebarButton: FC = () => {
  const { toggleSidebar } = useDashboard()

  return (
    <Button variant="ghost" className="!py-1 !px-1.5" onClick={toggleSidebar}>
      <SidebarIcon size="20" />
    </Button>
  )
}

export const MobileSidebarButton: FC = () => {
  const { isSidebarOpen, toggleSidebar } = useDashboard()

  return (
    <Button variant="ghost" className="!px-2" onClick={toggleSidebar}>
      {isSidebarOpen ? <ArrowLineLeft /> : <ArrowLineRight />}
    </Button>
  )
}
