import React, { useEffect, useMemo, useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { LogSeverity } from '@/components/log-stream/LogSeverity'
import { LogFilters } from '@/components/log-stream/log-filters/LogFilters'
import { LogLineSkeleton } from '@/components/log-stream/LogLine'
import type { TOTELLog, TActionConfig } from '@/types'
import type { TLogFiltersProps } from '@/hooks/use-log-filters'
import { cn } from '@/utils/classnames'

interface IInstallActionRunLogs {
  actionConfig: TActionConfig
  layout?: 'vertical' | 'horizontal'
  allLogs?: TOTELLog[]
  filteredLogs: TOTELLog[]
  loadMore: () => void
  hasMore: boolean
  isLoading: boolean
  isStreamOpen: boolean
  activeLog: TOTELLog | undefined
  handleActiveLog: (id: string) => void
  filters: TLogFiltersProps
  searchParamPanel?: string | null
  stepStatuses?: Record<string, string>
}

export const InstallActionRunLogs = ({
  actionConfig,
  layout = 'vertical',
  allLogs,
  filteredLogs,
  loadMore,
  hasMore,
  isLoading,
  isStreamOpen,
  activeLog,
  handleActiveLog,
  filters,
  searchParamPanel,
  stepStatuses,
}: IInstallActionRunLogs) => {
  const steps = actionConfig?.steps || []

  const allLogStepCounts = useMemo(() => {
    const source = allLogs ?? filteredLogs
    if (!source) return {}
    const counts: Record<string, number> = {}
    for (const log of source) {
      const stepName = log.log_attributes?.workflow_step_name
      if (stepName) counts[stepName] = (counts[stepName] ?? 0) + 1
    }
    return counts
  }, [allLogs, filteredLogs])

  const logSteps = useMemo(() => {
    if (!filteredLogs) return {}

    return filteredLogs.reduce(
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
  }, [filteredLogs])

  const stepKeys = useMemo(() => Object.keys(logSteps), [logSteps])
  const [activeStep, setActiveStep] = useState<string | undefined>(
    stepKeys?.[0]
  )
  const [showAllLogs, setShowAllLogs] = useState<boolean>(
    !activeStep ? true : false
  )

  useEffect(() => {
    if (showAllLogs) return
    if (!activeStep && stepKeys.length) {
      setActiveStep(stepKeys[0])
    }
  }, [stepKeys, activeStep, showAllLogs])

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
          <span className="flex items-center gap-2">
            {stepStatuses?.[step?.name] && (
              <Status status={stepStatuses[step.name]} isWithoutText />
            )}
            <span className="truncate">{step?.name}</span>
            {allLogStepCounts[step?.name] > 0 && (
              <Badge size="sm">{allLogStepCounts[step.name]}</Badge>
            )}
          </span>
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

  const logViewerProps = {
    activeStep,
    showAllLogs,
    logSteps,
    allLogStepCounts,
    loadMore,
    hasMore,
    isLoading,
    isStreamOpen,
    filteredLogs,
    activeLog,
    handleActiveLog,
    filters,
    searchParamPanel,
  }

  if (layout === 'horizontal') {
    return (
      <div className="flex flex-col gap-4 flex-auto">
        <div className="flex flex-wrap gap-2">
          {stepButtons}
        </div>
        <div className="w-full">
          <StepAwareLogViewer {...logViewerProps} />
        </div>
      </div>
    )
  }

  return (
    <div className="flex items-start flex-auto">
      <div className="flex flex-col gap-2 w-fit md:min-w-64 pr-2 h-full">
        {stepButtons}
      </div>
      <div className="pl-2 w-full border-l">
        <StepAwareLogViewer {...logViewerProps} />
      </div>
    </div>
  )
}

const LogsSkeleton = () => {
  return Array.from({ length: 20 }).map((_, idx) => (
    <LogLineSkeleton key={`log-line-${idx}`} />
  ))
}

const StepAwareLogViewer = ({
  activeStep,
  showAllLogs,
  logSteps,
  allLogStepCounts,
  loadMore,
  hasMore,
  isLoading,
  isStreamOpen,
  filteredLogs,
  activeLog,
  handleActiveLog,
  filters,
  searchParamPanel,
}: {
  activeStep?: string
  showAllLogs: boolean
  logSteps: Record<string, TOTELLog[]>
  allLogStepCounts: Record<string, number>
  loadMore: () => void
  hasMore: boolean
  isLoading: boolean
  isStreamOpen: boolean
  filteredLogs: TOTELLog[]
  activeLog: TOTELLog | undefined
  handleActiveLog: (id: string) => void
  filters: TLogFiltersProps
  searchParamPanel?: string | null
}) => {
  const displayLogs = useMemo(() => {
    if (showAllLogs) {
      return filteredLogs
    } else if (activeStep && logSteps[activeStep]) {
      return logSteps[activeStep]
    }
    return []
  }, [showAllLogs, activeStep, logSteps, filteredLogs])

  const scopedFilters = useMemo(() => {
    if (showAllLogs) return filters
    const selectedCount = displayLogs?.length ?? 0
    const totalCount = activeStep ? allLogStepCounts[activeStep] ?? selectedCount : selectedCount
    return {
      ...filters,
      filterStats: { selectedCount, totalCount },
    }
  }, [filters, displayLogs?.length, showAllLogs, activeStep, allLogStepCounts])

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col flex-auto">
        <div className="sticky bg-background border-b z-10 -top-6">
          <LogFilters filters={scopedFilters} />
          <div className="grid grid-cols-[3rem_15rem_8rem_1fr] gap-6 py-2">
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
          {!isStreamOpen && !displayLogs?.length && isLoading ? (
            <LogsSkeleton />
          ) : null}

          {displayLogs?.map((logLine) => (
            <LogLine
              key={logLine?.id}
              log={logLine}
              activeLog={activeLog}
              handleActiveLog={handleActiveLog}
              searchParamPanel={searchParamPanel}
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

const LogLine = ({
  log,
  activeLog,
  handleActiveLog,
  searchParamPanel,
}: {
  log: TOTELLog
  activeLog: TOTELLog | undefined
  handleActiveLog: (id: string) => void
  searchParamPanel?: string | null
}) => {
  useEffect(() => {
    if (log.id && log.id === searchParamPanel) {
      handleActiveLog(log.id)
    }
  }, [])

  return (
    <div>
      <Button
        className={cn(
          '!grid grid-cols-[3rem_15rem_8rem_1fr] gap-6 !py-1 !px-0 text-left w-full rounded-none h-fit',
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
