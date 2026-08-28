import { EmptyState } from '@/components/common/EmptyState'
import { Loading } from '@/components/common/Loading'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'

export const CustomerManagedSnapshotContent = ({
  children,
}: {
  children: React.ReactNode
}) => {
  const { snapshot, isLoading } = useCustomerManagedSupportSnapshot()
  if (isLoading) return <Loading />
  if (!snapshot) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No support snapshot yet"
        emptyMessage="Upload a support snapshot from the customer portal to view captured install data."
      />
    )
  }
  return children
}
