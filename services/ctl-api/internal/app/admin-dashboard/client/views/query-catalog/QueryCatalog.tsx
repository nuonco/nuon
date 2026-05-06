import { useQuery, useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { getQueryCatalog, runCatalogQuery, type TCatalogQuery, type TQueryTarget } from '@/lib/admin-api'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { Badge } from '@/components/common/Badge'

export const QueryCatalog = () => {
  const [activeQuery, setActiveQuery] = useState<string | null>(null)
  const [results, setResults] = useState<{ query: TCatalogQuery; rows: Record<string, any>[]; count: number; target: TQueryTarget } | null>(null)
  const [runError, setRunError] = useState<string | null>(null)
  const [target, setTarget] = useState<TQueryTarget>('replica')

  const { data, isLoading, error } = useQuery({
    queryKey: ['query-catalog'],
    queryFn: getQueryCatalog,
  })

  const runMutation = useMutation({
    mutationFn: (queryId: string) => runCatalogQuery(queryId, target),
    onSuccess: (data) => {
      setResults({ query: data.query, rows: data.results || [], count: data.count, target: data.target })
      setRunError(null)
    },
    onError: (err: any) => {
      setRunError(err?.error || err?.message || 'Query failed')
      setResults(null)
    },
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load catalog'} />

  const queries = data?.queries || []

  return (
    <div>
      <div className="flex items-center justify-between gap-3">
        <h1 className="page-heading">Query catalog</h1>
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

      {results && (
        <div className="mt-6">
          <h2 className="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-2">
            Results: {results.query.name}
            <span className="ml-2 font-normal text-gray-500 dark:text-gray-400">
              ({results.count} rows{results.query.db_type === 'psql' ? ` · ${results.target}` : ''})
            </span>
          </h2>
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
      )}
    </div>
  )
}
