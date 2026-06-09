import { useEffect, useMemo, useRef, useState } from 'react'
import { Button } from '@/components/common/Button'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import type { TOTELLog } from '@/types'
import { cn } from '@/utils/classnames'
import { getSeverityTextClasses } from '@/utils/log-stream-utils'

const isCommandOutput = (log: TOTELLog) =>
  log.log_attributes?.['nuon.command_output'] === 'true'

interface ICellTerminal {
  logs?: TOTELLog[]
  isLoading?: boolean
  connectionState?: string
  command?: string
  runCreatedAt?: string
  runUpdatedAt?: string
  isRunComplete?: boolean
  runFailed?: boolean
}

export const CellTerminal = ({
  logs,
  isLoading,
  connectionState,
  command,
  runCreatedAt,
  runUpdatedAt,
  isRunComplete,
  runFailed,
}: ICellTerminal) => {
  const [showAll, setShowAll] = useState(false)
  const [follow, setFollow] = useState(true)
  const [, setTick] = useState(0)
  const scrollRef = useRef<HTMLDivElement>(null)
  const isScrollingToBottom = useRef(false)

  const isLive = !isRunComplete

  const lines = useMemo(() => {
    const all = logs ?? []
    if (showAll) return all
    const output = all.filter(isCommandOutput)
    return output.length ? output : all.filter((l) => l.scope_name === 'oteljob')
  }, [logs, showAll])

  const hasChatter = (logs ?? []).some((l) => !isCommandOutput(l))

  useEffect(() => {
    if (!isLive) return
    const id = setInterval(() => setTick((t) => t + 1), 250)
    return () => clearInterval(id)
  }, [isLive])

  useEffect(() => {
    if (!follow) return
    const el = scrollRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [lines.length, follow])

  const onScroll = () => {
    if (isScrollingToBottom.current) return
    const el = scrollRef.current
    if (!el) return
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 24
    setFollow(atBottom)
  }

  const jumpToBottom = () => {
    const el = scrollRef.current
    if (!el) return
    isScrollingToBottom.current = true
    el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' })
    const check = () => {
      if (!el) return
      const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 24
      if (atBottom) {
        isScrollingToBottom.current = false
        setFollow(true)
      } else {
        requestAnimationFrame(check)
      }
    }
    requestAnimationFrame(check)
  }

  const promptLines = command?.split('\n') ?? []

  return (
    <div className="relative overflow-hidden rounded-md bg-code">
      <div className="flex items-center justify-between px-3 pt-2 pb-1">
        <span className="font-mono text-[10px] uppercase tracking-wider text-cool-grey-700 dark:text-cool-grey-400">
          Output
        </span>
        <div className="flex items-center gap-2">
          {hasChatter ? (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowAll((v) => !v)}
            >
              {showAll ? 'Output only' : 'All logs'}
            </Button>
          ) : null}
          <ClickToCopyButton
            textToCopy={lines.map((l) => l.body).join('\n')}
          />
        </div>
      </div>

      <div
        ref={scrollRef}
        onScroll={onScroll}
        className="max-h-[400px] overflow-auto px-3 pb-2 font-mono text-xs leading-relaxed whitespace-pre-wrap break-all"
      >
        {promptLines.map((line, i) => (
          <div key={`prompt-${i}`} className="text-cool-grey-700 dark:text-cool-grey-400">
            <span className="select-none">{i === 0 ? '$ ' : '  '}</span>
            {line}
          </div>
        ))}
        {lines.map((log) => (
          <div
            key={log.id}
            className={cn(getSeverityTextClasses(log.severity_number))}
          >
            {log.body}
          </div>
        ))}
        {isLive ? (
          <span className="ml-0.5 inline-block h-3.5 w-2 animate-pulse bg-primary-400 align-middle" />
        ) : null}
        {!lines.length && !promptLines.length ? (
          <div className="text-cool-grey-700 dark:text-cool-grey-400">
            {isLoading
              ? 'Waiting for output...'
              : connectionState === 'connected'
                ? 'No output yet.'
                : 'No output.'}
          </div>
        ) : null}
      </div>

      <div className="flex items-center gap-2 border-t border-cool-grey-200 dark:border-dark-grey-700 px-3 py-1.5 font-mono text-xs">
        {isRunComplete ? (
          <span
            className={cn(
              'inline-flex items-center gap-1',
              runFailed
                ? 'text-red-600 dark:text-red-500'
                : 'text-green-600 dark:text-green-500'
            )}
          >
            <Icon variant={runFailed ? 'XIcon' : 'CheckIcon'} size={12} />
            exit {runFailed ? 1 : 0}
          </span>
        ) : (
          <span className="text-cool-grey-700 dark:text-cool-grey-400">
            running...
          </span>
        )}
        {runCreatedAt ? (
          <>
            <span className="text-cool-grey-600 dark:text-cool-grey-500">·</span>
            <Duration
              variant="subtext"
              family="mono"
              className={
                isLive
                  ? 'text-blue-600 dark:text-blue-500'
                  : 'text-cool-grey-600 dark:text-cool-grey-400'
              }
              beginTime={runCreatedAt}
              endTime={isLive ? undefined : runUpdatedAt}
              durationUnits={
                isLive
                  ? ['minutes', 'seconds']
                  : ['minutes', 'seconds', 'milliseconds']
              }
            />
          </>
        ) : null}
      </div>

      {!follow ? (
        <Button
          variant="secondary"
          size="sm"
          onClick={jumpToBottom}
          className="absolute right-3 bottom-12 z-[1]"
        >
          Jump to bottom
          <Icon variant="ArrowDownIcon" size={12} />
        </Button>
      ) : null}
    </div>
  )
}
