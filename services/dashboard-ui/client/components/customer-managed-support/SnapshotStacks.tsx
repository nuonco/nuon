import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Hash } from '@/components/common/Hash'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Timeline } from '@/components/common/Timeline'
import { WorkflowTimelineItem } from '@/components/workflows/WorkflowTimeline/WorkflowTimelineItem'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type {
  TCustomerManagedBundleContent,
  TCustomerManagedSnapshotRun,
} from '@/lib/ctl-api/installs/customer-managed-support-snapshots'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'

type CapturedRun = TCustomerManagedSnapshotRun & { created_at: string }

function stackStatus(runs: TCustomerManagedSnapshotRun[]): string {
  const step = runs
    .flatMap((run) => run.steps)
    .filter(
      (item) => item.id === 'install-stack' || item.kind === 'cloudformation'
    )
    .sort((a, b) =>
      (b.finished_at ?? b.started_at ?? '').localeCompare(
        a.finished_at ?? a.started_at ?? ''
      )
    )[0]
  return step?.status === 'finished' ? 'active' : (step?.status ?? 'unknown')
}

export const CustomerManagedSnapshotStacks = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const { org } = useOrg()
  const { install } = useInstall()
  const data = snapshot?.snapshot
  const registration = data?.registration
  const runs = data?.runs ?? []
  const history: CapturedRun[] = runs
    .filter((run) =>
      run.steps.some(
        (step) => step.id === 'install-stack' || step.kind === 'cloudformation'
      )
    )
    .map((run) => ({ ...run, created_at: run.started_at }))
  const assets = (data?.active_bundle?.contents ?? []).filter(
    (content) => content.kind === 'stack-asset'
  )

  return (
    <PageSection>
      <PageTitle title={`Stacks | ${install.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/installs`, text: 'Installs' },
          { path: `/${org.id}/installs/${install.id}`, text: install.name },
          {
            path: `/${org.id}/installs/${install.id}/stacks`,
            text: 'Stacks',
          },
        ]}
      />

      <CustomerManagedSnapshotContent>
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Install stacks
          </Text>
          <Text variant="subtext" theme="neutral">
            View stack details and history captured in this support snapshot.
          </Text>
        </HeadingGroup>

        <Card>
          <div className="flex items-center justify-between gap-4">
            <HeadingGroup>
              <Text weight="strong">Current stack</Text>
              {registration?.stack.id ? <ID>{registration.stack.id}</ID> : null}
            </HeadingGroup>
            <Status variant="badge" status={stackStatus(runs)} />
          </div>
          <PropertyGrid
            values={[
              { property: 'Name', value: registration?.stack.name },
              { property: 'Type', value: registration?.stack.type },
              {
                property: 'Cloud provider',
                value: registration?.cloud.provider,
              },
              { property: 'Account ID', value: registration?.cloud.account_id },
              { property: 'Region', value: registration?.cloud.region },
              {
                property: 'Installed',
                value: registration?.installed_at ? (
                  <Time
                    time={registration.installed_at}
                    format="long-datetime"
                  />
                ) : undefined,
              },
            ]}
          />
        </Card>

        <Card>
          <HeadingGroup>
            <Text weight="strong">Stack assets</Text>
            <Text variant="subtext" theme="neutral">
              Templates and binaries captured in the active bundle.
            </Text>
          </HeadingGroup>
          <PropertyGrid<TCustomerManagedBundleContent>
            values={assets}
            columns={[
              { key: 'name', header: 'Name' },
              {
                key: 'detail',
                header: 'Source',
                render: (value) =>
                  typeof value === 'string' && value.startsWith('http') ? (
                    <Link href={value} isExternal className="break-words">
                      {value}
                    </Link>
                  ) : (
                    <Text
                      variant="subtext"
                      family="mono"
                      className="break-words"
                    >
                      {String(value ?? '—')}
                    </Text>
                  ),
              },
              {
                key: 'digest',
                header: 'Digest',
                render: (value) =>
                  value ? <Hash hash={String(value)} /> : undefined,
              },
            ]}
            emptyStateProps={{
              variant: 'table',
              emptyTitle: 'No stack assets captured',
              emptyMessage:
                'Stack assets will appear when they are included in the active bundle.',
            }}
          />
        </Card>

        <div className="flex flex-col gap-4">
          <Text weight="strong">Stack history</Text>
          {history.length ? (
            <Timeline<CapturedRun>
              events={history}
              getEventKey={(run) => run.run_id}
              pagination={{ hasNext: false, offset: 0, limit: history.length }}
              renderEvent={(run) => (
                <WorkflowTimelineItem
                  id={run.run_id}
                  title={run.ref_name}
                  status={run.status === 'finished' ? 'success' : run.status}
                  createdAt={run.started_at}
                  finishedAt={run.finished_at}
                  finished={!!run.finished_at}
                  createdBy={run.source}
                  titleBadges={
                    <Badge size="sm" theme="neutral">
                      Stack
                    </Badge>
                  }
                />
              )}
            />
          ) : (
            <EmptyState
              variant="table"
              emptyTitle="No stack history captured"
              emptyMessage="Stack history will appear after the install stack runs."
            />
          )}
        </div>
      </CustomerManagedSnapshotContent>
    </PageSection>
  )
}
