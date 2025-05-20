'use client'

import classNames from 'classnames'
import Image from 'next/image'
import React, { type FC } from 'react'
import { ArrowLineLeft, ArrowLineRight, Sidebar } from '@phosphor-icons/react'
import { useDashboard, useOrg } from '@/stratus/context'
import { Button, Text } from '@/stratus/components/common'
import { initialsFromString } from '@/utils'
import { DashboardSidebarNav } from './DashboardSidebarNav'
import { Logo } from './Logo'
import './Dashboard.css'

interface IDashboard {
  children: React.ReactNode
}

export const Dashboard: FC<IDashboard> = ({ children }) => {
  const { isSidebarOpen } = useDashboard()
  const { org } = useOrg()

  return (
    <div
      className={classNames('dashboard divide-x', {
        'is-open': isSidebarOpen,
      })}
    >
      <aside className="bg-cool-grey-50 dark:bg-dark-grey-200 flex flex-col">
        <header className="flex items-center justify-between">
          <Logo />
          <div className="md:hidden">
            <MobileDashboardSidebarButton />
          </div>
        </header>
        <div className="p-4 flex flex-col gap-4">
          <div className="flex" style={{ height: '56px' }}>
            <div
              className={classNames(
                'm-auto flex gap-4 w-full items-center border rounded-md overflow-hidden transition-all',
                {
                  'px-4 py-1.5 ': isSidebarOpen,
                  'p-1 w-[40px] h-[40px] ': !isSidebarOpen,
                }
              )}
            >
              <span
                className={classNames(
                  'flex items-center justify-center rounded-md bg-cool-grey-200 text-cool-grey-600 dark:bg-dark-grey-300 dark:text-white/50 font-medium font-sans transition-all',
                  {
                    'p-2': !org?.logo_url,
                    'w-[40px] h-[40px]': isSidebarOpen,
                    'w-[30px] h-[30px]': !isSidebarOpen,
                  }
                )}
              >
                {org?.logo_url ? (
                  <Image
                    className="rounded-md"
                    height={40}
                    width={40}
                    src={org?.logo_url}
                    alt="Logo"
                  />
                ) : (
                  initialsFromString(org.name)
                )}
              </span>
              <div
                className={classNames('flex flex-col transition-all', {
                  'opacity-100': isSidebarOpen,
                  'opacity-0': !isSidebarOpen,
                })}
              >
                <Text
                  weight="strong"
                  variant="subtext"
                  className="text-nowrap truncate w-fit"
                >
                  {org.name}
                </Text>
                <Text variant="subtext">{org?.status}</Text>
              </div>
            </div>
          </div>

          <DashboardSidebarNav />
        </div>
      </aside>
      <main>
        <header className="py-3 px-4 border-b flex items-center">
          <div className="flex items-center gap-8">
            <div className="md:hidden">
              <MobileDashboardSidebarButton />
            </div>
            <div className="hidden md:block">
              <DashboardSidebarButton />
            </div>
          </div>
        </header>
        {children}
      </main>
    </div>
  )
}

export const DashboardSidebarButton: FC = () => {
  const { toggleSidebar } = useDashboard()

  return (
    <Button variant="ghost" className="!px-1.5" onClick={toggleSidebar}>
      <Sidebar size="20" />
    </Button>
  )
}

export const MobileDashboardSidebarButton: FC = () => {
  const { isSidebarOpen, toggleSidebar } = useDashboard()

  return (
    <Button variant="ghost" className="!px-2" onClick={toggleSidebar}>
      {isSidebarOpen ? <ArrowLineLeft /> : <ArrowLineRight />}
    </Button>
  )
}
