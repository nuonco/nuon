import { api } from '@/lib/api'
import type { TQueueSignal } from '@/types/admin.types'

// Queue signals (global view)
export const getQueueSignalsGlobal = (params: { search?: string; signal_type?: string; page?: number }) =>
  api<{ signals: TQueueSignal[]; page: number; total_pages: number }>({ path: 'queue-signals/table', params })

export const getQueueSignalTypeOptions = () =>
  api<{ signal_types: string[] }>({ path: 'queue-signals/signal-type-options' })

// In-flight signals
export const getInFlightSignals = () =>
  api<{ signals: TQueueSignal[] }>({ path: 'in-flight-signals/table' })

// Signal catalog - Go returns { grouped: Record<string, SignalTypeInfo[]>, namespaces: string[] }
export const getSignalCatalog = () =>
  api<{ grouped: Record<string, any[]>; namespaces: string[] }>({ path: 'signal-catalog' })

// Signal catalog detail - Go returns { info: SignalTypeInfo, recent_signals: QueueSignal[] }
export const getSignalCatalogDetail = (signalType: string) =>
  api<{ info: any; recent_signals: TQueueSignal[] }>({ path: `signal-catalog/${encodeURIComponent(signalType)}` })
