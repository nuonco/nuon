'use client'

import { useState } from 'react'
import { Avatar } from '@/components/common/Avatar'
import { Button } from '@/components/common/Button'
import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { Text } from '@/components/common/Text'
import { GITHUB_APP_NAME } from '@/configs/github-app'
import { useSidebar } from '@/hooks/use-sidebar'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { useQuery } from '@/hooks/use-query'
import type { TOrg } from '@/types'
import { cn } from '@/utils/classnames'
import './OrgAvatar.css'

interface IOrgSwitcher
  extends Omit<IDropdown, 'buttonText' | 'children' | 'id'> {}

export const OrgSwitcher = ({}: IOrgSwitcher) => {
  const { isSidebarOpen } = useSidebar()
  const { org } = useOrg()
  return (
    <Dropdown
      alignment="overlay"
      className="w-full md:w-[248px] duration-fastest transition-all"
      buttonClassName={cn(
        'w-full text-left duration-fastest transition-all !text-foreground !border-[var(--border-color)]',
        {
          '!px-4 !py-1.5 ': isSidebarOpen,
          '!p-[3px] !size-10 ': !isSidebarOpen,
        }
      )}
      buttonText={
        <OrgSummary isButtonSummary isSidebarOpen={isSidebarOpen} org={org} />
      }
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
        <div className="p-3 border-b">
          <OrgSummary org={org} />
          <ID className="!flex mt-2">{org.id}</ID>
        </div>
        <div className="px-3 py-4 flex flex-col gap-4">
          <div className="flex justify-between items-center">
            <Text variant="subtext" weight="strong">
              GitHub connections
            </Text>

            <Button
              className="!px-2 text-primary-600 dark:text-primary-400"
              variant="ghost"
              size="sm"
              href={`https://github.com/apps/${GITHUB_APP_NAME}/installations/new?state=${org.id}`}
            >
              <Icon variant="Plus" /> Add
            </Button>
          </div>
          {org?.vcs_connections?.map((vcs) => (
            <Text
              key={vcs?.id}
              className="!flex items-center gap-2"
              family="mono"
              variant="subtext"
              theme="neutral"
            >
              <Icon variant="GitHub" />{' '}
              {vcs?.github_account_name || vcs?.github_install_id}
            </Text>
          ))}
        </div>
        <hr className="border-dashed mx-4" />
        <div className="px-1 py-4 flex flex-col gap-1.5">
          <div className="px-2">
            <Text variant="subtext" weight="strong">
              Organizations
            </Text>
          </div>
          <OrgsNav />
        </div>
      </Menu>
    </Dropdown>
  )
}

const OrgAvatar = ({
  isButtonSummary = false,
  org,
}: {
  isSidebarOpen?: boolean
  isButtonSummary?: boolean
  org: TOrg
}) => {
  const { isSidebarOpen } = useSidebar()
  return (
    <div className={cn({ 'org-avatar-summary relative': isButtonSummary })}>
      <Avatar
        {...(org?.logo_url ? { src: org?.logo_url } : { name: org.name })}
        size={!isSidebarOpen && isButtonSummary ? 'md' : 'xl'}
      />
      {isButtonSummary ? (
        <Status
          className={cn('absolute right-0 top-0 transition-all', {
            'opacity-0': isSidebarOpen,
            'delay-fastest opacity-100': !isSidebarOpen,
          })}
          status={org?.status}
          isWithoutText
        />
      ) : null}
    </div>
  )
}

interface IOrgSummary {
  isSidebarOpen?: boolean
  isButtonSummary?: boolean
  org: TOrg
}

const OrgSummary = ({
  isSidebarOpen = true,
  isButtonSummary = false,
  org,
}: IOrgSummary) => {
  return (
    <div className="flex gap-4 items-center overflow-hidden">
      <OrgAvatar {...{ isButtonSummary, isSidebarOpen, org }} />
      <div
        className={cn('transition-all max-w-full overflow-hidden', {
          'md:opacity-100': isSidebarOpen,
          'md:opacity-0': !isSidebarOpen,
        })}
      >
        <Text
          weight="strong"
          variant="subtext"
          className="!flex items-center gap-1.5 text-nowrap"
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
  const enablePaginationCount = 6
  const [offset, setOffset] = useState<number>(0)
  const [limit, setLimit] = useState<number>(10)
  const [searchTerm, setSearchTerm] = useState<string>('')

  const params = useQueryParams({ offset, limit, q: searchTerm })
  const {
    data: orgs,
    isLoading,
    error,
  } = useQuery<TOrg[]>({
    path: `/api/orgs${params}`,
  })

  return (
    <>
      {isLoading
        ? Array.from({ length: 5 }).map((_, i) => <LoadingOrgSummary key={i} />)
        : orgs?.map((o) => (
            <Link
              key={o?.id}
              className="!h-fit !block w-full"
              href={`/${o?.id}`}
              variant="ghost"
            >
              <OrgSummary org={o} />
            </Link>
          ))}
    </>
  )
}
