import { api } from '@/lib/api'

export type TQueueStatus = {
  ready: boolean
  stopped: boolean
  paused: boolean
  queue_depth_count: number
  in_flight_count: number
  in_flight: string[]
}

export const getQueueStatus = ({
  queueId,
  orgId,
}: {
  queueId: string
  orgId: string
}) =>
  api<TQueueStatus>({
    path: `queues/${queueId}/status`,
    orgId,
  })
