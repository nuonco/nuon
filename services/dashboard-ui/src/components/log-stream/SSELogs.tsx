'use client'

import { useSearchParams } from 'next/navigation'
import { useEffect, useMemo, memo } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useLogViewer, useUnifiedLogData } from '@/hooks/use-logs-temp'
import type { TOTELLog } from '@/types'
import { cn } from '@/utils/classnames'
import { LogSeverity } from './LogSeverity'
import { LogLineSkeleton } from './LogLine'
import { LogFilters } from './log-filters/LogFilters'

export const LogsSkeleton = () => {
  return Array.from({ length: 20 }).map((_, idx) => (
    <LogLineSkeleton key={`log-line-${idx}`} />
  ))
}

// demo sse logs
export const SSELogs = ({
  filterClassName = '-top-2',
}: {
  filterClassName?: string
}) => {
  const { loadMore, hasMore, isLoading, isStreamOpen, connectionState } =
    useUnifiedLogData()
  const { filteredLogs, filters, activeLog, handleActiveLog } = useLogViewer()

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col flex-auto">
        <div
          className={cn('sticky bg-background border-b z-10', filterClassName)}
        >
          <LogFilters filters={filters} />
          <div className="grid grid-cols-[3rem_15rem_3rem_1fr] gap-6 py-2">
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
        </div>

        <div className="flex flex-col divide-y">
          {!isStreamOpen && !filteredLogs?.length && isLoading ? (
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
  )
}

interface ILogLine {
  log: TOTELLog
  activeLogId?: string
  onActivate: (logId: string) => void
}

const LogLineComponent = ({ log, activeLogId, onActivate }: ILogLine) => {
  const searchParams = useSearchParams()

  useEffect(() => {
    if (log.id && log.id === searchParams?.get('panel')) {
      onActivate(log.id)
    }
  }, [])

  const isActive = activeLogId === log.id

  return (
    <div>
      <Button
        className={cn(
          '!grid grid-cols-[3rem_15rem_3rem_1fr] gap-6 !py-1 !px-0 text-left w-full rounded-none h-fit',
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

        <Text family="mono" variant="subtext">
          {log.service_name}
        </Text>
        <span className="!inline-block w-full max-w-full overflow-hidden">
          <Text
            className="!block !text-nowrap truncate"
            family="mono"
            variant="subtext"
          >
            {log.body}
          </Text>
        </span>
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
