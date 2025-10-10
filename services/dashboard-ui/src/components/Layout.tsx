'use client'

import classNames from 'classnames'
import { useSearchParams } from 'next/navigation'
import React, { type FC, useState } from 'react'
import {
  ArrowLineLeftIcon,
  ArrowLineRightIcon,
  ListIcon,
  XIcon,
} from '@phosphor-icons/react/dist/ssr'
import { setSidebarCookie } from '@/actions/layout/main-sidebar-cookie'
import { AdminModal } from '@/components/AdminModal'
import { Button } from '@/components/Button'
import { Logo } from '@/components/Logo'
import { OrgSwitcher } from '@/components/orgs/OldOrgSwitcher'
import { SignOutButton } from '@/components/Profile'
import { MainNav } from '@/components/Nav'
import { NuonVersions, type TNuonVersions } from '@/components/NuonVersions'
import type { TOrg } from '@/types'

interface ILayout {
  children: React.ReactElement
  orgs: Array<TOrg>
  versions: TNuonVersions
}

export const OldLayout: FC<ILayout> = ({ children, orgs, versions }) => {
  const [isOpen, setIsOpen] = useState(true)

  return (
    <div className="flex min-h-screen">
      <aside
        className={classNames('dashboard_sidebar flex flex-col w-full', {
          'md:w-72 md:min-w-72 md:max-w-72': isOpen,
          'md:w-[72px] md:min-w-[72px] md:max-w-[72px]': !isOpen,
        })}
      >
        <header className="flex flex-col gap-4">
          <div className="border-b flex items-center justify-between px-4 pt-6 pb-4 h-[75px]">
            {isOpen ? <Logo /> : null}
            <Button
              className={classNames('p-1.5', {
                'm-auto': !isOpen,
              })}
              hasCustomPadding
              variant="ghost"
              onClick={() => {
                setIsOpen(!isOpen)
              }}
            >
              {isOpen ? <ArrowLineLeftIcon /> : <ArrowLineRightIcon />}
            </Button>
          </div>

          <div className="px-4">
            <OrgSwitcher initOrgs={orgs} isSidebarOpen={isOpen} />
          </div>
        </header>

        <div className="dashboard_nav flex-auto flex flex-col justify-between px-4 pb-6 pt-8">
          <div className="flex gap-3">
            <MainNav isSidebarOpen={isOpen} />
          </div>

          <div className="flex flex-col gap-2">
            <SignOutButton isSidebarOpen={isOpen} />
            <AdminModal isSidebarOpen={isOpen} />
            {isOpen ? (
              <NuonVersions
                className="justify-center py-2 flex-initial"
                {...versions}
              />
            ) : null}
          </div>
        </div>
      </aside>
      <div className="dashboard_content h-screen md:flex-auto md:border-l">
        {children}
      </div>
    </div>
  )
}

export const Layout: FC<{
  children: React.ReactNode
  isSidebarOpen: boolean
  orgs: Array<TOrg>
  versions: TNuonVersions
}> = ({ children, isSidebarOpen, orgs, versions }) => {
  const [isOpen, setIsOpen] = useState(isSidebarOpen)
  const searchParams = useSearchParams()

  return (
    <div
      className={classNames('layout', {
        'layout--open': isOpen,
      })}
    >
      <aside className="layout_aside dashboard_sidebar border-r flex flex-col">
        <header className="flex flex-col gap-4">
          <div className="border-b flex items-center justify-between px-4 pt-6 pb-4 h-[75px]">
            {isOpen ? <Logo /> : null}
            <Button
              className={classNames('p-1.5', {
                'm-auto': !isOpen,
              })}
              hasCustomPadding
              variant="ghost"
              onClick={() => {
                setIsOpen(!isOpen)
                setSidebarCookie(!isOpen)
              }}
            >
              {isOpen ? (
                <>
                  <XIcon className="md:hidden" />
                  <ArrowLineLeftIcon className="hidden md:block" />
                </>
              ) : (
                <ArrowLineRightIcon />
              )}
            </Button>
          </div>
        </header>

        <div className="dashboard_nav flex-auto flex flex-col justify-between px-4 pb-6 pt-8">
          <div className="flex flex-col gap-8">
            <OrgSwitcher initOrgs={orgs} isSidebarOpen={isOpen} />

            <div className="flex gap-3">
              <MainNav isSidebarOpen={isOpen} />
            </div>
          </div>

          <div className="flex flex-col gap-2">
            <SignOutButton isSidebarOpen={isOpen} />
            <AdminModal
              isSidebarOpen={isOpen}
              isModalOpen={searchParams?.get('admin')}
            />
            {isOpen ? (
              <NuonVersions
                className="justify-center py-2 flex-initial"
                {...versions}
              />
            ) : (
              <div className="w-[32px] h-[32px]" />
            )}
          </div>
        </div>
      </aside>
      <div className="layout_content dashboard_content relative">
        <Button
          className={classNames('p-1.5 absolute top-6 left-4 md:hidden', {})}
          hasCustomPadding
          variant="ghost"
          onClick={() => {
            setIsOpen(!isOpen)
          }}
        >
          {isOpen ? <ArrowLineLeftIcon /> : <ListIcon />}
        </Button>
        {children}
      </div>
    </div>
  )
}
