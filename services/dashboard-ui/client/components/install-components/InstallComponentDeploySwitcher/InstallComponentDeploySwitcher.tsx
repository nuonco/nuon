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
import type { TDeploy } from '@/types'
import { cn } from '@/utils/classnames'

interface IInstallComponentDeploySwitcher
  extends Omit<IDropdown, 'children' | 'id' | 'buttonText'> {
  componentId: string
  deployId: string
  deploys: TDeploy[]
  isLoading: boolean
  hasNext: boolean
  orgId: string
  installId: string
  onLoadMore: () => void
  onDeploysLoaded: (data: TDeploy[]) => void
  scrollRef: React.RefObject<HTMLDivElement>
}

export const InstallComponentDeploySwitcher = ({
  alignment = 'right',
  componentId,
  deployId,
  deploys,
  isLoading,
  hasNext,
  orgId,
  installId,
  onLoadMore,
  onDeploysLoaded,
  scrollRef,
  ...props
}: IInstallComponentDeploySwitcher) => {
  const [searchTerm, setSearchTerm] = useState('')

  const filteredDeploys = deploys
    ? deploys.filter(
        (deploy) =>
          deploy.id.includes(searchTerm) ||
          deploy?.created_by?.email.includes(searchTerm) ||
          deploy?.status_v2?.status.includes(searchTerm)
      )
    : []

  return (
    <Dropdown
      id="deploy-switcher"
      alignment={alignment}
      buttonText="Latest deploys"
      {...props}
    >
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
          ref={scrollRef}
          className="flex flex-col gap-2 p-2 max-h-56 overflow-y-auto"
        >
          {filteredDeploys?.length ? (
            <>
              {filteredDeploys?.map((deploy, idx) => (
                <span key={deploy.id} className="rounded-lg border">
                  <Button
                    className={cn('!p-2 !h-fit w-full', {
                      '!bg-primary-600/5 dark:!bg-primary-600/5':
                        deploy?.id === deployId,
                    })}
                    href={`/${orgId}/installs/${installId}/components/${componentId}/deploys/${deploy?.id}`}
                    variant="ghost"
                  >
                    <DeploySummary deploy={deploy} isLatest={idx === 0} />
                  </Button>
                </span>
              ))}
              {isLoading ? <DeploysSkeleton limit={8} /> : null}
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
    </Dropdown>
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
