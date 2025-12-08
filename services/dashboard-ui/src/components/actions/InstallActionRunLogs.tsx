'use client'

import React, { useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { TransitionDiv } from '@/components/common/TransitionDiv'
import { Time } from '@/components/common/Time'
import { LogSeverity } from '@/components/log-stream/LogSeverity'
import { LogFilters } from '@/components/log-stream/log-filters/LogFilters'
import { LogLineSkeleton } from '@/components/log-stream/LogLine'
import { useUnifiedLogData, useLogViewer } from '@/hooks/use-logs-temp'
import type { TOTELLog, TActionConfig } from '@/types'
import { cn } from '@/utils/classnames'

export const InstallActionRunLogs = ({
  actionConfig,
  layout = 'vertical',
}: {
  actionConfig: TActionConfig
  layout?: 'vertical' | 'horizontal'
}) => {
  const { logs, isLoading } = useUnifiedLogData()

  const steps = actionConfig?.steps || []

  const logSteps = useMemo(() => {
    return (logs as unknown as TOTELLog[]).reduce(
      (acc, log) => {
        const stepName = log.log_attributes?.workflow_step_name
        if (stepName) {
          if (!acc[stepName]) acc[stepName] = []
          acc[stepName].push(log)
        }
        return acc
      },
      {} as Record<string, TOTELLog[]>
    )
  }, [logs])

  const stepKeys = useMemo(() => Object.keys(logSteps), [logSteps])
  const [activeStep, setActiveStep] = useState<string | undefined>(
    stepKeys?.[0]
  )
  const [showAllLogs, setShowAllLogs] = useState<boolean>(
    !activeStep ? true : false
  )

  useEffect(() => {
    if (showAllLogs) return
    if (!stepKeys.length) {
      setActiveStep(undefined)
      return
    }
    if (!activeStep) {
      setActiveStep(stepKeys[0])
      return
    }
    if (!stepKeys.includes(activeStep)) {
      setActiveStep(stepKeys[0])
    }
  }, [stepKeys, activeStep, showAllLogs])

  // Get the logs to display based on current selection
  const displayLogs = useMemo(() => {
    if (showAllLogs) {
      return logs
    } else if (activeStep && logSteps[activeStep]) {
      return logSteps[activeStep]
    }
    return []
  }, [showAllLogs, activeStep, logSteps, logs])

  const sortedSteps = steps.sort((a, b) => {
    if (a.idx === undefined && b.idx === undefined) return 0
    if (a.idx === undefined) return -1
    if (b.idx === undefined) return 1
    return a.idx - b.idx
  })

  const stepButtons = (
    <>
      {sortedSteps.map((step) => (
        <Button
          className={cn(layout === 'horizontal' ? 'w-auto' : 'w-full', {
            '!bg-primary-600/10 dark:!bg-primary-400/10':
              activeStep === step?.name && !showAllLogs,
          })}
          variant="ghost"
          key={step?.id}
          disabled={!stepKeys.includes(step?.name)}
          onClick={() => {
            if (showAllLogs) setShowAllLogs(false)
            setActiveStep(step?.name)
          }}
        >
          <span className="truncate">{step?.name}</span>
        </Button>
      ))}
      <Button
        className={cn(layout === 'horizontal' ? 'w-auto' : 'w-full', {
          '!bg-primary-600/10 dark:!bg-primary-400/10': showAllLogs,
        })}
        onClick={() => {
          setShowAllLogs(true)
        }}
        variant="ghost"
      >
        View all logs
      </Button>
    </>
  )

  if (layout === 'horizontal') {
    return (
      <div className="flex flex-col gap-4 flex-auto">
        <div className="flex flex-wrap gap-2">
          {stepButtons}
        </div>
        <div className="w-full">
          <StepAwareLogViewer logs={displayLogs} />
        </div>
      </div>
    )
  }

  // Default vertical layout
  return (
    <div className="flex items-start flex-auto">
      <div className="flex flex-col gap-2 w-fit md:min-w-64 pr-2 h-full">
        {stepButtons}
      </div>
      <div className="pl-2 w-full border-l">
        <StepAwareLogViewer logs={displayLogs} />
      </div>
    </div>
  )
}

// Skeleton component for loading states (matching SSELogs pattern)
const LogsSkeleton = () => {
  return Array.from({ length: 20 }).map((_, idx) => (
    <LogLineSkeleton key={`log-line-${idx}`} />
  ))
}

// Custom log viewer component that displays filtered logs with SSE-style animations
const StepAwareLogViewer = ({ logs }: { logs: TOTELLog[] }) => {
  const { loadMore, hasMore, isLoading, isStreamOpen } = useUnifiedLogData()
  const { activeLog, handleActiveLog, filters } = useLogViewer()
  const [animatingLogs, setAnimatingLogs] = useState<Set<string>>(new Set())
  const [logTimestamps, setLogTimestamps] = useState<Map<string, number>>(new Map())

  // Animate new logs when they appear
  useEffect(() => {
    if (!logs) return

    const newLogs = logs.filter((log) => !logTimestamps.has(log.id))

    if (newLogs.length > 0) {
      const now = Date.now()
      const newTimestamps = new Map(logTimestamps)

      if (isStreamOpen) {
        // Only stagger animations when SSE is enabled
        newLogs.forEach((log, index) => {
          newTimestamps.set(log.id, now + index * 50)
          setTimeout(() => {
            setAnimatingLogs((prev) => new Set(prev).add(log.id))
          }, index * 50)
        })
      } else {
        // Show all logs immediately without staggering
        newLogs.forEach((log) => {
          newTimestamps.set(log.id, now)
          setAnimatingLogs((prev) => new Set(prev).add(log.id))
        })
      }

      setLogTimestamps(newTimestamps)
    }
  }, [logs, logTimestamps, isStreamOpen])

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col flex-auto">
        <div className="sticky bg-background border-b z-10 -top-6">
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
          {!isStreamOpen && !logs?.length && isLoading ? (
            <LogsSkeleton />
          ) : null}

          {logs?.map((logLine) => (
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

// Log line component (copied from SSELogs)
const LogLine = ({ log }: { log: TOTELLog }) => {
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
