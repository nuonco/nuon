import { useParams } from 'react-router'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Hash } from '@/components/common/Hash'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { JSONViewer } from '@/components/common/JSONViewer'
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

type CapturedResource = {
  kind?: string
  name?: string
  namespace?: string
  provider?: string
  health?: string
  message?: string
}

type CapturedRun = TCustomerManagedSnapshotRun & { created_at: string }

function runStatus(status: string): string {
  return status === 'finished' ? 'success' : status
}

export const CustomerManagedSnapshotComponentDetail = () => {
  const { componentId } = useParams()
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const { org } = useOrg()
  const { install } = useInstall()
  const data = snapshot?.snapshot
  const health = data?.health?.components?.find(
    (component) =>
      component.component_id === componentId ||
      component.install_component_id === componentId ||
      component.component_name === componentId
  )
  const component = data?.active_bundle?.contents?.find(
    (content) =>
      content.kind === 'component' &&
      (content.name === componentId || content.name === health?.component_name)
  )
  const componentName = component?.name ?? health?.component_name
  const runs: CapturedRun[] = (data?.runs ?? [])
    .filter(
      (run) =>
        run.ref_kind !== 'drift' &&
        run.steps.some(
          (step) =>
            !!componentName &&
            (step.id === `sync-${componentName}` ||
              step.id === `deploy-${componentName}-plan` ||
              step.id === `deploy-${componentName}-apply`)
        )
    )
    .map((run) => ({ ...run, created_at: run.started_at }))

  return (
    <PageSection>
      <PageTitle title={`${componentName ?? 'Component'} | ${install.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/installs`, text: 'Installs' },
          { path: `/${org.id}/installs/${install.id}`, text: install.name },
          {
            path: `/${org.id}/installs/${install.id}/components`,
            text: 'Components',
          },
          {
            path: `/${org.id}/installs/${install.id}/components/${componentId}`,
            text: componentName,
          },
        ]}
      />

      <CustomerManagedSnapshotContent>
        {!component ? (
          <EmptyState
            variant="table"
            emptyTitle="Component not captured"
            emptyMessage="Return to Components to view the component inventory in this snapshot."
          />
        ) : (
          <div className="@container flex flex-col flex-auto gap-6">
            <HeadingGroup>
              <BackLink className="mb-6" />
              <span className="flex items-center gap-2">
                <ComponentType
                  type={
                    (health?.component_type ??
                      component.detail) as TComponentType
                  }
                  displayVariant="icon-only"
                  colorVariant="color"
                  iconSize="24"
                />
                <Text variant="base" weight="strong">
                  {component.name}
                </Text>
                <Status variant="badge" status={health?.health ?? 'unknown'} />
              </span>
              {health?.component_id ? <ID>{health.component_id}</ID> : null}
            </HeadingGroup>

            <div className="grid grid-cols-1 @5xl:grid-cols-12 gap-6">
              <div className="@5xl:col-span-8 flex flex-col gap-6">
                <Card>
                  <Text variant="base" weight="strong">
                    Captured component
                  </Text>
                  <PropertyGrid
                    values={[
                      {
                        property: 'Type',
                        value: health?.component_type ?? component.detail,
                      },
                      {
                        property: 'Content digest',
                        value: component.digest ? (
                          <Hash hash={component.digest} />
                        ) : undefined,
                      },
                      {
                        property: 'Config digest',
                        value: component.config_digest ? (
                          <Hash hash={component.config_digest} />
                        ) : undefined,
                      },
                      {
                        property: 'Bundle activated',
                        value: data?.active_bundle?.activated_at ? (
                          <Time
                            time={data.active_bundle.activated_at}
                            format="long-datetime"
                          />
                        ) : undefined,
                      },
                      {
                        property: 'Health observed',
                        value: data?.health?.observed_at ? (
                          <Time
                            time={data.health.observed_at}
                            format="long-datetime"
                          />
                        ) : undefined,
                      },
                    ]}
                  />
                </Card>

                <Card>
                  <HeadingGroup>
                    <Text variant="base" weight="strong">
                      Configuration
                    </Text>
                    <Text variant="subtext" theme="neutral">
                      Configuration captured in the active bundle.
                    </Text>
                  </HeadingGroup>
                  <JSONViewer
                    data={component.component_definition ?? {}}
                    expanded={2}
                  />
                </Card>

                <Card>
                  <HeadingGroup>
                    <Text variant="base" weight="strong">
                      Resources
                    </Text>
                    <Text variant="subtext" theme="neutral">
                      Resource health captured by the runner.
                    </Text>
                  </HeadingGroup>
                  <PropertyGrid<CapturedResource>
                    values={health?.resources ?? []}
                    columns={[
                      { key: 'name', header: 'Name' },
                      { key: 'kind', header: 'Type' },
                      { key: 'provider', header: 'Provider' },
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
                      emptyTitle: 'No resources captured',
                      emptyMessage:
                        'Resource health will appear after the runner reports component health.',
                    }}
                  />
                  {health?.truncated ? (
                    <Badge theme="warn">Resource list truncated</Badge>
                  ) : null}
                </Card>
              </div>

              <div className="@5xl:col-span-4 flex flex-col gap-4">
                <Text variant="base" weight="strong">
                  Deploy history
                </Text>
                {runs.length ? (
                  <Timeline<CapturedRun>
                    events={runs}
                    getEventKey={(run) => run.run_id}
                    pagination={{
                      hasNext: false,
                      offset: 0,
                      limit: runs.length,
                    }}
                    renderEvent={(run) => (
                      <WorkflowTimelineItem
                        id={run.run_id}
                        title={run.ref_name}
                        status={runStatus(run.status)}
                        createdAt={run.started_at}
                        finishedAt={run.finished_at}
                        finished={!!run.finished_at}
                        createdBy={run.source}
                        titleBadges={
                          <Badge size="sm" theme="neutral">
                            {run.ref_kind}
                          </Badge>
                        }
                      />
                    )}
                  />
                ) : (
                  <EmptyState
                    variant="table"
                    emptyTitle="No deploy history captured"
                    emptyMessage="Deploy history will appear after this component runs."
                  />
                )}
              </div>
            </div>
          </div>
        )}
      </CustomerManagedSnapshotContent>
    </PageSection>
  )
}
