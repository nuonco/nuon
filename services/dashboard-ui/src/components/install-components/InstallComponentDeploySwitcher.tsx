'use client'

import { useEffect, useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Dropdown, type IDropdown } from '@/components/common/Dropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Menu, type IMenu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useInstall } from '@/hooks/use-install'
import { useScrollToBottom } from '@/hooks/use-on-scroll-to-bottom'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { useQuery } from '@/hooks/use-query'
import type { TDeploy } from '@/types'
import { cn } from '@/utils/classnames'

interface IInstallComponentDeploySwitcher
  extends Omit<IDropdown, 'children' | 'id' | 'buttonText'> {
  componentId: string
  deployId: string
}

export const InstallComponentDeploySwitcher = ({
  alignment = 'right',
  componentId,
  deployId,
  ...props
}: IInstallComponentDeploySwitcher) => {
  return (
    <Dropdown
      id="deploy-switcher"
      alignment={alignment}
      buttonText="Latest deploys"
      {...props}
    >
      <InstallComponentDeployMenu
        activeDeployId={deployId}
        componentId={componentId}
      />
    </Dropdown>
  )
}

interface IInstallComponentDeployMenu extends Omit<IMenu, 'children'> {
  activeDeployId: string
  componentId: string
}

const InstallComponentDeployMenu = ({
  activeDeployId,
  componentId,
}: IInstallComponentDeployMenu) => {
  const limit = 8
  const [searchTerm, setSearchTerm] = useState('')
  const [offset, setOffset] = useState(0)
  const [deploys, setDeploys] = useState<TDeploy[]>([])
  const { install } = useInstall()
  const { org } = useOrg()
  const queryParams = useQueryParams({
    limit,
    offset,
  })
  const { data, error, headers, isLoading } = useQuery<TDeploy[]>({
    path: `/api/orgs/${org.id}/installs/${install.id}/components/${componentId}/deploys${queryParams}`,
    initData: [],
  })

  const scrollToBottom = useScrollToBottom({
    onScrollToBottom: () => {
      if (headers?.['x-nuon-page-next'] === 'true') {
        setOffset((prev) => {
          if (prev === 0) {
            return limit + 1
          } else {
            return prev + limit
          }
        })
      }
    },
  })

  useEffect(() => {
    setDeploys((prev) => {
      const deployMap = new Map(prev.map((deploy) => [deploy.id, deploy]))
      data.forEach((deploy) => deployMap.set(deploy.id, deploy))
      return Array.from(deployMap.values())
    })
    scrollToBottom.reset()
  }, [data])

  const filteredDeploys = deploys
    ? deploys.filter(
        (deploy) =>
          deploy.id.includes(searchTerm) ||
          deploy?.created_by?.email.includes(searchTerm) ||
          deploy?.status_v2?.status.includes(searchTerm)
      )
    : []

  return (
    <Menu className="w-100 !p-0">
      <div className="flex flex-col gap-2 p-2 border-b">
        <Text variant="label" theme="neutral">
          Latest deploys
        </Text>
        <SearchInput
          labelClassName="!w-full"
          className="w-full"
          value={searchTerm}
          onChange={setSearchTerm}
          placeholder="Search..."
        />
      </div>
      <div
        ref={scrollToBottom.elementRef}
        className="flex flex-col gap-2 p-2 max-h-56 overflow-y-auto"
      >
        {filteredDeploys?.length && !error ? (
          <>
            {filteredDeploys?.map((deploy, idx) => (
              <span key={deploy.id} className="rounded-lg border">
                <Button
                  className={cn('!p-2 !h-fit w-full', {
                    '!bg-primary-600/5 dark:!bg-primary-600/5':
                      deploy?.id === activeDeployId,
                  })}
                  href={`/${org.id}/installs/${install.id}/components/${componentId}/deploys/${deploy?.id}`}
                  variant="ghost"
                >
                  <DeploySummary deploy={deploy} isLatest={idx === 0} />
                </Button>
              </span>
            ))}
            {isLoading ? <DeploysSkeleton limit={limit} /> : null}
          </>
        ) : isLoading ? (
          <DeploysSkeleton />
        ) : (
          <EmptyState
            variant="history"
            emptyMessage="Unable to find any deployments."
            emptyTitle="No deploys found"
            size="sm"
          />
        )}
      </div>
    </Menu>
  )
}

const DeploySummary = ({
  deploy,
  isLatest = false,
}: {
  deploy: TDeploy
  isLatest?: boolean
}) => {
  return (
    <span className="flex flex-col w-full">
      <Text
        className="flex items-center justify-between"
        variant="subtext"
        weight="strong"
      >
        <span className="flex items-center gap-4">
          {deploy.id}
          {isLatest ? (
            <Badge theme="info" size="sm">
              Latest
            </Badge>
          ) : null}
        </span>
        <Time
          time={deploy.created_at}
          format="relative"
          variant="label"
          theme="neutral"
        />
      </Text>
      <span className="flex items-center gap-4 w-full">
        <Status status={deploy.status_v2?.status} />
        <Text variant="label" theme="neutral">
          {deploy?.created_by?.email}
        </Text>
      </span>
    </span>
  )
}

const DeploysSkeleton = ({ limit = 5 }: { limit?: number }) => {
  return Array.from({ length: limit }).map((_, idx) => (
    <span
      key={`deploy-skeleton-${idx}`}
      className="flex flex-col w-full gap-1 rounded-lg border p-2"
    >
      <span className="flex items-center justify-between">
        <Skeleton height="17px" width="160px" />
        <Skeleton height="14px" width="50px" />
      </span>
      <span className="flex items-center gap-4 w-full">
        <Skeleton height="17px" width="50px" />
        <Skeleton height="14px" width="70px" />
      </span>
    </span>
  ))
}
