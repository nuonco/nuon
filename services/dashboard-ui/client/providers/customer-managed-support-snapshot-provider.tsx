import { createContext, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import {
  getCustomerManagedSupportSnapshot,
  getCustomerManagedSupportSnapshots,
} from '@/lib'
import type {
  TCustomerManagedSupportSnapshot,
  TCustomerManagedSupportSnapshotSummary,
} from '@/lib'
import type { TAPIError } from '@/types'
import { isCustomerManagedInstall } from '@/utils/install-utils'

type CustomerManagedSupportSnapshotContextValue = {
  snapshots: TCustomerManagedSupportSnapshotSummary[]
  snapshot?: TCustomerManagedSupportSnapshot
  isLoading: boolean
  error: TAPIError | null
}

export const CustomerManagedSupportSnapshotContext = createContext<
  CustomerManagedSupportSnapshotContextValue | undefined
>(undefined)

export const CustomerManagedSupportSnapshotProvider = ({
  children,
}: {
  children: ReactNode
}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()
  const isCustomerManaged = isCustomerManagedInstall(install)
  const list = useQuery({
    queryKey: ['customer-managed-support-snapshots', org.id, install.id],
    queryFn: () =>
      getCustomerManagedSupportSnapshots({
        orgId: org.id,
        installId: install.id,
      }),
    enabled: isCustomerManaged,
  })
  const requestedId = searchParams.get('snapshot')
  const snapshotId =
    requestedId && list.data?.some(({ id }) => id === requestedId)
      ? requestedId
      : list.data?.[0]?.id
  const detail = useQuery({
    queryKey: [
      'customer-managed-support-snapshot',
      org.id,
      install.id,
      snapshotId,
    ],
    queryFn: () =>
      getCustomerManagedSupportSnapshot({
        orgId: org.id,
        installId: install.id,
        snapshotId: snapshotId!,
      }),
    enabled: isCustomerManaged && !!snapshotId,
  })

  return (
    <CustomerManagedSupportSnapshotContext.Provider
      value={{
        snapshots: list.data ?? [],
        snapshot: detail.data,
        isLoading: list.isLoading || (!!snapshotId && detail.isLoading),
        error: (list.error || detail.error) as TAPIError | null,
      }}
    >
      {children}
    </CustomerManagedSupportSnapshotContext.Provider>
  )
}
