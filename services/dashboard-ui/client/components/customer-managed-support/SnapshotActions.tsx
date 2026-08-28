import { useParams } from 'react-router'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { KeyValueList } from '@/components/common/KeyValueList'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  InstallActionsTableComponent,
  type InstallActionRow,
} from '@/components/actions/InstallActionsTable'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { objectToKeyValueArray } from '@/utils/data-utils'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'
import { CustomerManagedSnapshotRunHistory } from './SnapshotRunHistory'

const durationNanos = (startedAt?: string, finishedAt?: string) => {
  if (!startedAt || !finishedAt) return undefined
  return (Date.parse(finishedAt) - Date.parse(startedAt)) * 1e6
}

export const CustomerManagedSnapshotActions = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const { org } = useOrg()
  const { install } = useInstall()
  const data = snapshot?.snapshot
  const actions = (data?.catalog?.refs ?? []).filter(
    (reference) => reference.kind === 'action'
  )
  const rows: InstallActionRow[] = actions.map((action) => {
    const latestRun = (data?.runs ?? [])
      .filter((run) => run.ref_kind === 'action' && run.ref_id === action.id)
      .sort((a, b) => b.started_at.localeCompare(a.started_at))[0]
    return {
      actionId: action.id,
      actionName: action.name,
      actionRunDatetime: latestRun ? (
        <Text flex className="gap-2">
          <Icon variant="CalendarBlankIcon" />
          <Time
            time={latestRun.started_at}
            format="relative"
            variant="subtext"
          />
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      actionRunDuration: latestRun ? (
        <Text flex className="gap-2">
          <Icon variant="TimerIcon" />
          <Duration
            nanoseconds={durationNanos(
              latestRun.started_at,
              latestRun.finished_at
            )}
            variant="subtext"
          />
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      actionStatus: latestRun ? (
        <Status
          variant="badge"
          status={
            latestRun.status === 'finished' ? 'success' : latestRun.status
          }
        />
      ) : (
        <Icon variant="MinusIcon" />
      ),
      actionTrigger: latestRun ? (
        <Badge size="sm" theme="neutral">
          {latestRun.source}
        </Badge>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      labels: <Icon variant="MinusIcon" />,
      href: `/${org.id}/installs/${install.id}/actions/${action.id}`,
    }
  })

  return (
    <CustomerManagedSnapshotContent>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Actions
        </Text>
        <Text variant="subtext" theme="neutral">
          View action definitions and immutable run history captured from this
          install.
        </Text>
      </HeadingGroup>
      <InstallActionsTableComponent
        data={rows}
        pagination={{ hasNext: false, offset: 0, limit: rows.length }}
      />
    </CustomerManagedSnapshotContent>
  )
}

export const CustomerManagedSnapshotActionDetail = () => {
  const { actionId } = useParams()
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const { org } = useOrg()
  const { install } = useInstall()
  const data = snapshot?.snapshot
  const actionRef = data?.catalog?.refs.find(
    (reference) => reference.kind === 'action' && reference.id === actionId
  )
  const content = data?.active_bundle?.contents?.find(
    (item) => item.kind === 'action' && item.name === actionRef?.name
  )
  const definition = content?.action_definition
  const runs = (data?.runs ?? []).filter(
    (run) => run.ref_kind === 'action' && run.ref_id === actionId
  )
  const latestRun = runs
    .slice()
    .sort((a, b) => b.started_at.localeCompare(a.started_at))[0]

  return (
    <PageSection flush className="flex-1">
      <PageTitle title={`${actionRef?.name ?? 'Action'} | ${install.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/installs`, text: 'Installs' },
          { path: `/${org.id}/installs/${install.id}`, text: install.name },
          {
            path: `/${org.id}/installs/${install.id}/actions`,
            text: 'Actions',
          },
          {
            path: `/${org.id}/installs/${install.id}/actions/${actionId}`,
            text: actionRef?.name,
          },
        ]}
      />
      <CustomerManagedSnapshotContent>
        {!actionRef ? (
          <PageSection>
            <EmptyState
              variant="table"
              emptyTitle="Action not captured"
              emptyMessage="This action is not present in the selected support snapshot."
            />
          </PageSection>
        ) : (
          <div className="@container flex flex-col flex-1">
            <header className="p-6 border-b flex flex-col gap-6">
              <HeadingGroup>
                <BackLink className="mb-4" />
                <Text variant="h3" weight="strong">
                  {actionRef.name}
                </Text>
                <ID>{actionRef.id}</ID>
              </HeadingGroup>
              <PropertyGrid
                values={[
                  {
                    property: 'Last status',
                    value: latestRun ? (
                      <Status
                        variant="badge"
                        status={
                          latestRun.status === 'finished'
                            ? 'success'
                            : latestRun.status
                        }
                      />
                    ) : undefined,
                  },
                  {
                    property: 'Kube config',
                    value:
                      definition?.enable_kube_config === undefined
                        ? undefined
                        : definition.enable_kube_config
                          ? 'Enabled'
                          : 'Disabled',
                  },
                  {
                    property: 'Timeout',
                    value: definition?.timeout_nanos ? (
                      <Duration nanoseconds={definition.timeout_nanos} />
                    ) : undefined,
                  },
                  { property: 'Execution role', value: definition?.role },
                  {
                    property: 'Kubernetes context',
                    value: definition?.kubernetes_context_name,
                  },
                ]}
              />
            </header>

            <div className="grid grid-cols-1 @5xl:grid-cols-12 flex-1">
              <PageSection className="@5xl:col-span-8 flex flex-col gap-6">
                {definition?.triggers?.length ? (
                  <div className="flex flex-col gap-3">
                    <Text variant="base" weight="strong">
                      Triggers
                    </Text>
                    <div className="flex flex-wrap gap-2">
                      {definition.triggers.map((trigger, index) => (
                        <Badge key={`${trigger.type}-${index}`} theme="neutral">
                          {trigger.type}
                          {trigger.cron_schedule
                            ? ` · ${trigger.cron_schedule}`
                            : ''}
                          {trigger.component_name
                            ? ` · ${trigger.component_name}`
                            : ''}
                        </Badge>
                      ))}
                    </div>
                  </div>
                ) : null}

                <div className="flex flex-col gap-4">
                  <Text variant="base" weight="strong">
                    Steps
                  </Text>
                  {definition?.steps?.length ? (
                    definition.steps
                      .slice()
                      .sort((a, b) => (a.index ?? 0) - (b.index ?? 0))
                      .map((step, index) => (
                        <Expand
                          key={`${step.name}-${index}`}
                          className="border rounded-md"
                          heading={
                            <Text weight="strong">
                              {index + 1}. {step.name ?? 'Step'}
                            </Text>
                          }
                          id={`action-step-${index}`}
                          isOpen
                        >
                          <div className="flex flex-col gap-4 p-4 border-t">
                            {step.command ? (
                              <CodeBlock language="bash">
                                {step.command}
                              </CodeBlock>
                            ) : null}
                            {step.environment &&
                            Object.keys(step.environment).length ? (
                              <KeyValueList
                                values={objectToKeyValueArray(step.environment)}
                              />
                            ) : null}
                            {step.artifact_digest ? (
                              <Text variant="subtext" theme="neutral">
                                Artifact {step.artifact_digest}
                              </Text>
                            ) : null}
                          </div>
                        </Expand>
                      ))
                  ) : (
                    <EmptyState
                      variant="table"
                      emptyTitle="No steps captured"
                      emptyMessage="This bundle does not contain action step definitions."
                    />
                  )}
                </div>
              </PageSection>
              <PageSection className="@5xl:col-span-4 flex flex-col gap-4">
                <Text variant="base" weight="strong">
                  Run history
                </Text>
                <CustomerManagedSnapshotRunHistory runs={runs} kind="Action" />
              </PageSection>
            </div>
          </div>
        )}
      </CustomerManagedSnapshotContent>
    </PageSection>
  )
}
