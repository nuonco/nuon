import { useEffect, useState } from 'react'
import { DateTime } from 'luxon'
import { Duration } from '@/components/common/Duration'
import { Text, type IText } from '@/components/common/Text'
import type { TInstallActionRun } from '@/types'

const ACTIVE_RUN_STATUSES = ['in-progress', 'queued']
const QUEUED_JOB_STATUSES = ['queued', 'available']
const WARN_RATIO = 0.8

export const getRunPhase = (
  run?: TInstallActionRun
): 'queued' | 'running' | 'done' => {
  const runStatus = run?.status_v2?.status ?? run?.status
  if (!runStatus || !ACTIVE_RUN_STATUSES.includes(runStatus)) return 'done'

  const jobStatus = run?.runner_job?.status
  if (!jobStatus || QUEUED_JOB_STATUSES.includes(jobStatus)) return 'queued'
  if (jobStatus === 'in-progress') return 'running'
  return 'done'
}

const validTime = (time?: string) => {
  if (!time) return undefined
  const dt = DateTime.fromISO(time)
  return dt.isValid && dt.year > 1 ? dt : undefined
}

export interface IRunTiming extends Pick<IText, 'variant' | 'theme'> {
  run?: TInstallActionRun
  hideQueuedLabel?: boolean
}

export const RunTiming = ({
  run,
  hideQueuedLabel = false,
  ...textProps
}: IRunTiming) => {
  const phase = getRunPhase(run)
  const isRunning = phase !== 'done'
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    if (!isRunning) return
    setNow(Date.now())
    const id = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(id)
  }, [isRunning])

  if (!run) return null

  if (phase === 'queued') {
    const start =
      validTime(run.runner_job?.created_at) ?? validTime(run.created_at)
    const elapsedMs = start ? Math.max(0, now - start.toMillis()) : 0

    return (
      <Text {...textProps} className="inline-flex items-center gap-1">
        {hideQueuedLabel ? null : <span>queued</span>}
        <Duration {...textProps} nanoseconds={elapsedMs * 1e6} />
      </Text>
    )
  }

  if (phase === 'running') {
    const start =
      validTime(run.runner_job?.started_at) ?? validTime(run.created_at)
    const elapsedMs = start ? Math.max(0, now - start.toMillis()) : 0
    const timeout = run.effective_timeout
    const theme =
      timeout && elapsedMs * 1e6 >= timeout * WARN_RATIO
        ? 'warn'
        : textProps.theme
    const props = { ...textProps, theme }

    return (
      <Text {...props} className="inline-flex items-center gap-1">
        <Duration {...props} nanoseconds={elapsedMs * 1e6} />
        {timeout ? (
          <>
            <span>/</span>
            <Duration {...props} nanoseconds={timeout} />
          </>
        ) : null}
      </Text>
    )
  }

  if (run.execution_time) {
    return <Duration {...textProps} nanoseconds={run.execution_time} />
  }

  return (
    <Duration
      {...textProps}
      beginTime={run.created_at}
      endTime={run.updated_at}
    />
  )
}
