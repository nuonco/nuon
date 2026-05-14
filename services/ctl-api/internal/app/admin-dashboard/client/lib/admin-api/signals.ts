import { api } from '@/lib/api'
import type { TQueueSignal } from '@/types/admin.types'

// Queue signals (global view)
export const getQueueSignalsGlobal = (params: {
  search?: string
  signal_type?: string
  owner_id?: string
  org_id?: string
  namespace?: string
  status?: string
  enqueued?: string
  sort_by?: string
  since?: string
  page?: number
}) =>
  api<{
    signals: TQueueSignal[]
    page: number
    total_pages: number
    namespaces?: string[]
    signal_types?: string[]
    org_options?: { id: string; name: string }[]
  }>({
    path: 'queue-signals/table',
    params,
  })

export const getQueueSignalTypeOptions = (namespace?: string) =>
  api<{ signal_types: string[] }>({ path: 'queue-signals/signal-type-options', params: { namespace } })

// In-flight signals
export const getInFlightSignals = (params?: { namespace?: string }) =>
  api<{ signals: TQueueSignal[]; namespaces?: string[] }>({ path: 'in-flight-signals/table', params })

// Signal catalog - Go returns { grouped: Record<string, SignalTypeInfo[]>, namespaces: string[] }
export const getSignalCatalog = () =>
  api<{ grouped: Record<string, any[]>; namespaces: string[] }>({ path: 'signal-catalog' })

// Signal catalog detail - Go returns { info: SignalTypeInfo, recent_signals: QueueSignal[] }
export const getSignalCatalogDetail = (signalType: string) =>
  api<{ info: any; recent_signals: TQueueSignal[] }>({ path: `signal-catalog/${encodeURIComponent(signalType)}` })
