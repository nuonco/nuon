import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { getTemporalWorkflows } from '@/lib/admin-api'
import { JsonViewer } from '@/components/common/JsonViewer'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'

export const TemporalWorkflows = () => {
  const [workflowId, setWorkflowId] = useState('')
  const [runId, setRunId] = useState('')
  const [namespace, setNamespace] = useState('')
  const [submitted, setSubmitted] = useState(false)

  const { data, isLoading, error } = useQuery({
    queryKey: ['temporal-workflows', workflowId, runId, namespace],
    queryFn: () =>
      getTemporalWorkflows({
        workflow_id: workflowId || undefined,
        run_id: runId || undefined,
        namespace: namespace || undefined,
      }),
    enabled: submitted && !!(workflowId || runId),
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitted(true)
  }

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-bold text-gray-900">Temporal Workflows</h1>

      <form onSubmit={handleSubmit} className="rounded-lg border border-gray-200 bg-white p-4">
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1">Workflow ID</label>
            <input
              type="text"
              value={workflowId}
              onChange={(e) => { setWorkflowId(e.target.value); setSubmitted(false) }}
              placeholder="Workflow ID..."
              className="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1">Run ID</label>
            <input
              type="text"
              value={runId}
              onChange={(e) => { setRunId(e.target.value); setSubmitted(false) }}
              placeholder="Run ID..."
              className="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1">Namespace</label>
            <input
              type="text"
              value={namespace}
              onChange={(e) => { setNamespace(e.target.value); setSubmitted(false) }}
              placeholder="Namespace..."
              className="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-primary-600"
            />
          </div>
        </div>
        <div className="mt-3">
          <button
            type="submit"
            disabled={!workflowId && !runId}
            className="rounded-md bg-primary-600 px-4 py-1.5 text-sm font-medium text-white hover:bg-primary-700 disabled:opacity-50"
          >
            Search
          </button>
        </div>
      </form>

      {submitted && isLoading && <LoadingSpinner />}
      {submitted && error && <ErrorMessage message={(error as Error).message || 'Failed to load workflow'} />}
      {submitted && data?.workflow_info && (
        <div className="rounded-lg border border-gray-200 bg-white p-4">
          <h2 className="text-sm font-semibold text-gray-900 mb-2">Workflow Info</h2>
          <JsonViewer data={data.workflow_info} />
        </div>
      )}
      {submitted && !isLoading && !error && !data?.workflow_info && (
        <p className="text-sm text-gray-500">No workflow found with the given parameters</p>
      )}
    </div>
  )
}
