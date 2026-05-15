import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link, useParams } from 'react-router'
import { getSignalGraph } from '@/lib/admin-api'
import { SignalFlowGraph } from '@/components/common/SignalFlowGraph'
import { SignalTreeView } from '@/components/common/SignalTreeView'
import { LoadingSpinner } from '@/components/common/LoadingSpinner'
import { ErrorMessage } from '@/components/common/ErrorMessage'
import { truncateId } from '@/utils/format'

export const SignalGraphView = () => {
  const { id: queueId, signalId } = useParams<{ id: string; signalId: string }>()
  const [viewMode, setViewMode] = useState<'tree' | 'graph'>('tree')

  const { data, isLoading, error } = useQuery({
    queryKey: ['signal-graph', queueId, signalId],
    queryFn: () => getSignalGraph(queueId!, signalId!, 2),
    enabled: !!queueId && !!signalId,
  })

  if (isLoading) return <LoadingSpinner />
  if (error) return <ErrorMessage message={(error as Error).message || 'Failed to load signal graph'} />

  return (
    <div className="-m-6 lg:-m-8 flex flex-col" style={{ height: 'calc(100vh - 3rem)' }}>
      {/* Compact toolbar */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 dark:border-gray-800 shrink-0">
        <div className="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
          <Link to={`/queues/${queueId}/signals/${signalId}`} className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300">
            &larr; Signal {truncateId(signalId || '')}
          </Link>
          <span className="text-gray-300 dark:text-gray-400">|</span>
          <span className="font-medium text-gray-900 dark:text-gray-100">Signal tree</span>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex rounded-md border border-gray-300 dark:border-gray-700 text-xs overflow-hidden">
            <button
              onClick={() => setViewMode('tree')}
              className={`px-2.5 py-1 font-medium ${viewMode === 'tree' ? 'bg-primary-600 dark:bg-primary-500 text-white' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'}`}
            >
              Tree
            </button>
            <button
              onClick={() => setViewMode('graph')}
              className={`px-2.5 py-1 font-medium ${viewMode === 'graph' ? 'bg-primary-600 dark:bg-primary-500 text-white' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'}`}
            >
              Graph
            </button>
          </div>
          <Link
            to={`/queues/${queueId}/signals/${signalId}`}
            className="rounded-md bg-gray-100 dark:bg-gray-800 px-2.5 py-1 text-xs font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700"
          >
            Back to detail
          </Link>
        </div>
      </div>

      {/* View fills remaining space */}
      <div className="flex-1 min-h-0">
        {data?.graph ? (
          viewMode === 'tree' ? (
            <SignalTreeView
              graphData={data.graph}
              temporalUIUrl={data.temporal_ui_url}
              height="100%"
            />
          ) : (
            <SignalFlowGraph graphData={data.graph} height="100%" />
          )
        ) : (
          <div className="flex items-center justify-center h-full text-sm text-gray-500 dark:text-gray-400">No graph data available</div>
        )}
      </div>
    </div>
  )
}
