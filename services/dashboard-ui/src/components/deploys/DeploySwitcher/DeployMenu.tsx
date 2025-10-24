'use client'

import { useEffect, useState } from 'react'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Menu, type IMenu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import { useInstall } from '@/hooks/use-install'
import { useScrollToBottom } from '@/hooks/use-on-scroll-to-bottom'
import { useOrg } from '@/hooks/use-org'
import { useQueryParams } from '@/hooks/use-query-params'
import { useQuery } from '@/hooks/use-query'
import type { TDeploy } from '@/types'
import { cn } from '@/utils/classnames'
import { DeploySummary } from './DeploySummary'
import { DeploysSkeleton } from './DeploysSkeleton'

interface IDeployMenu extends Omit<IMenu, 'children'> {
  activeDeployId: string
  componentId: string
}

export const DeployMenu = ({ activeDeployId, componentId }: IDeployMenu) => {
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
                  href={`/${org.id}/installs/${install.id}/components/${componentId}/${deploy?.id}`}
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
