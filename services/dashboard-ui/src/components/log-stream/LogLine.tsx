'use client'

import { useSearchParams } from 'next/navigation'
import { useEffect } from 'react'
import { Button } from '@/components/common/Button'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useLogs } from '@/hooks/use-logs'
import type { TOTELLog } from '@/types'
import { cn } from '@/utils/classnames'
import { LogSeverity } from './LogSeverity'

interface ILogLine {
  log: TOTELLog
}

export const LogLine = ({ log }: ILogLine) => {
  const searchParams = useSearchParams()
  const { activeLog, handleActiveLog } = useLogs()

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

export const LogLineSkeleton = () => {
  return (
    <div className="grid grid-cols-[3rem_15rem_3rem_1fr] gap-6 py-2">
      <Skeleton height="17px" width="40px" />
      <Skeleton height="17px" width="240px" />
      <Skeleton height="17px" width="50px" />
      <Skeleton height="17px" width="100%" />
    </div>
  )
}
