import { useQuery } from '@tanstack/react-query'
import { getTemporalWorkflowStats } from '@/lib/admin-api'
import { Badge } from './Badge'
import { DateTime } from 'luxon'

interface ITemporalWorkflowCard {
  temporalUIUrl: string
  namespace: string
  workflowId: string
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function age(startTime: string): string {
  if (!startTime) return '-'
  const dt = DateTime.fromISO(startTime)
  if (!dt.isValid) return '-'
  const diff = dt.diffNow().negate()
  if (diff.as('days') >= 1) return `${Math.floor(diff.as('days'))}d ${Math.floor(diff.as('hours') % 24)}h`
  if (diff.as('hours') >= 1) return `${Math.floor(diff.as('hours'))}h ${Math.floor(diff.as('minutes') % 60)}m`
  return `${Math.floor(diff.as('minutes'))}m`
}

export const TemporalWorkflowCard = ({ temporalUIUrl, namespace, workflowId }: ITemporalWorkflowCard) => {
  const { data: stats } = useQuery({
    queryKey: ['temporal-workflow-stats', namespace, workflowId],
    queryFn: () => getTemporalWorkflowStats({ namespace, workflow_id: workflowId }),
    enabled: !!namespace && !!workflowId,
    staleTime: 60_000,
  })

  return (
    <div className="rounded-lg border border-gray-200 dark:border-gray-800 p-3 text-xs space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-gray-500 dark:text-gray-400 font-medium uppercase">Temporal workflow</span>
        {temporalUIUrl && (
          <a
            href={`${temporalUIUrl}/namespaces/${namespace}/workflows/${workflowId}`}
            target="_blank"
            rel="noopener noreferrer"
            className="text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300"
          >
            View in Temporal &rarr;
          </a>
        )}
      </div>
      {stats ? (
        <div className="grid grid-cols-2 gap-x-6 gap-y-1.5 sm:grid-cols-4">
          <div>
            <div className="text-gray-400 dark:text-gray-500">Status</div>
            <Badge variant="status" status={stats.status}>{stats.status}</Badge>
          </div>
          <div>
            <div className="text-gray-400 dark:text-gray-500">Events</div>
            <div className="font-mono">{stats.history_length.toLocaleString()}</div>
          </div>
          <div>
            <div className="text-gray-400 dark:text-gray-500">Size</div>
            <div className="font-mono">{formatBytes(stats.history_size_bytes)}</div>
          </div>
          <div>
            <div className="text-gray-400 dark:text-gray-500">CAN count</div>
            <div className="font-mono">{stats.can_count > 0 ? stats.can_count.toLocaleString() : '0'}</div>
          </div>
          <div>
            <div className="text-gray-400 dark:text-gray-500">Age</div>
            <div className="font-mono">{age(stats.start_time)}</div>
          </div>
          <div>
            <div className="text-gray-400 dark:text-gray-500">Namespace</div>
            <div className="font-mono">{namespace}</div>
          </div>
        </div>
      ) : (
        <div className="text-gray-400 dark:text-gray-500">Loading stats...</div>
      )}
    </div>
  )
}
