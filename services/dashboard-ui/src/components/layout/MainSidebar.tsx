'use client'

import React from 'react'
import { useSidebar } from '@/hooks/use-sidebar'
import { Icon } from '@/components/common/Icon'
import { SidebarLogo } from '@/components/common/Logo/Logo'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { Text } from '@/components/common/Text'
import { MainNav } from '@/components/navigation/MainNav'
import { OrgSwitcher } from '@/components/orgs/OrgSwitcher'
import { UserDropdown } from '@/components/users/UserDropdown'
import type { TNuonVersion } from '@/types'
import { cn } from '@/utils/classnames'
import { MainSidebarButton } from './MainSidebarButton'

export const MainSidebar = ({ versions }: { versions: TNuonVersion }) => {
  const { isSidebarOpen } = useSidebar()
  return (
    <aside
      className={cn(
        'md:transition-[width] md:duration-fast md:ease-cubic md:w-[4.5rem]',
        'h-screen w-screen flex-none flex flex-col border-r',
        'bg-gradient',
        { 'md:w-[17.5rem]': isSidebarOpen }
      )}
    >
      <header className="flex items-center justify-between h-16 px-4">
        <SidebarLogo />
        <div className="md:hidden">
          <MainSidebarButton variant="mobile" />
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
      {isSidebarOpen ? (
        <TransitionDiv
          className="flex gap-4 items-center justify-center justify-self-end p-4 fade"
          isVisible={isSidebarOpen}
        >
          <Text variant="subtext">
            API: <b>{versions?.api?.version}</b>
          </Text>
          <Text variant="subtext">
            UI: <b>{versions?.ui?.version}</b>
          </Text>
        </TransitionDiv>
      ) : null}
    </aside>
  )
}
