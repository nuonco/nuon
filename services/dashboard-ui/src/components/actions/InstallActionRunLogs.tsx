'use client'

import React, { useEffect, useMemo, useState } from 'react'
import { Button } from '@/components/common/Button'
import { LogsViewer } from '@/components/log-stream/Logs'
import { useLogs } from '@/hooks/use-logs'
import type { TOTELLog } from '@/types'
import { cn } from '@/utils/classnames'

export const InstallActionRunLogs = () => {
  const { isLoading, logs } = useLogs()

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

  return (
    <div className="flex items-start flex-auto divide-x">
      <div className="flex flex-col gap-2 w-fit md:min-w-64 pr-2 h-full">
        {stepKeys.map((key) => (
          <Button
            className={cn('w-full', {
              '!bg-primary-600/10 dark:!bg-primary-400/10':
                activeStep === key && !showAllLogs,
            })}
            variant="ghost"
            key={key}
            onClick={() => {
              if (showAllLogs) setShowAllLogs(false)
              setActiveStep(key)
            }}
          >
            {key}
          </Button>
        ))}
        <Button
          className={cn('w-full', {
            '!bg-primary-600/10 dark:!bg-primary-400/10': showAllLogs,
          })}
          onClick={() => {
            setShowAllLogs(true)
          }}
          variant="ghost"
        >
          View all logs
        </Button>
      </div>
      <div className="pl-2 w-full">
        <LogsViewer
          stratusPage
          isLoading={isLoading}
          logs={showAllLogs ? logs : logSteps[activeStep]}
        />
      </div>
    </div>
  )
}
