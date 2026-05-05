import { useQuery, useMutation } from '@tanstack/react-query'
import { useState } from 'react'
import { getQueryCatalog, runCatalogQuery, type TCatalogQuery } from '@/lib/admin-api'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { Badge } from '@/components/common/Badge'

export const QueryCatalog = () => {
  const [activeQuery, setActiveQuery] = useState<string | null>(null)
  const [results, setResults] = useState<{ query: TCatalogQuery; rows: Record<string, any>[]; count: number } | null>(null)
  const [runError, setRunError] = useState<string | null>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: ['query-catalog'],
    queryFn: getQueryCatalog,
  })

  const runMutation = useMutation({
    mutationFn: (queryId: string) => runCatalogQuery(queryId),
    onSuccess: (data) => {
      setResults({ query: data.query, rows: data.results || [], count: data.count })
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
      <h1 className="page-heading">Query catalog</h1>

      <div className="mt-4 space-y-2">
        {queries.map((q) => {
          const isActive = activeQuery === q.id
          const isRunning = runMutation.isPending && runMutation.variables === q.id

          return (
            <div key={q.id} className="rounded-md border border-gray-200 bg-white">
              <button
                onClick={() => setActiveQuery(isActive ? null : q.id)}
                className="flex w-full items-center gap-3 px-3 py-2.5 text-left text-sm hover:bg-gray-50"
              >
                <span className="font-medium text-gray-900">{q.name}</span>
                <Badge>{q.db_type}</Badge>
                <span className="ml-auto text-xs text-gray-400">{q.description}</span>
              </button>
              {isActive && (
                <div className="border-t border-gray-100 px-3 py-3 space-y-3">
                  <pre className="whitespace-pre-wrap text-xs text-gray-700 font-mono bg-gray-50 rounded p-2 max-h-40 overflow-auto">
                    {q.sql}
                  </pre>
                  <div className="flex items-center gap-3">
                    <button
                      onClick={() => runMutation.mutate(q.id)}
                      disabled={isRunning}
                      className="rounded-md bg-primary-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-primary-700 disabled:opacity-50"
                    >
                      {isRunning ? 'Running...' : 'Run query'}
                    </button>
                    {results && results.query.id === q.id && (
                      <span className="text-xs text-gray-500">{results.count} rows returned</span>
                    )}
                    {runError && runMutation.variables === q.id && (
                      <span className="text-xs text-red-600">{runError}</span>
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
          <h2 className="text-sm font-semibold text-gray-900 mb-2">
            Results: {results.query.name}
            <span className="ml-2 font-normal text-gray-500">({results.count} rows)</span>
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
                <tbody className="divide-y divide-gray-200">
                  {results.rows.map((row, i) => (
                    <tr key={i}>
                      {Object.values(row).map((val, j) => (
                        <td key={j} className="text-xs font-mono text-gray-700 whitespace-nowrap">
                          {val === null ? <span className="text-gray-300">null</span> : String(val)}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-sm text-gray-500 py-4">No results</div>
          )}
        </div>
      )}
    </div>
  )
}
