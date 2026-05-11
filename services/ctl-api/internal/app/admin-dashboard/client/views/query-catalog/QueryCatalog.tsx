import { useQuery, useMutation } from '@tanstack/react-query'
import { useEffect, useState } from 'react'
import { getQueryCatalog, runCatalogQuery, type TCatalogQuery, type TQueryTarget } from '@/lib/admin-api'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { Badge } from '@/components/common/Badge'

export const QueryCatalog = () => {
  const [activeQuery, setActiveQuery] = useState<string | null>(null)
  const [results, setResults] = useState<{ query: TCatalogQuery; rows: Record<string, any>[]; count: number; target: TQueryTarget } | null>(null)
  const [runError, setRunError] = useState<string | null>(null)
  const [target, setTarget] = useState<TQueryTarget>('replica')
  const [resultsOpen, setResultsOpen] = useState(false)

  const { data, isLoading, error } = useQuery({
    queryKey: ['query-catalog'],
    queryFn: getQueryCatalog,
  })

  const runMutation = useMutation({
    mutationFn: (queryId: string) => runCatalogQuery(queryId, target),
    onSuccess: (data) => {
      setResults({ query: data.query, rows: data.results || [], count: data.count, target: data.target })
      setRunError(null)
      setResultsOpen(true)
    },
    onError: (err: any) => {
      setRunError(err?.error || err?.message || 'Query failed')
      setResults(null)
      setResultsOpen(false)
    },
  })

  useEffect(() => {
    if (!resultsOpen) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setResultsOpen(false)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [resultsOpen])

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load catalog'} />

  const queries = data?.queries || []

  return (
    <div>
      <div className="flex items-center justify-between gap-3">
        <h1 className="page-heading">Query catalog</h1>
        <div className="flex items-center gap-3">
          {results && !resultsOpen && (
            <button
              onClick={() => setResultsOpen(true)}
              className="rounded-md bg-primary-600 dark:bg-primary-500 px-3 py-1.5 text-xs font-medium text-white hover:bg-primary-700 dark:hover:bg-primary-600"
            >
              Show results ({results.count})
            </button>
          )}
          <label className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
            <span>Connection</span>
            <select
              value={target}
              onChange={(e) => setTarget(e.target.value as TQueryTarget)}
              className="rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 dark:text-gray-100 dark:bg-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 dark:ring-gray-700"
            >
              <option value="replica">Replica (default)</option>
              <option value="primary">Primary</option>
            </select>
          </label>
        </div>
      </div>

      <div className="mt-4 space-y-2">
        {queries.map((q) => {
          const isActive = activeQuery === q.id
          const isRunning = runMutation.isPending && runMutation.variables === q.id

          return (
            <div key={q.id} className="rounded-md border border-gray-200 dark:border-gray-800">
              <button
                onClick={() => setActiveQuery(isActive ? null : q.id)}
                className="flex w-full items-center gap-3 px-3 py-2.5 text-left text-sm hover:bg-gray-50 dark:hover:bg-gray-800"
              >
                <span className="font-medium text-gray-900 dark:text-gray-100">{q.name}</span>
                <Badge>{q.db_type}</Badge>
                <span className="ml-auto text-xs text-gray-400 dark:text-gray-500">{q.description}</span>
              </button>
              {isActive && (
                <div className="border-t border-gray-100 dark:border-gray-800 px-3 py-3 space-y-3">
                  <pre className="whitespace-pre-wrap text-xs text-gray-700 dark:text-gray-300 font-mono bg-gray-50 dark:bg-gray-800 rounded p-2 max-h-40 overflow-auto">
                    {q.sql}
                  </pre>
                  <div className="flex items-center gap-3">
                    <button
                      onClick={() => runMutation.mutate(q.id)}
                      disabled={isRunning}
                      className="rounded-md bg-primary-600 dark:bg-primary-500 px-3 py-1.5 text-xs font-medium text-white hover:bg-primary-700 dark:hover:bg-primary-600 disabled:opacity-50"
                    >
                      {isRunning ? 'Running...' : 'Run query'}
                    </button>
                    {results && results.query.id === q.id && (
                      <span className="text-xs text-gray-500 dark:text-gray-400">{results.count} rows returned</span>
                    )}
                    {runError && runMutation.variables === q.id && (
                      <span className="text-xs text-red-600 dark:text-red-400">{runError}</span>
                    )}
                  </div>
                </div>
              )}
            </div>
          )
        })}
      </div>

      {results && resultsOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div
            className="absolute inset-0 bg-black/40 dark:bg-black/60"
            onClick={() => setResultsOpen(false)}
          />
          <div className="relative z-10 flex max-h-[85vh] w-full max-w-6xl flex-col overflow-hidden rounded-lg bg-white dark:bg-gray-900 shadow-2xl ring-1 ring-gray-200 dark:ring-gray-800">
            <div className="flex items-center justify-between gap-3 border-b border-gray-200 dark:border-gray-800 px-4 py-3">
              <div className="min-w-0">
                <h2 className="truncate text-sm font-semibold text-gray-900 dark:text-gray-100">
                  Results: {results.query.name}
                </h2>
                <div className="text-xs text-gray-500 dark:text-gray-400">
                  {results.count} rows{results.query.db_type === 'psql' ? ` · ${results.target}` : ''}
                </div>
              </div>
              <button
                onClick={() => setResultsOpen(false)}
                aria-label="Close results"
                className="rounded-md p-1 text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800"
              >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className="h-5 w-5">
                  <path d="M6.28 5.22a.75.75 0 0 0-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 1 0 1.06 1.06L10 11.06l3.72 3.72a.75.75 0 1 0 1.06-1.06L11.06 10l3.72-3.72a.75.75 0 0 0-1.06-1.06L10 8.94 6.28 5.22Z" />
                </svg>
              </button>
            </div>
            <div className="flex-1 overflow-auto p-4">
              {results.rows.length > 0 ? (
                <div className="table-card overflow-x-auto">
                  <table>
                    <thead>
                      <tr>
                        {Object.keys(results.rows[0]).map((col) => (
                          <th key={col} className="text-xs">{col}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200 dark:divide-gray-800">
                      {results.rows.map((row, i) => (
                        <tr key={i}>
                          {Object.values(row).map((val, j) => (
                            <td key={j} className="text-xs font-mono text-gray-700 dark:text-gray-300 whitespace-nowrap">
                              {val === null ? <span className="text-gray-300 dark:text-gray-600">null</span> : String(val)}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="text-sm text-gray-500 dark:text-gray-400 py-4">No results</div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
