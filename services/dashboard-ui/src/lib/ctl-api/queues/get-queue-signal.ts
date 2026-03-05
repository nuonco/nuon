import { api } from '@/lib/api'
import type { TQueueSignal } from './get-queue-signals'

export const getQueueSignal = ({
  queueId,
  signalId,
  orgId,
}: {
  queueId: string
  signalId: string
  orgId: string
}) =>
  api<TQueueSignal>({
    path: `queues/${queueId}/signals/${signalId}`,
    orgId,
  })
