import { Card } from '@/components/common/Card'
import { Hash } from '@/components/common/Hash'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'

export const CustomerManagedSnapshotOverview = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const data = snapshot?.snapshot
  const registration = data?.registration
  const sandbox = data?.active_bundle?.contents?.find(
    ({ kind }) => kind === 'sandbox'
  )
  const installationRun = data?.runs?.find(
    ({ ref_kind }) => ref_kind === 'install' || ref_kind === 'upgrade'
  )

  return (
    <CustomerManagedSnapshotContent>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Install overview
        </Text>
        <Text variant="subtext" theme="neutral">
          View the captured stack, sandbox, runner, and bundle state.
        </Text>
      </HeadingGroup>
      <div className="grid grid-cols-1 gap-6 xl:grid-cols-2">
        <Card>
          <Text variant="base" weight="strong">
            Installation
          </Text>
          <div className="grid grid-cols-2 gap-4">
            <LabeledValue label="Status">
              <Status status={installationRun?.status ?? 'unknown'} />
            </LabeledValue>
            <LabeledValue label="Installed">
              <Time time={registration?.installed_at} format="long-datetime" />
            </LabeledValue>
            <LabeledValue label="Deployment">
              <ID>{registration?.deployment_id}</ID>
            </LabeledValue>
            <LabeledValue label="Customer install">
              <ID>{registration?.install_id}</ID>
            </LabeledValue>
          </div>
        </Card>
        <Card>
          <Text variant="base" weight="strong">
            Cloud stack
          </Text>
          <div className="grid grid-cols-2 gap-4">
            <LabeledValue label="Provider">
              {registration?.cloud.provider ?? '—'}
            </LabeledValue>
            <LabeledValue label="Region">
              {registration?.cloud.region ?? '—'}
            </LabeledValue>
            <LabeledValue label="Account">
              <ID>{registration?.cloud.account_id}</ID>
            </LabeledValue>
            <LabeledValue label="Stack">
              {registration?.stack.name ?? '—'}
            </LabeledValue>
          </div>
        </Card>
        <Card>
          <Text variant="base" weight="strong">
            Sandbox
          </Text>
          {sandbox ? (
            <div className="grid grid-cols-2 gap-4">
              <LabeledValue label="Type">{sandbox.name}</LabeledValue>
              <LabeledValue label="Source">
                {sandbox.detail ?? '—'}
              </LabeledValue>
              <LabeledValue label="Content">
                <Hash hash={sandbox.digest ?? ''} />
              </LabeledValue>
              <LabeledValue label="Config">
                <Hash hash={sandbox.config_digest ?? ''} />
              </LabeledValue>
            </div>
          ) : (
            <Text theme="neutral">No sandbox was included in this bundle.</Text>
          )}
        </Card>
        <Card>
          <Text variant="base" weight="strong">
            Runner
          </Text>
          {data?.runner ? (
            <div className="grid grid-cols-2 gap-4">
              <LabeledValue label="Runner">
                <ID>{data.runner.runner_id}</ID>
              </LabeledValue>
              <LabeledValue label="Version">
                {data.runner.version || '—'}
              </LabeledValue>
              <LabeledValue label="Last observed">
                <Time time={data.runner.observed_at} format="long-datetime" />
              </LabeledValue>
              <LabeledValue label="Bundle">
                <Hash hash={data.runner.bundle_digest} />
              </LabeledValue>
            </div>
          ) : (
            <Text theme="neutral">
              Runner heartbeat was unavailable when this snapshot was captured.
            </Text>
          )}
        </Card>
      </div>
    </CustomerManagedSnapshotContent>
  )
}
