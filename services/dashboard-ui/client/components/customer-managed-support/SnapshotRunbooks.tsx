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
import { Link } from '@/components/common/Link'
import { PropertyGrid } from '@/components/common/PropertyGrid'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  RunbooksTable,
  type TRunbookRow,
} from '@/components/runbooks/RunbooksTable/RunbooksTable'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useCustomerManagedSupportSnapshot } from '@/hooks/use-customer-managed-support-snapshot'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { objectToKeyValueArray } from '@/utils/data-utils'
import { CustomerManagedSnapshotContent } from './SnapshotEmpty'
import { CustomerManagedSnapshotRunHistory } from './SnapshotRunHistory'

export const CustomerManagedSnapshotRunbooks = () => {
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const { org } = useOrg()
  const { install } = useInstall()
  const data = snapshot?.snapshot
  const runbooks = (data?.catalog?.refs ?? []).filter(
    (reference) => reference.kind === 'runbook'
  )
  const rows: TRunbookRow[] = runbooks.map((runbook) => {
    const href = `/${org.id}/installs/${install.id}/runbooks/${runbook.id}`
    const content = data?.active_bundle?.contents?.find(
      (item) => item.kind === 'runbook' && item.name === runbook.name
    )
    const latestRun = (data?.runs ?? [])
      .filter((run) => run.ref_kind === 'runbook' && run.ref_id === runbook.id)
      .sort((a, b) => b.started_at.localeCompare(a.started_at))[0]
    return {
      runbookId: runbook.id,
      runbookName: runbook.name,
      description: content?.detail ? (
        <Text variant="subtext" theme="neutral">
          {content.detail}
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      labels: <Icon variant="MinusIcon" />,
      lastUpdated: data?.active_bundle?.activated_at ? (
        <Text flex nowrap className="gap-2">
          <Icon variant="CalendarBlankIcon" />
          <Time
            time={data.active_bundle.activated_at}
            format="relative"
            variant="subtext"
          />
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      lastRun: latestRun ? (
        <Text flex nowrap className="gap-2">
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
      href,
      actions: null,
    }
  })

  return (
    <CustomerManagedSnapshotContent>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Runbooks
        </Text>
        <Text variant="subtext" theme="neutral">
          View operational procedures and immutable run history captured from
          this install.
        </Text>
      </HeadingGroup>
      <RunbooksTable
        data={rows}
        filterActions={null}
        isLoading={false}
        pagination={{ hasNext: false, offset: 0, limit: rows.length }}
        scope="install"
      />
    </CustomerManagedSnapshotContent>
  )
}

export const CustomerManagedSnapshotRunbookDetail = () => {
  const { runbookId } = useParams()
  const { snapshot } = useCustomerManagedSupportSnapshot()
  const { org } = useOrg()
  const { install } = useInstall()
  const data = snapshot?.snapshot
  const runbookRef = data?.catalog?.refs.find(
    (reference) => reference.kind === 'runbook' && reference.id === runbookId
  )
  const content = data?.active_bundle?.contents?.find(
    (item) => item.kind === 'runbook' && item.name === runbookRef?.name
  )
  const definition = content?.runbook_definition
  const runs = (data?.runs ?? []).filter(
    (run) => run.ref_kind === 'runbook' && run.ref_id === runbookId
  )

  return (
    <PageSection flush className="flex-1">
      <PageTitle title={`${runbookRef?.name ?? 'Runbook'} | ${install.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org.id}`, text: org.name },
          { path: `/${org.id}/installs`, text: 'Installs' },
          { path: `/${org.id}/installs/${install.id}`, text: install.name },
          {
            path: `/${org.id}/installs/${install.id}/runbooks`,
            text: 'Runbooks',
          },
          {
            path: `/${org.id}/installs/${install.id}/runbooks/${runbookId}`,
            text: runbookRef?.name,
          },
        ]}
      />
      <CustomerManagedSnapshotContent>
        {!runbookRef ? (
          <PageSection>
            <EmptyState
              variant="table"
              emptyTitle="Runbook not captured"
              emptyMessage="This runbook is not present in the selected support snapshot."
            />
          </PageSection>
        ) : (
          <div className="@container flex flex-col flex-1">
            <header className="p-6 border-b flex flex-col gap-6">
              <HeadingGroup>
                <BackLink className="mb-4" />
                <Text variant="h3" weight="strong">
                  {runbookRef.name}
                </Text>
                {content?.detail ? (
                  <Text variant="subtext">{content.detail}</Text>
                ) : null}
                <ID>{runbookRef.id}</ID>
              </HeadingGroup>
              <PropertyGrid
                values={[
                  {
                    property: 'Steps',
                    value: definition?.steps?.length ?? runbookRef.steps,
                  },
                  {
                    property: 'Inputs',
                    value: definition?.inputs?.length,
                  },
                  {
                    property: 'Readme digest',
                    value: definition?.readme_digest,
                  },
                ]}
              />
            </header>

            <div className="grid grid-cols-1 @5xl:grid-cols-12 flex-1">
              <PageSection className="@5xl:col-span-8 flex flex-col gap-6">
                {definition?.inputs?.length ? (
                  <div className="flex flex-col gap-4">
                    <Text variant="base" weight="strong">
                      Inputs
                    </Text>
                    {definition.inputs
                      .slice()
                      .sort((a, b) => (a.index ?? 0) - (b.index ?? 0))
                      .map((input) => (
                        <Expand
                          key={input.name}
                          className="border rounded-md"
                          heading={
                            <span className="flex items-center gap-2">
                              <Text weight="strong">
                                {input.display_name ?? input.name}
                              </Text>
                              {input.required ? (
                                <Badge size="sm" theme="warn">
                                  Required
                                </Badge>
                              ) : null}
                              {input.sensitive ? (
                                <Badge size="sm" theme="neutral">
                                  Sensitive
                                </Badge>
                              ) : null}
                            </span>
                          }
                          id={`runbook-input-${input.name}`}
                        >
                          <div className="p-4 border-t">
                            <PropertyGrid
                              values={[
                                { property: 'Name', value: input.name },
                                { property: 'Type', value: input.type },
                                {
                                  property: 'Description',
                                  value: input.description,
                                },
                                {
                                  property: 'Default',
                                  value: input.sensitive
                                    ? 'Redacted'
                                    : input.default === undefined
                                      ? undefined
                                      : JSON.stringify(input.default),
                                },
                              ]}
                            />
                          </div>
                        </Expand>
                      ))}
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
                            <span className="flex items-center gap-2">
                              <Text weight="strong">
                                {index + 1}. {step.name ?? step.kind ?? 'Step'}
                              </Text>
                              {step.kind ? (
                                <Badge size="sm" theme="neutral">
                                  {step.kind}
                                </Badge>
                              ) : null}
                            </span>
                          }
                          id={`runbook-step-${index}`}
                          isOpen
                        >
                          <div className="flex flex-col gap-4 p-4 border-t">
                            <PropertyGrid
                              values={[
                                {
                                  property: 'Component',
                                  value: step.component,
                                },
                                {
                                  property: 'Reference',
                                  value: step.reference,
                                },
                                { property: 'Role', value: step.role },
                                {
                                  property: 'Timeout',
                                  value: step.timeout_nanos ? (
                                    <Duration
                                      nanoseconds={step.timeout_nanos}
                                    />
                                  ) : undefined,
                                },
                                {
                                  property: 'Plan only',
                                  value: step.plan_only ? 'Yes' : undefined,
                                },
                                {
                                  property: 'Trigger',
                                  value: step.trigger_name,
                                },
                                {
                                  property: 'Event types',
                                  value: step.event_types?.join(', '),
                                },
                              ]}
                            />
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
                          </div>
                        </Expand>
                      ))
                  ) : (
                    <EmptyState
                      variant="table"
                      emptyTitle="No steps captured"
                      emptyMessage="This bundle does not contain runbook step definitions."
                    />
                  )}
                </div>
              </PageSection>
              <PageSection className="@5xl:col-span-4 flex flex-col gap-4">
                <Text variant="base" weight="strong">
                  Run history
                </Text>
                <CustomerManagedSnapshotRunHistory runs={runs} kind="Runbook" />
              </PageSection>
            </div>
          </div>
        )}
      </CustomerManagedSnapshotContent>
    </PageSection>
  )
}
