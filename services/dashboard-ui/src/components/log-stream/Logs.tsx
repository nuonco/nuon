'use client'

import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { useLogs } from '@/hooks/use-logs'
import { LogLine, LogLineSkeleton } from './LogLine'
import { LogFilters } from './log-filters/LogFilters'

interface ILogs {}

export const Logs = ({}: ILogs) => {
  const { error, isLoading, logs } = useLogs()

  return (
    <div className="flex flex-col flex-auto">
      <div className="sticky -top-2 bg-background border-b">
        <LogFilters />
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

      {logs.length ? (
        <div className="flex flex-col divide-y">
          {isLoading ? (
            <TransitionDiv className="fade" isVisible={isLoading}>
              <LogLineSkeleton />
            </TransitionDiv>
          ) : null}
          {logs.map((log) => (
            <LogLine key={log.id} log={log} />
          ))}
        </div>
      ) : isLoading ? (
        <LogsSkeleton />
      ) : (
        <EmptyState
          className="!my-8"
          variant="table"
          emptyMessage="There are no logs to display. This could be because no logs have been created yet, or your current filters do not match any results."
          emptyTitle="No logs found"
        />
      )}
    </div>
  )
}

export const LogsSkeleton = () => {
  return Array.from({ length: 20 }).map((_, idx) => (
    <LogLineSkeleton key={`log-line-${idx}`} />
  ))
}
