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
import type { TSandboxRun } from '@/types'
import { cn } from '@/utils/classnames'
import { SandboxRunSummary } from './SandboxRunSummary'
import { SandboxRunsSkeleton } from './SandboxRunsSkeleton'

interface ISandboxRunMenu extends Omit<IMenu, 'children'> {
  activeSandboxRunId: string
}

export const SandboxRunMenu = ({ activeSandboxRunId }: ISandboxRunMenu) => {
  const limit = 8
  const [searchTerm, setSearchTerm] = useState('')
  const [offset, setOffset] = useState(0)
  const [sandboxRuns, setSandboxRuns] = useState<TSandboxRun[]>([])
  const { install } = useInstall()
  const { org } = useOrg()
  const queryParams = useQueryParams({
    limit,
    offset,
  })
  const { data, error, headers, isLoading } = useQuery<TSandboxRun[]>({
    path: `/api/orgs/${org.id}/installs/${install.id}/sandbox/runs${queryParams}`,
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
    setSandboxRuns((prev) => {
      const sandboxRunMap = new Map(
        prev.map((sandboxRun) => [sandboxRun.id, sandboxRun])
      )
      data.forEach((sandboxRun) => sandboxRunMap.set(sandboxRun.id, sandboxRun))
      return Array.from(sandboxRunMap.values())
    })
    scrollToBottom.reset()
  }, [data])

  const filteredSandboxRuns = sandboxRuns
    ? sandboxRuns.filter(
        (sandboxRun) =>
          sandboxRun.id.includes(searchTerm) ||
          sandboxRun?.created_by?.email.includes(searchTerm) ||
          sandboxRun?.status_v2?.status.includes(searchTerm)
      )
    : []

  return (
    <Menu className="w-100 !p-0">
      <div className="flex flex-col gap-2 p-2 border-b">
        <Text variant="label" theme="neutral">
          Latest sandbox runs
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
        {filteredSandboxRuns?.length && !error ? (
          <>
            {filteredSandboxRuns?.map((sandboxRun, idx) => (
              <span key={sandboxRun.id} className="rounded-lg border">
                <Button
                  className={cn('!p-2 !h-fit w-full', {
                    '!bg-primary-600/5 dark:!bg-primary-600/5':
                      sandboxRun?.id === activeSandboxRunId,
                  })}
                  href={`/${org.id}/installs/${install.id}/sandbox/${sandboxRun?.id}`}
                  variant="ghost"
                >
                  <SandboxRunSummary
                    sandboxRun={sandboxRun}
                    isLatest={idx === 0}
                  />
                </Button>
              </span>
            ))}
            {isLoading ? <SandboxRunsSkeleton limit={limit} /> : null}
          </>
        ) : isLoading ? (
          <SandboxRunsSkeleton limit={limit} />
        ) : (
          <EmptyState
            variant="history"
            emptyMessage="Unable to find any sandbox runs."
            emptyTitle="No sandbox runs found"
            size="sm"
          />
        )}
      </div>
    </Menu>
  )
}
