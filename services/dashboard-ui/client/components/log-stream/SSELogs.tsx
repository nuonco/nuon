import { useSearchParams } from 'react-router'
import { useEffect, useRef, memo } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useLogViewer, useUnifiedLogData } from '@/hooks/use-logs'
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

export const SSELogs = ({
  filterClassName = '-top-6',
}: {
  filterClassName?: string
}) => {
  const { loadMore, hasMore, isLoading, isStreamOpen } = useUnifiedLogData()
  const { filteredLogs, filters, activeLog, handleActiveLog } = useLogViewer()
  const [searchParams] = useSearchParams()
  const deepLinkHandledRef = useRef(false)

  useEffect(() => {
    if (deepLinkHandledRef.current || !filteredLogs?.length) return
    const targetLogId = searchParams?.get('log')
    if (!targetLogId) return

    const idx = filteredLogs.findIndex(l => l.id === targetLogId)
    if (idx === -1) return

    deepLinkHandledRef.current = true
    handleActiveLog(targetLogId)
    const el = document.getElementById(`log-${targetLogId}`)
    el?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, [filteredLogs, searchParams, handleActiveLog])

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col flex-auto">
        <div
          className={cn('sticky bg-background border-b z-10', filterClassName)}
        >
          <div className="flex items-center gap-4">
            <LogFilters filters={filters} />
          </div>
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
            nowrap
            as="div"
            className="truncate"
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
