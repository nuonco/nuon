'use client'

import React, { type FC, useState, useEffect } from 'react'
import classNames from 'classnames'
import { ArrowLineLeftIcon, ArrowLineRightIcon } from '@phosphor-icons/react'
import { getIsSidebarOpenFromCookie } from "@/actions/layout/main-sidebar-cookie"
import { Button } from '@/components/Button'
import { Logo } from '@/components/Logo'
import { SignOutButton } from '@/components/Profile'
import { NuonVersions, type TNuonVersions } from '@/components/NuonVersions'
import { getAPIVersion } from '@/lib'
import { VERSION } from '@/utils'

const NoOrgs: FC = () => {
  return (
    <div className="flex-1">
      {/* Empty container - journey modal handles all user communication */}
    </div>
  )
}

const AppLayout: FC<{
  children: React.ReactNode
  isSidebarOpen: boolean
  setIsSidebarOpen: (open: boolean) => void
  versions: TNuonVersions
}> = ({ children, isSidebarOpen, setIsSidebarOpen, versions }) => {
  return (
    <div
      className={classNames('layout', {
        'layout--open': isSidebarOpen,
      })}
    >
      <aside className="layout_aside dashboard_sidebar border-r flex flex-col">
        <header className="flex flex-col gap-4">
          <div className="border-b flex items-center justify-between px-4 pt-6 pb-4 h-[75px]">
            {isSidebarOpen ? <Logo /> : null}
            <Button
              className={classNames('p-1.5', {
                'm-auto': !isSidebarOpen,
              })}
              hasCustomPadding
              variant="ghost"
              onClick={() => setIsSidebarOpen(!isSidebarOpen)}
            >
              {isSidebarOpen ? <ArrowLineLeftIcon /> : <ArrowLineRightIcon />}
            </Button>
          </div>

          <div className="px-4">
            {/* Empty org switcher area - no orgs to switch between */}
          </div>
        </header>

        <div className="dashboard_nav flex-auto flex flex-col justify-between px-4 pb-6 pt-8">
          <div className="flex gap-3">
            {/* No navigation items - user needs org first */}
          </div>

          <div className="flex flex-col gap-2">
            <SignOutButton isSidebarOpen={isSidebarOpen} />
            {isSidebarOpen ? (
              <NuonVersions
                className="justify-center py-2 flex-initial"
                {...versions}
              />
            ) : null}
          </div>
        </div>
      </aside>
      <div className="layout_content dashboard_content">{children}</div>
    </div>
  )
}

export const AppHomePage: FC = () => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(true)
  const [versions, setVersions] = useState<TNuonVersions>({
    api: { git_ref: 'unknown' as any, version: 'unknown' as any },
    ui: { version: VERSION as any },
  })

  useEffect(() => {
    // Get sidebar state from cookie (client-side fallback)
    const sidebarCookie = document.cookie
      .split('; ')
      .find((row) => row.startsWith('is-sidebar-open='))
    setIsSidebarOpen(sidebarCookie?.split('=')[1] === 'true')

    // Get API version
    getAPIVersion().then(({ data: apiVersion }) => {
      if (apiVersion) {
        setVersions({
          api: apiVersion as any,
          ui: { version: VERSION as any },
        })
      }
    })
  }, [])


  return (
    <>
      <AppLayout
        isSidebarOpen={isSidebarOpen}
        setIsSidebarOpen={setIsSidebarOpen}
        versions={versions}
      >
        <div className="h-full">
          <NoOrgs />
        </div>
      </AppLayout>
    </>
  )
}

