import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Hash } from '@/components/common/Hash'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Timeline } from '@/components/common/Timeline'
import { ComponentType } from '@/components/components/ComponentType'
import { WorkflowTimelineItem } from '@/components/workflows/WorkflowTimeline/WorkflowTimelineItem'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type { TCustomerManagedSnapshotRun } from '@/lib/ctl-api/installs/customer-managed-support-snapshots'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import type { TComponentType } from '@/types'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'

type CapturedRun = TCustomerManagedSnapshotRun & { created_at: string }

type SandboxResource = {
  release: string
  namespace?: string
  kind?: string
  name?: string
  health?: string
}

export const CustomerManagedSnapshotSandbox = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const { org } = useOrg()
  const { install } = useInstall()
  const data = snapshot?.snapshot
  const sandbox = data?.active_bundle?.contents?.find(
    (content) => content.kind === 'sandbox'
  )
  const runs: CapturedRun[] = (data?.runs ?? [])
    .filter((run) =>
      run.steps.some(
        (step) => step.id === 'sandbox-plan' || step.id === 'sandbox-apply'
      )
    )
    .map((run) => ({ ...run, created_at: run.started_at }))
  const latestApply = runs
    .flatMap((run) => run.steps)
    .filter((step) => step.id === 'sandbox-apply')
    .sort((a, b) =>
      (b.finished_at ?? b.started_at ?? '').localeCompare(
        a.finished_at ?? a.started_at ?? ''
      )
    )[0]
  const resources: SandboxResource[] = (
    data?.health?.sandbox_releases ?? []
  ).flatMap((release) =>
    (release.resources ?? []).map((resource) => ({
      release: release.release_name,
      namespace: release.namespace,
      kind: resource.kind,
      name: resource.name,
      health: resource.health,
    }))
  )

  return (
    <PageSection>
      <PageTitle title={`Sandbox | ${install.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/installs`, text: 'Installs' },
          { path: `/${org.id}/installs/${install.id}`, text: install.name },
          {
            path: `/${org.id}/installs/${install.id}/sandbox`,
            text: 'Sandbox',
          },
        ]}
      />

      <CustomerManagedSnapshotContent>
        <div className="@container flex flex-col gap-6">
          <HeadingGroup>
            <Text variant="base" weight="strong">
              Sandbox details
            </Text>
            <Text variant="subtext" theme="neutral">
              View sandbox configuration and history captured in this support
              snapshot.
            </Text>
          </HeadingGroup>

          <div className="grid grid-cols-1 @5xl:grid-cols-12 gap-6">
            <div className="@5xl:col-span-8 flex flex-col gap-6 min-w-0">
              <Card>
                <div className="flex items-center justify-between gap-4">
                  <span className="flex items-center gap-2">
                    <ComponentType
                      type={
                        (sandbox?.name === 'terraform'
                          ? 'terraform_module'
                          : (sandbox?.name ??
                            'terraform_module')) as TComponentType
                      }
                      displayVariant="icon-only"
                      colorVariant="color"
                      iconSize="24"
                    />
                    <Text weight="strong">Current sandbox</Text>
                  </span>
                  <Status
                    variant="badge"
                    status={
                      latestApply?.status === 'finished'
                        ? 'active'
                        : (latestApply?.status ?? 'unknown')
                    }
                  />
                </div>
                <PropertyGrid
                  values={[
                    { property: 'Type', value: sandbox?.name },
                    { property: 'Source', value: sandbox?.detail },
                    {
                      property: 'Content digest',
                      value: sandbox?.digest ? (
                        <Hash hash={sandbox.digest} />
                      ) : undefined,
                    },
                    {
                      property: 'Config digest',
                      value: sandbox?.config_digest ? (
                        <Hash hash={sandbox.config_digest} />
                      ) : undefined,
                    },
                    {
                      property: 'Last applied',
                      value: latestApply?.finished_at ? (
                        <Time
                          time={latestApply.finished_at}
                          format="long-datetime"
                        />
                      ) : undefined,
                    },
                  ]}
                />
              </Card>

              <Card>
                <HeadingGroup>
                  <Text weight="strong">Sandbox resources</Text>
                  <Text variant="subtext" theme="neutral">
                    Release resources captured by the runner health report.
                  </Text>
                </HeadingGroup>
                <PropertyGrid<SandboxResource>
                  values={resources}
                  columns={[
                    { key: 'release', header: 'Release' },
                    { key: 'name', header: 'Name' },
                    { key: 'kind', header: 'Type' },
                    { key: 'namespace', header: 'Namespace' },
                    {
                      key: 'health',
                      header: 'Health',
                      render: (value) => (
                        <Status
                          variant="badge"
                          status={String(value ?? 'unknown')}
                        />
                      ),
                    },
                  ]}
                  emptyStateProps={{
                    variant: 'table',
                    emptyTitle: 'No sandbox resources captured',
                    emptyMessage:
                      'Sandbox resources will appear after the runner reports release health.',
                  }}
                />
              </Card>
            </div>

            <div className="@5xl:col-span-4 flex flex-col gap-4 min-w-0">
              <Text variant="base" weight="strong">
                Sandbox history
              </Text>
              {runs.length ? (
                <Timeline<CapturedRun>
                  events={runs}
                  getEventKey={(run) => run.run_id}
                  pagination={{ hasNext: false, offset: 0, limit: runs.length }}
                  renderEvent={(run) => (
                    <WorkflowTimelineItem
                      id={run.run_id}
                      title={run.ref_name}
                      status={
                        run.status === 'finished' ? 'success' : run.status
                      }
                      createdAt={run.started_at}
                      finishedAt={run.finished_at}
                      finished={!!run.finished_at}
                      createdBy={run.source}
                      titleBadges={
                        <Badge size="sm" theme="neutral">
                          Sandbox
                        </Badge>
                      }
                    />
                  )}
                />
              ) : (
                <EmptyState
                  variant="table"
                  emptyTitle="No sandbox history captured"
                  emptyMessage="Sandbox history will appear after the sandbox runs."
                />
              )}
            </div>
          </div>
        </div>
      </CustomerManagedSnapshotContent>
    </PageSection>
  )
}
