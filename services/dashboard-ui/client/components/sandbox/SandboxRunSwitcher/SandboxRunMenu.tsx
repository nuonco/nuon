import { useState } from 'react'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Menu, type IMenu } from '@/components/common/Menu'
import { SearchInput } from '@/components/common/SearchInput'
import { Text } from '@/components/common/Text'
import type { TSandboxRun } from '@/types'
import { cn } from '@/utils/classnames'
import { SandboxRunSummary } from './SandboxRunSummary'
import { SandboxRunsSkeleton } from './SandboxRunsSkeleton'

interface ISandboxRunMenu extends Omit<IMenu, 'children'> {
  activeSandboxRunId: string
  sandboxRuns: TSandboxRun[]
  isLoading: boolean
  hasError: boolean
  orgId: string
  installId: string
  scrollRef: React.RefObject<HTMLDivElement | null>
  limit: number
}

export const SandboxRunMenu = ({
  activeSandboxRunId,
  sandboxRuns,
  isLoading,
  hasError,
  orgId,
  installId,
  scrollRef,
  limit,
}: ISandboxRunMenu) => {
  const [searchTerm, setSearchTerm] = useState('')

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
        ref={scrollRef}
        className="flex flex-col gap-2 p-2 max-h-56 overflow-y-auto"
      >
        {filteredSandboxRuns?.length && !hasError ? (
          <>
            {filteredSandboxRuns?.map((sandboxRun, idx) => (
              <span key={sandboxRun.id} className="rounded-lg border">
                <Button
                  className={cn('!p-2 !h-fit w-full', {
                    '!bg-primary-600/5 dark:!bg-primary-600/5':
                      sandboxRun?.id === activeSandboxRunId,
                  })}
                  href={`/${orgId}/installs/${installId}/sandbox/${sandboxRun?.id}`}
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
