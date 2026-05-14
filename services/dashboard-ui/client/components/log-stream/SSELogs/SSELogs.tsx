import { useEffect, useRef, memo } from 'react'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TOTELLog } from '@/types'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'
import { cn } from '@/utils/classnames'
import { LogSeverity } from '../LogSeverity'
import { LogLineSkeleton } from '../LogLine'
import { Skeleton } from '@/components/common/Skeleton'
import { LogFilters } from '../log-filters/LogFilters'

export const LogsSkeleton = () => {
  return Array.from({ length: 20 }).map((_, idx) => (
    <LogLineSkeleton key={`log-line-${idx}`} />
  ))
}

export const LogsPageSkeleton = () => {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col flex-auto">
        <div className="border-b">
          <div className="flex items-center justify-between gap-4 py-4 w-full">
            <Skeleton height="36px" width="240px" />
            <div className="flex items-center gap-2">
              <Skeleton height="36px" width="100px" />
              <Skeleton height="36px" width="100px" />
            </div>
          </div>
          <div className="grid grid-cols-[3rem_15rem_8rem_1fr] gap-6 py-2">
            <Skeleton height="14px" width="3rem" />
            <Skeleton height="14px" width="4rem" />
            <Skeleton height="14px" width="3rem" />
            <Skeleton height="14px" width="4rem" />
          </div>
        </div>
        <div className="flex flex-col divide-y">
          <LogsSkeleton />
        </div>
      </div>
    </div>
  )
}

interface ISSELogs {
  filterClassName?: string
  filteredLogs: TOTELLog[] | undefined
  filters: TLogFiltersProps
  activeLog: TOTELLog | undefined
  handleActiveLog: (logId: string) => void
  loadMore: () => void
  hasMore: boolean
  isLoading: boolean
  isStreamOpen: boolean
  deepLinkLogId?: string | null
}

export const SSELogs = ({
  filterClassName = 'top-0',
  filteredLogs,
  filters,
  activeLog,
  handleActiveLog,
  loadMore,
  hasMore,
  isLoading,
  isStreamOpen,
  deepLinkLogId,
}: ISSELogs) => {
  const deepLinkHandledRef = useRef(false)

  useEffect(() => {
    if (deepLinkHandledRef.current || !filteredLogs?.length) return
    if (!deepLinkLogId) return

    const idx = filteredLogs.findIndex(l => l.id === deepLinkLogId)
    if (idx === -1) return

    deepLinkHandledRef.current = true
    handleActiveLog(deepLinkLogId)
    const el = document.getElementById(`log-${deepLinkLogId}`)
    el?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, [filteredLogs, deepLinkLogId, handleActiveLog])

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col flex-auto">
        <div
          className={cn('sticky bg-background border-b z-10', filterClassName)}
        >
          <div className="flex items-center gap-4">
            <LogFilters filters={filters} />
          </div>
        </div>

        <div className="overflow-x-auto">
          <div className="min-w-full w-max">
            <div className="grid grid-cols-[3rem_15rem_8rem_auto] gap-6 py-2 border-b bg-background">
              <Text variant="subtext" weight="strong" theme="neutral">
                Severity
              </Text>
              <Text variant="subtext" weight="strong" theme="neutral">
                Datetime
              </Text>
              <Text variant="subtext" weight="strong" theme="neutral">
                Service
              </Text>
              <Text variant="subtext" weight="strong" theme="neutral">
                Content
              </Text>
            </div>

            {!filteredLogs?.length && filters.filterStats.totalCount > 0 ? (
              <EmptyState
                variant="search"
                emptyTitle="Filters are hiding logs"
                emptyMessage={`${filters.filterStats.totalCount} log(s) are hidden by the current filters.`}
                action={
                  <Button size="sm" onClick={filters.handleResetAll}>
                    Reset filters
                  </Button>
                }
              />
            ) : null}

            {!filteredLogs?.length && !filters.filterStats.totalCount && !isLoading && isStreamOpen ? (
              <EmptyState
                variant="table"
                emptyTitle="Waiting for logs"
                emptyMessage="Logs will appear here once they start coming in."
              />
            ) : null}

            {!filteredLogs?.length && !filters.filterStats.totalCount && !isLoading && !isStreamOpen ? (
              <EmptyState
                variant="table"
                emptyTitle="No logs available"
                emptyMessage=""
              />
            ) : null}

            <div className="flex flex-col divide-y">
              {!filteredLogs?.length && isLoading ? (
                <LogsSkeleton />
              ) : null}

              {filteredLogs?.map((logLine) => (
                <LogLine
                  key={logLine?.id}
                  log={logLine}
                  activeLogId={activeLog?.id}
                  onActivate={handleActiveLog}
                />
              ))}

              {!isStreamOpen && hasMore ? (
                <Button
                  onClick={loadMore}
                  disabled={isLoading}
                  variant="ghost"
                  className="mx-auto mt-4"
                >
                  {isLoading ? (
                    <>
                      <Icon variant="Loading" /> Loading
                    </>
                  ) : (
                    <>Load more</>
                  )}
                </Button>
              ) : null}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

interface ILogLine {
  log: TOTELLog
  activeLogId?: string
  onActivate: (logId: string) => void
}

const LogLineComponent = ({ log, activeLogId, onActivate }: ILogLine) => {
  const isActive = activeLogId === log.id

  return (
    <div id={`log-${log.id}`} className="border-b">
      <Button
        className={cn(
          '!grid grid-cols-[3rem_15rem_8rem_auto] gap-6 !py-1 !px-0 text-left w-full rounded-none h-fit',
          'hover:!bg-black/10 dark:hover:!bg-white/10 focus:!bg-black/10 dark:focus:!bg-white/10',
          {
            '!bg-cool-grey-100 dark:!bg-dark-grey-800':
              log.service_name === 'runner',
            '!bg-primary-600/40 dark:!bg-primary-600/30': isActive,
          }
        )}
        onClick={() => {
          onActivate(log.id)
        }}
        variant="ghost"
      >
        <LogSeverity
          severityNumber={log.severity_number}
          severityText={log.severity_text}
          variant="subtext"
        />
        <Time
          className=""
          time={log.timestamp}
          format="log-datetime"
          family="mono"
          variant="subtext"
        />

        <Text
          family="mono"
          variant="subtext"
          nowrap
          as="div"
          className="truncate"
          title={log.service_name}
        >
          {log.service_name?.split('.').pop()}
        </Text>
        <Text
          nowrap
          as="div"
          family="mono"
          variant="subtext"
        >
          {log.body}
        </Text>
      </Button>
    </div>
  )
}

export const LogLine = memo(LogLineComponent, (prev, next) => {
  return (
    prev.log.id === next.log.id &&
    prev.activeLogId === next.activeLogId &&
    prev.log.body === next.log.body &&
    prev.log.timestamp === next.log.timestamp
  )
})

LogLine.displayName = 'LogLine'
