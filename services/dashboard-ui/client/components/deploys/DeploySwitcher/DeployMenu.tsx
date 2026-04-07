import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Menu, type IMenu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import type { TDeploy } from '@/types'
import { cn } from '@/utils/classnames'
import { DeploySummary } from './DeploySummary'
import { DeploysSkeleton } from './DeploysSkeleton'

interface IDeployMenu extends Omit<IMenu, 'children'> {
  activeDeployId: string
  deploys: TDeploy[]
  isLoading: boolean
  hasError: boolean
  orgId: string
  installId: string
  componentId: string
  scrollRef: React.RefObject<HTMLDivElement | null>
  limit: number
}

export const DeployMenu = ({
  activeDeployId,
  deploys,
  isLoading,
  hasError,
  orgId,
  installId,
  componentId,
  scrollRef,
  limit,
}: IDeployMenu) => {
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
        {filteredDeploys?.length && !hasError ? (
          <>
            {filteredDeploys?.map((deploy, idx) => (
              <span key={deploy.id} className="rounded-lg border">
                <Button
                  className={cn('!p-2 !h-fit w-full', {
                    '!bg-primary-600/5 dark:!bg-primary-600/5':
                      deploy?.id === activeDeployId,
                  })}
                  href={`/${orgId}/installs/${installId}/components/${componentId}/deploys/${deploy?.id}`}
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
