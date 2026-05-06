import { Link } from 'react-router'
import { truncateId } from '@/utils/format'

interface ISignalLink {
  signalId: string
  queueId: string
  showGraph?: boolean
  children?: React.ReactNode
}

export const SignalLink = ({ signalId, queueId, showGraph = true, children }: ISignalLink) => (
  <span className="inline-flex items-center gap-1.5">
    <Link to={`/queues/${queueId}/signals/${signalId}`} className="font-mono text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300">
      {children || truncateId(signalId)}
    </Link>
    {showGraph && (
      <Link
        to={`/queues/${queueId}/signals/${signalId}/graph`}
        className="inline-flex items-center rounded bg-primary-50 border border-primary-200 px-1 py-0.5 text-[9px] font-medium text-primary-600 hover:bg-primary-100 dark:bg-primary-950 dark:border-primary-900 dark:text-primary-300 dark:hover:bg-primary-900"
        title="View signal graph"
      >
        graph
      </Link>
    )}
  </span>
)
