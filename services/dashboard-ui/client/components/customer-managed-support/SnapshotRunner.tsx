import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { Hash } from '@/components/common/Hash'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'
import { CustomerManagedSnapshotRunHistory } from './SnapshotRunHistory'

export const CustomerManagedSnapshotRunner = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const data = snapshot?.snapshot
  const runner = data?.runner
  const runs = (data?.runs ?? [])
    .slice()
    .sort((a, b) => b.started_at.localeCompare(a.started_at))
    .slice(0, 10)

  return (
    <CustomerManagedSnapshotContent>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install runner
        </Text>
        <Text variant="subtext" theme="neutral">
          Runner identity and activity as observed when this support snapshot
          was captured.
        </Text>
      </HeadingGroup>

      {!runner ? (
        <EmptyState
          variant="table"
          emptyTitle="No runner heartbeat captured"
          emptyMessage="Runner details will appear after the customer captures a snapshot containing a heartbeat."
        />
      ) : (
        <div className="flex flex-col gap-6">
          <Card>
            <div className="flex flex-wrap items-start justify-between gap-4">
              <HeadingGroup>
                <span className="flex items-center gap-2">
                  <Text weight="strong">Captured heartbeat</Text>
                  <Status variant="badge" status="success">
                    Reported
                  </Status>
                </span>
                <ID>{runner.runner_id}</ID>
              </HeadingGroup>
              <span className="flex items-center gap-2">
                <Text variant="subtext" theme="neutral">
                  Observed
                </Text>
                <Time
                  time={runner.observed_at}
                  format="relative"
                  variant="subtext"
                />
              </span>
            </div>
            <PropertyGrid
              values={[
                { property: 'Session ID', value: <ID>{runner.session_id}</ID> },
                { property: 'Version', value: runner.version },
                {
                  property: 'Bundle digest',
                  value: runner.bundle_digest ? (
                    <Hash hash={runner.bundle_digest} />
                  ) : undefined,
                },
                {
                  property: 'Started',
                  value: (
                    <Time time={runner.started_at} format="long-datetime" />
                  ),
                },
                {
                  property: 'Uptime at capture',
                  value: (
                    <Duration
                      beginTime={runner.started_at}
                      endTime={runner.observed_at}
                    />
                  ),
                },
                {
                  property: 'Snapshot captured',
                  value: data?.captured_at ? (
                    <Time time={data.captured_at} format="long-datetime" />
                  ) : undefined,
                },
              ]}
            />
          </Card>

          <Card>
            <HeadingGroup>
              <Text weight="strong">Capabilities</Text>
              <Text variant="subtext" theme="neutral">
                Features advertised by this runner session.
              </Text>
            </HeadingGroup>
            {runner.capabilities?.length ? (
              <div className="flex flex-wrap gap-2">
                {runner.capabilities.map((capability) => (
                  <Badge key={capability} variant="code" theme="neutral">
                    {capability}
                  </Badge>
                ))}
              </div>
            ) : (
              <Text variant="subtext" theme="neutral">
                No capabilities were advertised.
              </Text>
            )}
          </Card>

          <div className="flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Recent jobs
            </Text>
            <CustomerManagedSnapshotRunHistory runs={runs} kind="Run" />
          </div>
        </div>
      )}
    </CustomerManagedSnapshotContent>
  )
}
