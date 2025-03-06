'use client'

import classNames from 'classnames'
import React, { type FC, useState } from 'react'
import { ArrowLineLeft, ArrowLineRight } from '@phosphor-icons/react/dist/ssr'
import { AdminModal } from '@/components/AdminModal'
import { Button } from '@/components/Button'
import { Logo } from '@/components/Logo'
import { OrgSwitcher } from '@/components/OrgSwitcher'
import { SignOutButton } from '@/components/Profile'
import { MainNav } from '@/components/Nav'
import { NuonVersions, type TNuonVersions } from '@/components/NuonVersions'
import type { TOrg } from '@/types'

interface ILayout {
  children: React.ReactElement
  org: TOrg
  orgs: Array<TOrg>
  versions: TNuonVersions
  featureFlags?: Record<string, boolean>
}

export const Layout: FC<ILayout> = ({
  children,
  org,
  orgs,
  versions,
  featureFlags,
}) => {
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
              {isOpen ? <ArrowLineLeft /> : <ArrowLineRight />}
            </Button>
          </div>

          <div className="px-4">
            <OrgSwitcher initOrg={org} initOrgs={orgs} isSidebarOpen={isOpen} />
          </div>
        </header>

        <div className="dashboard_nav flex-auto flex flex-col justify-between px-4 pb-6 pt-8">
          <div className="flex gap-3">
            <MainNav
              orgId={org?.id}
              isSidebarOpen={isOpen}
              featureFlags={featureFlags}
            />
          </div>

          <div className="flex flex-col gap-2">
            <SignOutButton isSidebarOpen={isOpen} />
            <AdminModal orgId={org?.id} isSidebarOpen={isOpen} org={org} />
            {isOpen ? (
              <NuonVersions
                className="justify-center py-2 flex-initial"
                {...versions}
              />
            ) : null}
          </div>
        </div>
      </aside>
      <div className="dashboard_content h-screen flex-auto md:border-l">
        {children}
      </div>
    </div>
  )
}
