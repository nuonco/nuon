import { useMemo, useState } from 'react'
import { cn } from '@/utils/classnames'
import { generateDiffLines, type DiffLine } from '@/utils/terraform-utils'

interface TreeDiffValueProps {
  before: any
  after: any
}

const COLLAPSE_THRESHOLD = 200
const INITIAL_VISIBLE = 50

const lineStyles: Record<DiffLine['type'], string> = {
  added: 'bg-green-500/15 dark:bg-green-500/5 text-green-800 dark:text-green-400',
  removed: 'bg-red-500/15 dark:bg-red-500/5 text-red-800 dark:text-red-400',
  changed: 'bg-orange-500/15 dark:bg-orange-500/5 text-orange-800 dark:text-orange-400',
  unchanged: 'text-current',
}

const prefixChars: Record<DiffLine['prefix'], string> = {
  '+': '+',
  '-': '-',
  '~': '~',
  ' ': ' ',
}

export function TreeDiffValue({ before, after }: TreeDiffValueProps) {
  const [expanded, setExpanded] = useState(false)

  const lines = useMemo(
    () => generateDiffLines(before, after),
    [before, after]
  )

  const needsCollapse = lines.length > COLLAPSE_THRESHOLD
  const visibleLines =
    needsCollapse && !expanded ? lines.slice(0, INITIAL_VISIBLE) : lines

  return (
    <div className="ml-4 my-1">
      <div className="font-mono text-[13px] leading-6 overflow-x-auto">
        {visibleLines.map((line, i) => (
          <div
            key={i}
            className={cn('flex whitespace-pre', lineStyles[line.type])}
          >
            <span
              className="inline-block w-[2ch] shrink-0 select-none text-right mr-2 opacity-70"
            >
              {prefixChars[line.prefix]}
            </span>
            <span
              style={{ paddingLeft: `${line.indent * 1.5}rem` }}
              className={cn(line.type === 'removed' && 'line-through opacity-70')}
            >
              {line.text === 'Known after apply' ||
              line.text === '"Known after apply"' ? (
                <em className="opacity-60">{line.text}</em>
              ) : (
                line.text
              )}
            </span>
          </div>
        ))}
        {needsCollapse && !expanded && (
          <button
            type="button"
            className="ml-[2ch] pl-2 text-xs text-blue-500 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 py-1"
            onClick={() => setExpanded(true)}
          >
            Show all {lines.length} lines
          </button>
        )}
      </div>
    </div>
  )
}
