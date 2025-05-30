'use client'

import classNames from 'classnames'
import React, { type FC, useEffect, useState } from 'react'
import { CaretUpDown, GithubLogo } from '@phosphor-icons/react'
import {
  Avatar,
  Dropdown,
  Link,
  Menu,
  Skeleton,
  Text,
  type IDropdown,
} from '@/stratus/components/common'
import { useDashboard, useOrg } from '@/stratus/context'
import type { TOrg } from '@/types'

interface IOrgSwitcher
  extends Omit<IDropdown, 'buttonText' | 'children' | 'id'> {}

export const OrgSwitcher: FC<IOrgSwitcher> = () => {
  const { isSidebarOpen } = useDashboard()
  const { org } = useOrg()
  return (
    <Dropdown
      alignment="overlay"
      className="w-full"
      buttonClassName={classNames(
        'w-full text-left transition-all !border-cool-grey-300 dark:!border-dark-grey-500',
        {
          '!px-4 !py-1.5 ': isSidebarOpen,
          '!p-[3px] !size-10 ': !isSidebarOpen,
        }
      )}
      buttonText={<OrgSummary isSidebarOpen={isSidebarOpen} org={org} />}
      icon={isSidebarOpen ? <CaretUpDown /> : null}
      id="org-switcher"
      position="overlay"
      variant="ghost"
    >
      <Menu
        className="w-full !min-w-72 min-h-14 max-h-80 overflow-auto focus:outline-primay-400"
        tabIndex={-1}
      >
        <OrgSummary org={org} />
        <div className="py-4 flex flex-col">
          <div className="flex justify-between">
            <Text weight="strong">GitHub connections</Text>
          </div>
          {org?.vcs_connections?.map((vcs) => (
            <Text
              key={vcs?.id}
              className="flex items-center gap-2"
              family="mono"
              variant="subtext"
              theme="muted"
            >
              <GithubLogo size="16" /> {vcs?.github_install_id}
            </Text>
          ))}
        </div>
        <hr />
        <OrgsNav />
      </Menu>
    </Dropdown>
  )
}

interface IOrgSummary {
  isSidebarOpen?: boolean
  org: TOrg
}

const OrgSummary: FC<IOrgSummary> = ({ isSidebarOpen = true, org }) => {
  return (
    <div className="flex gap-4 items-center">
      <Avatar
        {...(org?.logo_url ? { src: org?.logo_url } : { name: org.name })}
        size={isSidebarOpen ? 'xl' : 'md'}
      />
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
  )
}

const LoadingOrgSummary: FC = () => {
  return (
    <div className="flex gap-4 items-center p-2 w-full">
      <Avatar size="xl" isLoading />
      <div
        className={classNames('flex flex-col gap-1 transition-all w-full', {})}
      >
        <Skeleton height="14px" width="80%" />
        <Skeleton height="14px" width="40%" />
      </div>
    </div>
  )
}

interface IOrgsNav {}

const OrgsNav: FC<IOrgsNav> = () => {
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string>()
  const [orgs, setOrgs] = useState<Array<TOrg>>()

  useEffect(() => {
    fetch('/api/orgs').then((r) =>
      r.json().then(({ data, error }) => {
        setIsLoading(false)
        if (error) {
          setError(error?.error)
        } else {
          setOrgs(data)
        }
      })
    )
  }, [])

  return (
    <>
      {isLoading
        ? [0, 1, 2].map((k) => <LoadingOrgSummary key={k} />)
        : orgs?.map((o) => (
            <Link
              key={o?.id}
              className="!h-fit"
              href={`/stratus/${o?.id}`}
              variant="ghost"
            >
              <OrgSummary org={o} />
            </Link>
          ))}
    </>
  )
}
