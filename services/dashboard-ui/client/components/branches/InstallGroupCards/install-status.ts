import {
  stepStatusCategory,
  type TStepStatusCategory,
} from '@/components/branches/shared/step-status'
import { statusAccent } from '@/components/branches/graph/accents'

export type TInstallStatusSlice = {
  status: string
  label: string
  count: number
}

export type TStatusInstall = {
  id: string
  name: string
  status: string
}

const SLICE_ORDER: { status: string; label: string }[] = [
  { status: 'success', label: 'success' },
  { status: 'error', label: 'failed' },
  { status: 'in-progress', label: 'in progress' },
  { status: 'approval-awaiting', label: 'awaiting approval' },
  { status: 'cancelled', label: 'cancelled' },
  { status: 'pending', label: 'pending' },
]

const CATEGORY_ORDER: TStepStatusCategory[] = [
  'success',
  'error',
  'active',
  'awaiting',
  'pending',
]

export const installStatusBucket = (status?: string) => {
  if (!status || status === 'cancelled' || status.endsWith('skipped')) {
    return status?.endsWith('skipped') ? 'cancelled' : status || 'pending'
  }
  switch (stepStatusCategory(status)) {
    case 'success':
      return 'success'
    case 'error':
      return 'error'
    case 'active':
      return 'in-progress'
    case 'awaiting':
      return 'approval-awaiting'
    default:
      return 'pending'
  }
}

export const installStatusClass = (status: string) =>
  installStatusBucket(status) === 'pending'
    ? 'bg-black/10 dark:bg-white/15'
    : statusAccent(installStatusBucket(status)).dot

export const statusSlices = (
  installs: { status: string }[]
): TInstallStatusSlice[] => {
  const counts = new Map<string, number>()
  for (const install of installs) {
    const key = installStatusBucket(install.status)
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }
  return SLICE_ORDER.flatMap((slice) => {
    const count = counts.get(slice.status) ?? 0
    return count ? [{ ...slice, count }] : []
  })
}

export const sortedStatusInstalls = <T extends { status: string }>(
  installs: T[]
) =>
  [...installs].sort(
    (a, b) =>
      CATEGORY_ORDER.indexOf(
        stepStatusCategory(installStatusBucket(a.status))
      ) -
      CATEGORY_ORDER.indexOf(stepStatusCategory(installStatusBucket(b.status)))
  )

export const withQueuedInstalls = <T extends TStatusInstall>(
  installs: T[],
  plannedCount?: number
): TStatusInstall[] => {
  const queued = Math.max((plannedCount ?? 0) - installs.length, 0)
  return [
    ...installs,
    ...Array.from({ length: queued }, (_, index) => ({
      id: `queued-${index}`,
      name: 'Queued',
      status: 'pending',
    })),
  ]
}

export const filterInstalls = <T extends { name: string; status: string }>(
  installs: T[],
  { query, status }: { query: string; status?: string }
) => {
  const needle = query.trim().toLowerCase()
  return installs.filter(
    (install) =>
      (!needle || install.name.toLowerCase().includes(needle)) &&
      (!status || install.status === status)
  )
}
