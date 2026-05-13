import { useState } from 'react'
import { Badge } from './Badge'

function getStatus(s: any): string {
  if (!s) return ''
  if (typeof s === 'string') return s
  if (typeof s === 'object' && s.status) return getStatus(s.status)
  return String(s)
}

interface IStatusHistory {
  status: any
  defaultExpanded?: boolean
  maxCollapsed?: number
}

export const StatusHistory = ({ status, defaultExpanded = false, maxCollapsed = 3 }: IStatusHistory) => {
  const [expanded, setExpanded] = useState(defaultExpanded)

  if (!status) return null

  const currentStatus = getStatus(status)
  const history: any[] = status.history || []
  // Most recent first: current status, then history reversed (newest to oldest)
  const allEntries = [{ ...status, _isCurrent: true }, ...[...history].reverse()]
  const visibleEntries = expanded ? allEntries : allEntries.slice(0, maxCollapsed)
  const hasMore = allEntries.length > maxCollapsed

  return (
    <div className="space-y-1.5">
      {/* Most recent status shown first */}
      {visibleEntries.map((h: any, i: number) => {
        const s = getStatus(h)
        return (
          <div key={i} className="flex items-start gap-2 text-xs border-b border-gray-100 pb-1.5 last:border-0 dark:border-gray-800">
            <Badge variant="status" status={s}>{s}</Badge>
            <div className="flex-1 space-y-0.5 min-w-0">
              <div className="flex items-center gap-1 flex-wrap">
                <span>{s}</span>
                {h._isCurrent && <span className="text-primary-600 font-medium dark:text-primary-400">(current)</span>}
                {h.status_human_description && <span className="text-gray-500 dark:text-gray-400">— {h.status_human_description}</span>}
              </div>
              {h.created_at_ts > 0 && (
                <div className="text-gray-400 font-mono text-[10px] dark:text-gray-500">
                  {new Date(h.created_at_ts * 1000).toISOString().replace('T', ' ').slice(0, 19)} UTC
                </div>
              )}
              {h.metadata && Object.keys(h.metadata).length > 0 && (
                <div className="flex flex-wrap gap-x-3 gap-y-0.5 text-gray-400 dark:text-gray-500">
                  {Object.entries(h.metadata).map(([k, v]) => (
                    <span key={k}><span className="text-gray-500 dark:text-gray-400">{k}:</span> {String(v)}</span>
                  ))}
                </div>
              )}
            </div>
          </div>
        )
      })}
      {hasMore && (
        <button
          onClick={() => setExpanded(!expanded)}
          className="text-xs text-primary-600 hover:text-primary-700 font-medium dark:text-primary-400 dark:hover:text-primary-300"
        >
          {expanded ? `Show less` : `Show ${allEntries.length - maxCollapsed} more...`}
        </button>
      )}
    </div>
  )
}
