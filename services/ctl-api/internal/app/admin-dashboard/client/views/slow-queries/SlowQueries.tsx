import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { getQueries, explainQuery, type TQueryRecord } from '@/lib/admin-api'
import { SearchInput } from '@/components/common/SearchInput'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { Badge } from '@/components/common/Badge'

const SQL_KEYWORDS = /\b(SELECT|FROM|WHERE|AND|OR|LEFT JOIN|RIGHT JOIN|INNER JOIN|OUTER JOIN|FULL JOIN|JOIN|ON|ORDER BY|GROUP BY|HAVING|LIMIT|OFFSET|INSERT INTO|VALUES|UPDATE|SET|DELETE FROM|IN|NOT|IS|NULL|AS|CASE|WHEN|THEN|ELSE|END|DISTINCT|UNION|ALL|EXISTS|BETWEEN|LIKE|ILIKE|RETURNING)\b/gi

function formatSQL(sql: string): string {
  const breakBefore = /\b(SELECT|FROM|WHERE|AND|OR|LEFT JOIN|RIGHT JOIN|INNER JOIN|OUTER JOIN|FULL JOIN|JOIN|ON|ORDER BY|GROUP BY|HAVING|LIMIT|OFFSET|RETURNING|UNION)\b/gi
  let formatted = sql.replace(/\s+/g, ' ').trim()
  formatted = formatted.replace(breakBefore, (match) => {
    const upper = match.toUpperCase()
    if (upper === 'SELECT') return 'SELECT'
    if (['AND', 'OR'].includes(upper)) return `\n  ${upper}`
    if (upper.includes('JOIN') || upper === 'ON') return `\n  ${upper}`
    return `\n${upper}`
  })
  return formatted
}

const SORT_OPTIONS = [
  { value: 'max_ms', label: 'Slowest (max)' },
  { value: 'avg_ms', label: 'Slowest (avg)' },
  { value: 'total_ms', label: 'Most total time' },
  { value: 'count', label: 'Most frequent' },
  { value: 'last_seen', label: 'Most recent' },
] as const

const TIME_RANGE_OPTIONS = [
  { value: '5m', label: '5 minutes' },
  { value: '15m', label: '15 minutes' },
  { value: '30m', label: '30 minutes' },
  { value: '1h', label: '1 hour' },
  { value: '6h', label: '6 hours' },
  { value: '24h', label: '24 hours' },
  { value: '7d', label: '7 days' },
] as const

const fmt = (ms: number) => {
  if (ms < 1) return `${(ms * 1000).toFixed(0)}us`
  if (ms < 1000) return `${ms.toFixed(1)}ms`
  return `${(ms / 1000).toFixed(2)}s`
}

const dbBadgeColor = (dbType: string) =>
  dbType === 'ch'
    ? 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300'
    : 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300'

const sourceBadgeColor = (source: string) =>
  source === 'worker'
    ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/40 dark:text-purple-300'
    : 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300'

export const SlowQueries = () => {
  const [search, setSearch] = useState('')
  const [table, setTable] = useState('')
  const [dbType, setDbType] = useState('')
  const [source, setSource] = useState('')
  const [sortBy, setSortBy] = useState('max_ms')
  const [timeRange, setTimeRange] = useState('1h')
  const [minDuration, setMinDuration] = useState('')
  const [expanded, setExpanded] = useState<string | null>(null)
  const [explainResult, setExplainResult] = useState<{ key: string; rows?: Record<string, unknown>[]; error?: string; loading?: boolean } | null>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: ['queries', search, table, dbType, source, sortBy, minDuration, timeRange],
    queryFn: () => getQueries({
      search: search || undefined,
      table: table || undefined,
      db_type: dbType || undefined,
      source: source || undefined,
      sort: sortBy,
      min_duration_ms: minDuration || undefined,
      time_range: timeRange,
    }),
    refetchInterval: 10000,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load queries'} />

  const queries = data?.queries || []
  const tables = [...(data?.tables || [])].sort()

  return (
    <div>
      <div className="flex items-center justify-between">
        <h1 className="page-heading">Queries</h1>
        <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <span>{data?.total?.toLocaleString()} total in range</span>
          <span className="text-gray-300 dark:text-gray-600">|</span>
          <span>{queries.length} unique</span>
        </div>
      </div>

      <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:flex-wrap">
        <div className="w-full sm:w-64">
          <SearchInput value={search} onChange={setSearch} placeholder="Search SQL, table, caller..." />
        </div>
        <select
          value={timeRange}
          onChange={(e) => setTimeRange(e.target.value)}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 dark:bg-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
        >
          {TIME_RANGE_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
        <select
          value={table}
          onChange={(e) => setTable(e.target.value)}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 dark:bg-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
        >
          <option value="">All tables</option>
          {tables.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>
        <select
          value={dbType}
          onChange={(e) => setDbType(e.target.value)}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 dark:bg-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
        >
          <option value="">All databases</option>
          <option value="psql">PostgreSQL</option>
          <option value="ch">ClickHouse</option>
        </select>
        <select
          value={source}
          onChange={(e) => setSource(e.target.value)}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 dark:bg-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
        >
          <option value="">All sources</option>
          <option value="worker">Worker</option>
          <option value="api">API</option>
        </select>
        <select
          value={sortBy}
          onChange={(e) => setSortBy(e.target.value)}
          className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 dark:bg-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
        >
          {SORT_OPTIONS.map((o) => (
            <option key={o.value} value={o.value}>{o.label}</option>
          ))}
        </select>
        <input
          type="number"
          value={minDuration}
          onChange={(e) => setMinDuration(e.target.value)}
          placeholder="Min ms"
          className="w-24 rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 dark:bg-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
        />
      </div>

      <div className="mt-4 space-y-1">
        {queries.map((q: TQueryRecord, i: number) => {
          const key = `${i}`
          const isExpanded = expanded === key
          const isSlow = q.max_ms >= 50

          return (
            <div key={key} className="rounded-md border border-gray-200 dark:border-gray-800">
              <button
                onClick={() => setExpanded(isExpanded ? null : key)}
                className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-gray-50 dark:hover:bg-gray-800"
              >
                <span className={`font-mono text-xs font-semibold tabular-nums w-16 flex-shrink-0 ${isSlow ? 'text-red-600 dark:text-red-400' : 'text-gray-700 dark:text-gray-300'}`}>
                  {fmt(q.max_ms)}
                </span>
                <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium ${dbBadgeColor(q.db_type)}`}>
                  {q.db_type}
                </span>
                <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-[10px] font-medium ${sourceBadgeColor(q.source)}`}>
                  {q.source || 'worker'}
                </span>
                <Badge>{q.operation}</Badge>
                <span className="text-gray-500 dark:text-gray-400 text-xs truncate max-w-[120px]">{q.table}</span>
                <span className="ml-auto flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500 flex-shrink-0">
                  <span title="Execution count" className="font-semibold text-gray-600 dark:text-gray-300">{q.count}x</span>
                  <span title="Avg duration">avg {fmt(q.avg_ms)}</span>
                  <span title="Max rows">{q.max_rows} rows</span>
                  {q.endpoint && <span title="Endpoint" className="truncate max-w-[140px]">{q.endpoint}</span>}
                </span>
                {q.last_error && <span className="text-xs text-red-500 dark:text-red-400">err</span>}
              </button>
              {isExpanded && (
                <div className="border-t border-gray-100 dark:border-gray-800 px-3 py-2 space-y-2">
                  <div className="relative group">
                    <pre className="whitespace-pre-wrap break-all text-xs text-gray-700 dark:text-gray-300 font-mono bg-gray-50 dark:bg-gray-800 rounded p-2 pr-10 max-h-60 overflow-auto">
                      {formatSQL(q.sql)}
                    </pre>
                    <button
                      onClick={() => navigator.clipboard.writeText(q.sql)}
                      title="Copy SQL"
                      className="absolute top-1.5 right-1.5 opacity-0 group-hover:opacity-100 transition-opacity rounded bg-gray-200 dark:bg-gray-700 px-1.5 py-0.5 text-[10px] text-gray-600 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600"
                    >
                      Copy
                    </button>
                  </div>
                  <div className="flex items-center gap-2">
                    {q.operation === 'Query' && (
                      <button
                        onClick={async (e) => {
                          e.stopPropagation()
                          setExplainResult({ key, loading: true })
                          try {
                            const res = await explainQuery({ sql: q.sql, db_type: q.db_type })
                            if (res.error) {
                              setExplainResult({ key, error: res.error })
                            } else {
                              setExplainResult({ key, rows: res.rows })
                            }
                          } catch (err: any) {
                            setExplainResult({ key, error: err?.error || err?.message || 'Explain failed' })
                          }
                        }}
                        disabled={explainResult?.key === key && explainResult?.loading}
                        className="rounded-md bg-indigo-50 dark:bg-indigo-900/30 px-2 py-1 text-xs font-medium text-indigo-700 dark:text-indigo-300 hover:bg-indigo-100 dark:hover:bg-indigo-900/50 disabled:opacity-50"
                      >
                        {explainResult?.key === key && explainResult?.loading ? 'Running...' : 'Explain'}
                      </button>
                    )}
                    {q.caller && (
                      <div className="text-xs text-blue-600 dark:text-blue-400 font-mono">
                        {q.caller_url ? (
                          <a href={q.caller_url} target="_blank" rel="noopener noreferrer" className="underline hover:text-blue-800 dark:hover:text-blue-300">
                            {q.caller}
                          </a>
                        ) : (
                          q.caller
                        )}
                      </div>
                    )}
                  </div>
                  {explainResult?.key === key && !explainResult.loading && (
                    <div className="rounded border border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 p-2">
                      {explainResult.error ? (
                        <div className="text-xs text-red-600 dark:text-red-400 font-mono">{explainResult.error}</div>
                      ) : (
                        <pre className="whitespace-pre-wrap break-all text-xs text-gray-700 dark:text-gray-300 font-mono max-h-80 overflow-auto">
                          {(explainResult.rows || []).map((row) =>
                            Object.values(row).join('\n')
                          ).join('\n')}
                        </pre>
                      )}
                    </div>
                  )}
                  {q.endpoint && (
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      Endpoint: <span className="font-mono">{q.endpoint}</span>
                    </div>
                  )}
                  {q.last_error && (
                    <div className="text-xs text-red-600 dark:text-red-400">
                      Error: {q.last_error}
                    </div>
                  )}
                  <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
                    <span>Count: {q.count}</span>
                    <span>Min: {fmt(q.min_ms)}</span>
                    <span>Avg: {fmt(q.avg_ms)}</span>
                    <span>Max: {fmt(q.max_ms)}</span>
                    <span>Total: {fmt(q.total_ms)}</span>
                    <span>Max rows: {q.max_rows}</span>
                    <span>Max resp: {q.max_response_size}</span>
                    <span>Total rows: {q.total_rows}</span>
                    <span>DB: {q.db_type}</span>
                    <span>Source: {q.source || 'worker'}</span>
                  </div>
                </div>
              )}
            </div>
          )
        })}
        {queries.length === 0 && (
          <div className="text-center text-gray-500 dark:text-gray-400 py-8 text-sm">No queries found in the selected time range</div>
        )}
      </div>
    </div>
  )
}
