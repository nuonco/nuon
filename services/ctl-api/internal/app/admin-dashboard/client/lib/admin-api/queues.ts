import { api } from '@/lib/api'
import type { TQueuesResponse, TQueue, TQueueSignal, TQueueEmitter } from '@/types/admin.types'

export const getQueues = (params: { search?: string; name?: string; namespace?: string; page?: number }) =>
  api<TQueuesResponse>({ path: 'queues', params })

export const getQueueDetail = (id: string) =>
  api<{ queue: TQueue; signals: TQueueSignal[]; in_flight_signals: TQueueSignal[] }>({ path: `queues/${id}` })

export const getQueueEmitters = (id: string, params: { page?: number }) =>
  api<{ emitters: TQueueEmitter[]; page: number; total_pages: number }>({ path: `queues/${id}/emitters`, params })

export const getQueueSignals = (id: string, params: { page?: number }) =>
  api<{ signals: TQueueSignal[]; page: number; total_pages: number }>({ path: `queues/${id}/signals`, params })

export const getQueueInFlightSignals = (id: string) =>
  api<{ signals: TQueueSignal[] }>({ path: `queues/${id}/in-flight-signals` })

export const getQueueSignalDetail = (queueId: string, signalId: string) =>
  api<{ signal: TQueueSignal; workflow_info: any }>({ path: `queues/${queueId}/signals/${signalId}` })

export const getQueueEmitterDetail = (queueId: string, emitterId: string) =>
  api<{ emitter: TQueueEmitter }>({ path: `queues/${queueId}/emitters/${emitterId}` })

export const restartQueue = (id: string) =>
  api<{ status: string }>({ path: `queues/${id}/restart`, method: 'POST' })

export const clearQueue = (id: string) =>
  api<{ status: string }>({ path: `queues/${id}/clear`, method: 'POST' })
