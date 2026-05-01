import { api } from '@/lib/api'
import type { TQueuesResponse, TQueueDetailResponse, TQueue, TQueueSignal, TQueueEmitter } from '@/types/admin.types'

export const getQueues = (params: { search?: string; name?: string; namespace?: string; owner_id?: string; owner_type?: string; page?: number }) => {
  const { name, ...rest } = params
  return api<TQueuesResponse>({ path: 'queues', params: { ...rest, queue_name: name } })
}

export const getQueueDetail = (id: string) =>
  api<TQueueDetailResponse>({ path: `queues/${id}` })

export const getQueueEmitters = (id: string, params: { page?: number }) =>
  api<{ emitters: TQueueEmitter[]; page: number; total_pages: number }>({ path: `queues/${id}/emitters`, params })

export const getQueueSignals = (id: string, params: { page?: number }) =>
  api<{ signals: TQueueSignal[]; page: number; total_pages: number }>({ path: `queues/${id}/signals`, params })

export const getQueueInFlightSignals = (id: string) =>
  api<{ signals: TQueueSignal[] }>({ path: `queues/${id}/in-flight-signals` })

export const getQueueSignalDetail = (queueId: string, signalId: string) =>
  api<{
    signal: TQueueSignal & {
      workflow?: { id: string; namespace: string }
      execution_count?: number
      emitter_id?: string
      signal?: any
    }
    queue: TQueue
    temporal_ui_url: string
    workflow_info: any
    signal_attrs: any
    signals_ahead: TQueueSignal[]
  }>({ path: `queues/${queueId}/signals/${signalId}` })

export const getQueueEmitterDetail = (queueId: string, emitterId: string) =>
  api<{ emitter: TQueueEmitter; queue: TQueue; signals: TQueueSignal[]; temporal_ui_url: string }>({ path: `queues/${queueId}/emitters/${emitterId}` })

export const getSignalGraph = (queueId: string, signalId: string, depth = 1) =>
  api<{ graph: any; temporal_ui_url: string }>({ path: `queues/${queueId}/signals/${signalId}/graph`, params: { depth } })

export const restartQueue = (id: string) =>
  api<{ status: string }>({ path: `queues/${id}/restart`, method: 'POST' })

export const forceRestartQueue = (id: string) =>
  api<{ status: string }>({ path: `queues/${id}/force-restart`, method: 'POST' })

export const clearQueue = (id: string) =>
  api<{ status: string }>({ path: `queues/${id}/clear`, method: 'POST' })

export const directExecuteSignal = (queueId: string, signalId: string) =>
  api<{ status: string; queue_signal_id: string }>({
    path: `queues/${queueId}/signals/${signalId}/direct-execute`,
    method: 'POST',
  })
