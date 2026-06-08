import { useEffect, useMemo, useRef, useState } from 'react'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { useLogStreamData } from '@/hooks/use-logs'
import { LogStreamProvider } from '@/providers/log-stream-provider'
import type { TOTELLog } from '@/types'
import { cn } from '@/utils/classnames'

interface INotebookCellLogs {
  logStreamId: string
  command?: string
  runCreatedAt?: string
  runUpdatedAt?: string
  isRunComplete?: boolean
  runFailed?: boolean
}

// A cell's output is a terminal, not a log table: default to the raw script
// stdout/stderr (lines tagged `nuon.command_output` by the runner) and hide
// the runner's own job-lifecycle chatter behind a toggle.
const isCommandOutput = (log: TOTELLog) =>
  log.log_attributes?.['nuon.command_output'] === 'true'

// On a dark terminal surface the shared severity palette (INFO → blue) reads
// wrong, so map to terminal-friendly colors: normal output bright, warnings
// amber, errors red, trace/debug dim.
const lineClass = (severityNumber?: number) => {
  const n = severityNumber ?? 9
  if (n >= 17) return 'text-red-400'
  if (n >= 13) return 'text-amber-400'
  if (n <= 4) return 'text-zinc-500'
  return 'text-zinc-100'
}

const CellTerminal = ({
  command,
  runCreatedAt,
  runUpdatedAt,
  isRunComplete,
  runFailed,
}: Omit<INotebookCellLogs, 'logStreamId'>) => {
  const { logs, isLoading, connectionState } = useLogStreamData()
  const [showAll, setShowAll] = useState(false)
  const [follow, setFollow] = useState(true)
  const [, setTick] = useState(0)
  const scrollRef = useRef<HTMLDivElement>(null)

  const isLive = !isRunComplete

  const lines = useMemo(() => {
    const all = logs ?? []
    if (showAll) return all
    const output = all.filter(isCommandOutput)
    // Fallback for runs recorded before the runner tagged command output:
    // show job-output scope so the terminal isn't mysteriously empty.
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
    const el = scrollRef.current
    if (!el) return
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 24
    setFollow(atBottom)
  }

  const copy = () =>
    navigator.clipboard?.writeText(lines.map((l) => l.body).join('\n'))

  const promptLines = command?.split('\n') ?? []

  return (
    <div className="relative overflow-hidden rounded-md bg-zinc-950">
      <div className="flex items-center justify-between px-3 pt-2 pb-1">
        <span className="font-mono text-[10px] uppercase tracking-wider text-zinc-500">
          Output
        </span>
        <div className="flex items-center gap-3">
          {hasChatter ? (
            <button
              type="button"
              onClick={() => setShowAll((v) => !v)}
              className="text-xs text-zinc-500 hover:text-zinc-300"
            >
              {showAll ? 'Output only' : 'All logs'}
            </button>
          ) : null}
          <button
            type="button"
            onClick={copy}
            title="Copy output"
            className="text-zinc-500 hover:text-zinc-300"
          >
            <Icon variant="CopyIcon" size={14} />
          </button>
        </div>
      </div>

      <div
        ref={scrollRef}
        onScroll={onScroll}
        className="max-h-[400px] overflow-auto px-3 pb-2 font-mono text-xs leading-relaxed whitespace-pre-wrap break-all"
      >
        {promptLines.map((line, i) => (
          <div key={`prompt-${i}`} className="text-zinc-500">
            <span className="select-none">{i === 0 ? '$ ' : '  '}</span>
            {line}
          </div>
        ))}
        {lines.map((log) => (
          <div key={log.id} className={cn(lineClass(log.severity_number))}>
            {log.body}
          </div>
        ))}
        {isLive ? (
          <span className="ml-0.5 inline-block h-3.5 w-2 animate-pulse bg-green-400 align-middle" />
        ) : null}
        {!lines.length && !promptLines.length ? (
          <div className="text-zinc-500">
            {isLoading
              ? 'Waiting for output...'
              : connectionState === 'connected'
                ? 'No output yet.'
                : 'No output.'}
          </div>
        ) : null}
      </div>

      <div className="flex items-center gap-2 border-t border-white/10 px-3 py-1.5 font-mono text-xs">
        {isRunComplete ? (
          <span
            className={cn(
              'inline-flex items-center gap-1',
              runFailed ? 'text-red-400' : 'text-green-400'
            )}
          >
            <Icon variant={runFailed ? 'XIcon' : 'CheckIcon'} size={12} />
            exit {runFailed ? 1 : 0}
          </span>
        ) : (
          <span className="text-zinc-500">running...</span>
        )}
        {runCreatedAt ? (
          <>
            <span className="text-zinc-600">·</span>
            <Duration
              variant="subtext"
              family="mono"
              className={isLive ? 'text-blue-400' : 'text-zinc-400'}
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
        <button
          type="button"
          onClick={() => setFollow(true)}
          className="absolute right-3 bottom-12 z-10 inline-flex items-center gap-1 rounded-full bg-white/10 px-2.5 py-1 text-xs text-zinc-100 backdrop-blur hover:bg-white/20"
        >
          Jump to bottom <Icon variant="ArrowDownIcon" size={12} />
        </button>
      ) : null}
    </div>
  )
}

// NotebookCellLogs renders a cell run's live output beneath the cell, reusing
// the shared SSE log stream provider but presenting it as a terminal.
export const NotebookCellLogs = ({ logStreamId, ...rest }: INotebookCellLogs) => (
  <LogStreamProvider logStreamId={logStreamId}>
    <CellTerminal {...rest} />
  </LogStreamProvider>
)
