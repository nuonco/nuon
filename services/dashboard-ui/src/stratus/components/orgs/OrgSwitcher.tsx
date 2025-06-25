'use client'

import React, { useEffect, useState } from 'react'
import { cn } from '@/stratus/components/helpers'
import {
  Avatar,
  Dropdown,
  Icon,
  Link,
  Menu,
  Skeleton,
  Text,
  type IDropdown,
} from '@/stratus/components/common'
import { Status } from '@/stratus/components/statuses'
import { useDashboard, useOrg } from '@/stratus/context'
import type { TOrg } from '@/types'
import { GITHUB_APP_NAME } from '@/utils'

interface IOrgSwitcher
  extends Omit<IDropdown, 'buttonText' | 'children' | 'id'> {}

export const OrgSwitcher = ({}: IOrgSwitcher) => {
  const { isSidebarOpen } = useDashboard()
  const { org } = useOrg()
  return (
    <Dropdown
      alignment="overlay"
      className="w-[248px]"
      buttonClassName={cn(
        'w-full text-left transition-all !border-cool-grey-300 dark:!border-dark-grey-500',
        {
          '!px-4 !py-1.5 ': isSidebarOpen,
          '!p-[3px] !size-10 ': !isSidebarOpen,
        }
      )}
      buttonText={<OrgSummary isSidebarOpen={isSidebarOpen} org={org} />}
      icon={isSidebarOpen ? <Icon variant="CaretUpDown" /> : null}
      id="org-switcher"
      position="overlay"
      variant="ghost"
    >
      <Menu
        className="w-[248px] h-[308px] overflow-y-scroll overflow-x-hidden focus:outline-primay-400 !p-0"
        tabIndex={-1}
        style={{ scrollbarGutter: 'stable' }}
      >
        <div className="p-3">
          <OrgSummary org={org} />
        </div>
        <div className="px-3 py-4 flex flex-col gap-4">
          <div className="flex justify-between">
            <Text variant="subtext" weight="strong">
              GitHub connections
            </Text>
            <Text variant="subtext">
              <Link
                className="flex items-center gap-1.5"
                href={`https://github.com/apps/${GITHUB_APP_NAME}/installations/new?state=${org.id}`}
              >
                <Icon variant="Plus" /> Add
              </Link>
            </Text>
          </div>
          {org?.vcs_connections?.map((vcs) => (
            <Text
              key={vcs?.id}
              className="flex items-center gap-2"
              family="mono"
              variant="subtext"
              theme="muted"
            >
              <Icon variant="GitHub" /> {vcs?.github_install_id}
            </Text>
          ))}
        </div>
        <hr className="border-dashed mx-4" />
        <div className="px-1 py-4 flex flex-col gap-1.5">
          <div className="px-2">
            <Text variant="subtext" weight="strong">
              GitHub connections
            </Text>
          </div>
          <OrgsNav />
        </div>
      </Menu>
    </Dropdown>
  )
}

interface IOrgSummary {
  isSidebarOpen?: boolean
  org: TOrg
}

const OrgSummary = ({ isSidebarOpen = true, org }: IOrgSummary) => {
  return (
    <div className="flex gap-4 items-center overflow-hidden">
      <Avatar
        {...(org?.logo_url ? { src: org?.logo_url } : { name: org.name })}
        size={isSidebarOpen ? 'xl' : 'md'}
      />
      <div
        className={cn('transition-all max-w-full overflow-hidden', {
          'opacity-100': isSidebarOpen,
          'opacity-0': !isSidebarOpen,
        })}
      >
        <Text
          weight="strong"
          variant="subtext"
          className="text-nowrap flex items-center gap-1.5"
        >
          {org.sandbox_mode && (
            <Icon
              variant="TestTube"
              className="!w-[12px] !h-[12px] shrink-0"
              size="12"
            />
          )}
          <span className="truncate">{org.name}</span>
        </Text>
        <Status status={org?.status} />
      </div>
    </div>
  )
}

const LoadingOrgSummary = () => {
  return (
    <div className="flex gap-4 items-center p-2 w-full">
      <Avatar size="xl" isLoading />
      <div className="flex flex-col gap-1 transition-all w-full">
        <Skeleton height="14px" width="80%" />
        <Skeleton height="14px" width="40%" />
      </div>
    </div>
  )
}

interface IOrgsNav {}

const OrgsNav = ({}: IOrgsNav) => {
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
        ? [0, 1, 2, 3].map((k) => <LoadingOrgSummary key={k} />)
        : orgs?.map((o) => (
            <Link
              key={o?.id}
              className="!h-fit !block"
              href={`/stratus/${o?.id}`}
              variant="ghost"
            >
              <OrgSummary org={o} />
            </Link>
          ))}
    </>
  )
}
