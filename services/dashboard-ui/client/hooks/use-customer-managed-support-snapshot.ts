import { useContext } from 'react'
import { CustomerManagedSupportSnapshotContext } from '@/providers/customer-managed-support-snapshot-provider'

export const useCustomerManagedSupportSnapshot = () => {
  const context = useContext(CustomerManagedSupportSnapshotContext)
  if (!context)
    throw new Error(
      'useCustomerManagedSupportSnapshot must be used within CustomerManagedSupportSnapshotProvider'
    )
  return context
}
