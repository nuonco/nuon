'use client'

import { useSearchParams } from 'next/navigation'
import { useEffect, useState } from 'react'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { Time } from '@/components/common/Time'
import { LogSeverity } from './LogSeverity'
import type { TOTELLog } from '@/types'
import { cn } from '@/utils/classnames'
import { LogLineSkeleton } from './LogLine'
import { LogFilters } from './log-filters/LogFilters'
import { useLogViewer, useUnifiedLogData } from '@/hooks/use-logs-temp'

export const LogsSkeleton = () => {
  return Array.from({ length: 20 }).map((_, idx) => (
    <LogLineSkeleton key={`log-line-${idx}`} />
  ))
}

// demo sse logs
export const SSELogs = () => {
  const { logs, loadMore, hasMore, isLoading, isStreamOpen, connectionState } =
    useUnifiedLogData()
  const { filteredLogs, filters } = useLogViewer()
  const [animatingLogs, setAnimatingLogs] = useState<Set<string>>(new Set())
  const [logTimestamps, setLogTimestamps] = useState<Map<string, number>>(
    new Map()
  )

  useEffect(() => {
    if (!filteredLogs) return

    // Find new filteredLogs that don't have timestamps yet
    const newLogs = filteredLogs.filter((log) => !logTimestamps.has(log.id))

    if (newLogs.length > 0) {
      const now = Date.now()
      const newTimestamps = new Map(logTimestamps)

      // Assign timestamps to new filteredLogs with staggered delays
      newLogs.forEach((log, index) => {
        newTimestamps.set(log.id, now + index * 50) // 50ms between each log

        // Trigger animation after the delay
        setTimeout(() => {
          setAnimatingLogs((prev) => new Set(prev).add(log.id))
        }, index * 50)
      })

      setLogTimestamps(newTimestamps)
    }
  }, [filteredLogs, logTimestamps])

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-8 font-mono text-xs text-cool-grey-500">
        <span>SSE connection: {connectionState}</span>
        <span>log lines: {logs?.length || 0}</span>
      </div>

      <div className="flex flex-col flex-auto">
        <div className="sticky bg-background border-b z-10 -top-2">
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
          {filteredLogs?.slice().map((logLine) => (
            <TransitionDiv
              key={logLine?.id}
              isVisible={animatingLogs.has(logLine.id)}
              className={cn({
                'slide-in': isStreamOpen,
                fade: !isStreamOpen,
              })}
            >
              <LogLine log={logLine} />
            </TransitionDiv>
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
}

export const LogLine = ({ log }: ILogLine) => {
  const searchParams = useSearchParams()
  const { activeLog, handleActiveLog } = useLogViewer()

  useEffect(() => {
    if (log.id && log.id === searchParams?.get('panel')) {
      handleActiveLog(log.id)
    }
  }, [])

  return (
    <div>
      <Button
        className={cn(
          '!grid grid-cols-[3rem_15rem_3rem_1fr] gap-6 !py-1 !px-0 text-left w-full rounded-none h-fit',
          'hover:!bg-black/10 dark:hover:!bg-white/10 focus:!bg-black/10 dark:focus:!bg-white/10',
          {
            '!bg-cool-grey-100 dark:!bg-dark-grey-800':
              log.service_name === 'runner',
            '!bg-primary-600/40 dark:!bg-primary-600/30':
              activeLog?.id === log?.id,
          }
        )}
        onClick={() => {
          handleActiveLog(log.id)
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
