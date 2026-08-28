import { Banner } from '@/components/common/Banner'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'

export const CustomerManagedSnapshotBanner = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { snapshot, isLoading, error } = useCustomerManagedSupportSnapshot()
  const supportPath = `/${org.id}/installs/${install.id}/support`

  if (isLoading)
    return <Banner theme="info">Loading the latest support snapshot.</Banner>
  if (error) {
    return (
      <Banner theme="error">
        <div className="flex flex-col gap-1">
          <Text weight="strong">Support snapshot loading failed</Text>
          <Text variant="subtext">{error.description || error.error}</Text>
        </div>
      </Banner>
    )
  }
  if (!snapshot) {
    return (
      <Banner theme="warn">
        <Text weight="strong">No support snapshot uploaded</Text>
        <Text variant="subtext">
          <Link href={supportPath}>Upload a support snapshot</Link> to view this
          offline customer-managed install.
        </Text>
      </Banner>
    )
  }

  return (
    <Banner theme="info">
      <div className="flex flex-col gap-1">
        <Text weight="strong">Viewing captured data — not live</Text>
        <Text variant="subtext">
          Captured <Time time={snapshot.captured_at} format="long-datetime" />.{' '}
          <Link href={`${supportPath}?snapshot=${snapshot.id}`}>
            View snapshot details
          </Link>
        </Text>
      </div>
    </Banner>
  )
}
