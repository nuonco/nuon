import { api } from '@/lib/api'
import type { TQueueSignalsResponse, TInFlightSignalsResponse, TSignalCatalogResponse, TSignalCatalogDetailResponse } from '@/types/admin.types'

export const getQueueSignalsGlobal = (params: { search?: string; signal_type?: string; page?: number }) =>
  api<TQueueSignalsResponse>({ path: 'queue-signals', params })

export const getQueueSignalTypeOptions = () =>
  api<{ signal_types: string[] }>({ path: 'queue-signals/signal-type-options' })

export const getInFlightSignals = (params: { page?: number }) =>
  api<TInFlightSignalsResponse>({ path: 'in-flight-signals', params })

export const getSignalCatalog = () =>
  api<TSignalCatalogResponse>({ path: 'signal-catalog' })

export const getSignalCatalogDetail = (signalType: string) =>
  api<TSignalCatalogDetailResponse>({ path: `signal-catalog/${encodeURIComponent(signalType)}` })
