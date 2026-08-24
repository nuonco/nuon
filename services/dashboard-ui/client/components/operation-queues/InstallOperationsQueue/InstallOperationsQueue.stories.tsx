export default {
  title: 'Operation queues/Install operations queue',
}

import { useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Divider } from '@/components/common/Divider'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import {
  InstallRunbooksTableComponent,
  type TInstallRunbookRow,
} from '@/components/runbooks/InstallRunbooksTable'
import {
  AddQueueOperationStudio,
  InstallOperationsQueue,
  QueueAdmissionRulesStudio,
  type IOperationQueueEntry,
  type IQueueAdmissionExemptionOption,
  type IQueueOperationOption,
  type TQueueAdmissionPolicy,
  type TScheduledOperationsPolicy,
  type TOperationQueueStatus,
} from './InstallOperationsQueue'
import {
  installQueueReorderSchema,
  type InstallQueueReorderValues,
} from './schema'

const initialEntries: IOperationQueueEntry[] = [
  {
    id: 'operation-backup',
    name: 'Back up database',
    type: 'runbook',
    status: 'success',
    detail: 'Runbook · Completed 8 minutes ago',
    required: true,
  },
  {
    id: 'operation-rotate',
    name: 'Rotate database credentials',
    type: 'runbook',
    status: 'in-progress',
    detail: 'Runbook · Started by an operator',
    required: true,
  },
  {
    id: 'operation-deploy',
    name: 'Deploy application',
    type: 'deploy',
    status: 'blocked',
    detail: 'Deploy all components from the approved build',
    blockedBy: 'Rotate database credentials',
  },
  {
    id: 'operation-verify',
    name: 'Verify service health',
    type: 'action',
    status: 'queued',
    detail: 'Action · Run health and connectivity checks',
  },
]

const rules = [
  {
    id: 'rule-admission',
    name: 'Mutations use this queue',
    description: 'Runbooks, deploys, and actions must be submitted here.',
  },
  {
    id: 'rule-deploy',
    name: 'Credentials before deploy',
    description: 'The deploy remains blocked until credential rotation passes.',
  },
  {
    id: 'rule-concurrency',
    name: 'One operation at a time',
    description:
      'The next eligible operation starts after the current one finishes.',
  },
]

const operationOptions: IQueueOperationOption[] = [
  {
    id: 'restart-workers',
    name: 'Restart workers',
    type: 'runbook',
    detail: 'Runbook · Restart workers and verify queue processing',
  },
  {
    id: 'deploy-approved-build',
    name: 'Deploy approved build',
    type: 'deploy',
    detail: 'Deploy · Roll out the approved build to all components',
  },
  {
    id: 'verify-connectivity',
    name: 'Verify connectivity',
    type: 'action',
    detail: 'Action · Check service and dependency connectivity',
  },
]

const exemptionOptions: IQueueAdmissionExemptionOption[] = [
  {
    id: 'runbook-independent',
    operationType: 'runbook',
    labelKey: 'operations',
    labelValue: 'independent',
    matchingOperations: ['Restart stateless workers', 'Verify service health'],
  },
  {
    id: 'component-deploy-canary',
    operationType: 'component deploy',
    labelKey: 'environment',
    labelValue: 'canary',
    matchingOperations: ['Deploy metrics collector'],
  },
  {
    id: 'sandbox-read-only',
    operationType: 'sandbox operation',
    labelKey: 'risk',
    labelValue: 'read-only',
    matchingOperations: ['Check sandbox drift'],
  },
  {
    id: 'action-routine',
    operationType: 'action',
    labelKey: 'operations',
    labelValue: 'routine',
    matchingOperations: ['Verify connectivity', 'Flush cache'],
  },
]

export const Interactive = () => {
  const [status, setStatus] = useState<TOperationQueueStatus>('running')
  const [entries, setEntries] = useState(initialEntries)
  const [addingOperation, setAddingOperation] = useState(false)
  const [configuringRules, setConfiguringRules] = useState(false)
  const [selectedOptionId, setSelectedOptionId] = useState('')
  const [required, setRequired] = useState(false)
  const [admissionPolicy, setAdmissionPolicy] =
    useState<TQueueAdmissionPolicy>('enqueue')
  const [scheduledOperationsPolicy, setScheduledOperationsPolicy] =
    useState<TScheduledOperationsPolicy>('enqueue')
  const [selectedExemptionIds, setSelectedExemptionIds] = useState<string[]>([])

  const addOperation = () => {
    const selectedOption = operationOptions.find(
      (option) => option.id === selectedOptionId
    )
    if (!selectedOption) return

    setEntries((current) =>
      current.some((entry) => entry.id === `operation-${selectedOption.id}`)
        ? current
        : [
            ...current,
            {
              ...selectedOption,
              id: `operation-${selectedOption.id}`,
              status: 'queued',
              required,
            },
          ]
    )
    setAddingOperation(false)
    setSelectedOptionId('')
    setRequired(false)
  }

  if (addingOperation) {
    return (
      <AddQueueOperationStudio
        installName="production-us-west"
        queueName="August maintenance"
        entries={entries}
        options={operationOptions}
        selectedOptionId={selectedOptionId}
        required={required}
        onSelectOption={setSelectedOptionId}
        onRequiredChange={setRequired}
        onCancel={() => {
          setAddingOperation(false)
          setSelectedOptionId('')
          setRequired(false)
        }}
        onAdd={addOperation}
      />
    )
  }

  if (configuringRules) {
    return (
      <QueueAdmissionRulesStudio
        installName="production-us-west"
        queueName="August maintenance"
        admissionPolicy={admissionPolicy}
        scheduledOperationsPolicy={scheduledOperationsPolicy}
        exemptionOptions={exemptionOptions}
        selectedExemptionIds={selectedExemptionIds}
        onCancel={() => setConfiguringRules(false)}
        onSave={(nextRules) => {
          setAdmissionPolicy(nextRules.admissionPolicy)
          setScheduledOperationsPolicy(nextRules.scheduledOperationsPolicy)
          setSelectedExemptionIds(nextRules.selectedExemptionIds)
          setConfiguringRules(false)
        }}
      />
    )
  }

  const configuredRules =
    admissionPolicy === 'strict'
      ? [
          {
            id: 'rule-strict-admission',
            name: 'Strict queue admission',
            description:
              'Blocks every operation that was not submitted through this queue.',
          },
          rules[2],
        ]
      : [
          rules[0],
          {
            id: 'rule-scheduled',
            name: 'Scheduled operations',
            description:
              scheduledOperationsPolicy === 'direct'
                ? 'Cron-triggered operations run directly.'
                : 'Cron-triggered operations join this queue.',
          },
          ...exemptionOptions
            .filter((exemption) => selectedExemptionIds.includes(exemption.id))
            .map((exemption) => ({
              id: `rule-operation-exemption-${exemption.id}`,
              name: 'Operation exemption',
              description: `Allows ${exemption.operationType} operations labeled ${exemption.labelKey}=${exemption.labelValue} to run directly.`,
            })),
          rules[2],
        ]

  return (
    <InstallOperationsQueue
      installName="production-us-west"
      queueName="August maintenance"
      status={status}
      entries={entries}
      rules={configuredRules}
      onAddOperation={() => setAddingOperation(true)}
      onConfigureRules={() => setConfiguringRules(true)}
      onStart={() => setStatus('running')}
      onRunOutsideQueue={() => setStatus('out-of-band')}
      onFinishOutsideOperation={() => setStatus('running')}
    />
  )
}

export const AdmissionRules = () => {
  return (
    <QueueAdmissionRulesStudio
      installName="production-us-west"
      queueName="August maintenance"
      admissionPolicy="enqueue"
      scheduledOperationsPolicy="enqueue"
      exemptionOptions={exemptionOptions}
      selectedExemptionIds={exemptionOptions.map((exemption) => exemption.id)}
      onCancel={() => {}}
      onSave={() => {}}
    />
  )
}

export const AdmissionRulesWithoutExemptions = () => {
  return (
    <QueueAdmissionRulesStudio
      installName="production-us-west"
      queueName="August maintenance"
      admissionPolicy="enqueue"
      scheduledOperationsPolicy="enqueue"
      exemptionOptions={exemptionOptions}
      selectedExemptionIds={[]}
      onCancel={() => {}}
      onSave={() => {}}
    />
  )
}

export const AdmissionRulesStrictQueue = () => {
  return (
    <QueueAdmissionRulesStudio
      installName="production-us-west"
      queueName="August maintenance"
      admissionPolicy="strict"
      scheduledOperationsPolicy="direct"
      exemptionOptions={exemptionOptions}
      selectedExemptionIds={exemptionOptions.map((exemption) => exemption.id)}
      onCancel={() => {}}
      onSave={() => {}}
    />
  )
}

export const Draft = () => (
  <InstallOperationsQueue
    installName="production-us-west"
    queueName="August maintenance"
    status="draft"
    entries={initialEntries.map((entry) => ({
      ...entry,
      status: 'queued',
      blockedBy: undefined,
    }))}
    rules={rules}
  />
)

export const OutsideOperationActive = () => (
  <InstallOperationsQueue
    installName="production-us-west"
    queueName="August maintenance"
    status="out-of-band"
    entries={initialEntries}
    rules={rules}
  />
)

export const DirectRunBlocked = () => {
  const [attempted, setAttempted] = useState(false)
  const runbookNames = [
    'Back up database',
    'Rotate database credentials',
    'Deploy application',
  ]
  const runbookRows: TInstallRunbookRow[] = runbookNames.map((name, index) => {
    const managedByQueue = name === 'Deploy application'
    return {
      runbookId: `runbook-${index + 1}`,
      runbookName: name,
      description: (
        <Text variant="subtext" theme="neutral">
          {managedByQueue
            ? 'Deploy the approved application build and verify its services.'
            : 'Run an operational procedure for this install.'}
        </Text>
      ),
      labels: (
        <Badge size="sm" theme={managedByQueue ? 'warn' : 'neutral'}>
          {managedByQueue ? 'In operations queue' : 'production'}
        </Badge>
      ),
      lastUpdated: (
        <Text variant="subtext" theme="neutral">
          3 days ago
        </Text>
      ),
      lastRun: (
        <Text variant="subtext" theme="neutral">
          {index === 0 ? '8 minutes ago' : '2 days ago'}
        </Text>
      ),
      href: `/org-1/installs/install-1/runbooks/runbook-${index + 1}`,
      latestRunHref: `/org-1/installs/install-1/workflows/workflow-${index + 1}`,
      installRunbook: {
        id: `install-runbook-${index + 1}`,
        runbook_id: `runbook-${index + 1}`,
      } as TInstallRunbookRow['installRunbook'],
      runAction: (
        <Button
          isMenuButton
          disabled={managedByQueue && attempted}
          tooltipProps={
            managedByQueue && attempted
              ? {
                  className: 'block !w-full',
                  position: 'left',
                  tipContent:
                    'Cannot run — this runbook is managed by August maintenance',
                }
              : undefined
          }
          onClick={() => managedByQueue && setAttempted(true)}
        >
          Run runbook
          <Icon
            variant={managedByQueue && attempted ? 'LockIcon' : 'PlayIcon'}
          />
        </Button>
      ),
    }
  })

  return (
    <PageSection>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Runbooks
        </Text>
        <Text variant="subtext" theme="neutral">
          View and run operational procedures for this install.
        </Text>
      </HeadingGroup>

      {attempted && (
        <Banner theme="warn">
          <div className="flex flex-col gap-1">
            <Text weight="strong">Run blocked by operations queue</Text>
            <Text variant="subtext">
              Deploy application is already queued in August maintenance. It can
              run only after Rotate database credentials completes.
            </Text>
          </div>
        </Banner>
      )}

      <InstallRunbooksTableComponent
        key={attempted ? 'blocked' : 'ready'}
        data={runbookRows}
        isLoading={false}
        pagination={{ hasNext: false, offset: 0, limit: 20 }}
      />
    </PageSection>
  )
}

type TInstallGroupQueueTarget = {
  id: string
  installName: string
  queueName: string
  queueState: string
  queuePosition: number
  alreadyQueued?: boolean
  reordered?: boolean
}

const installGroupTargets: TInstallGroupQueueTarget[] = [
  {
    id: 'install-west',
    installName: 'production-us-west',
    queueName: 'August maintenance',
    queueState: '2 operations ahead',
    queuePosition: 3,
  },
  {
    id: 'install-east',
    installName: 'production-us-east',
    queueName: 'Routine operations',
    queueState: '1 operation running',
    queuePosition: 2,
  },
  {
    id: 'install-eu-west',
    installName: 'production-eu-west',
    queueName: 'No active queue',
    queueState: 'No operations ahead',
    queuePosition: 1,
  },
  {
    id: 'install-eu-central',
    installName: 'production-eu-central',
    queueName: 'August maintenance',
    queueState: 'Already queued at position 2',
    queuePosition: 2,
    alreadyQueued: true,
  },
  {
    id: 'install-ap-southeast',
    installName: 'production-ap-southeast',
    queueName: 'Patch window',
    queueState: '3 operations ahead',
    queuePosition: 4,
  },
  {
    id: 'install-ap-northeast',
    installName: 'production-ap-northeast',
    queueName: 'No active queue',
    queueState: 'No operations ahead',
    queuePosition: 1,
  },
]

const installGroupInstallColumn: ColumnDef<TInstallGroupQueueTarget> = {
  accessorKey: 'installName',
  header: 'Install',
  cell: ({ row }) => (
    <div className="flex flex-col items-start gap-0.5">
      <Link href={`/org-1/installs/${row.original.id}`}>
        {row.original.installName}
      </Link>
      <Text variant="subtext" theme="neutral">
        {row.original.queueName}
      </Text>
    </div>
  ),
}

const getInstallGroupColumns = (
  onReorder: (target: TInstallGroupQueueTarget) => void
): ColumnDef<TInstallGroupQueueTarget>[] => [
  installGroupInstallColumn,
  {
    accessorKey: 'queueState',
    header: 'Current queue',
    cell: ({ row }) => (
      <Text variant="subtext" theme="neutral">
        {row.original.queueState}
      </Text>
    ),
  },
  {
    accessorKey: 'queuePosition',
    header: 'Result',
    cell: ({ row }) =>
      row.original.reordered ? (
        <div className="flex flex-col items-start gap-1">
          <Badge size="sm" theme="brand">
            Reordered · Position {row.original.queuePosition}
          </Badge>
          <Text variant="subtext" theme="neutral">
            {row.original.alreadyQueued
              ? 'Existing required entry'
              : 'New required entry'}
          </Text>
        </div>
      ) : row.original.alreadyQueued ? (
        <div className="flex flex-col items-start gap-1">
          <Badge size="sm" theme="info">
            Existing entry reused
          </Badge>
          <Text variant="subtext" theme="neutral">
            Marked required
          </Text>
        </div>
      ) : (
        <div className="flex flex-col items-start gap-1">
          <Badge size="sm" theme="brand">
            Position {row.original.queuePosition}
          </Badge>
          <Text variant="subtext" theme="neutral">
            New required entry
          </Text>
        </div>
      ),
  },
  {
    id: 'actions',
    header: '',
    cell: ({ row }) => (
      <Button variant="ghost" size="xs" onClick={() => onReorder(row.original)}>
        <Icon variant="ListIcon" />
        Reorder
      </Button>
    ),
  },
]

const trackedInstallGroupColumns: ColumnDef<TInstallGroupQueueTarget>[] = [
  installGroupInstallColumn,
  {
    accessorKey: 'queuePosition',
    header: 'Queue position',
    cell: ({ row }) => <Text family="mono">{row.original.queuePosition}</Text>,
  },
  {
    id: 'status',
    header: 'Status',
    cell: ({ row }) => (
      <Status status={row.original.queuePosition === 1 ? 'info' : 'default'}>
        {row.original.queuePosition === 1
          ? 'Next to run'
          : `Waiting for ${row.original.queuePosition - 1} operation${row.original.queuePosition === 2 ? '' : 's'}`}
      </Status>
    ),
  },
]

const InstallQueueReorder = ({
  target,
  reorder,
  onCancel,
  onSave,
}: {
  target: TInstallGroupQueueTarget
  reorder?: InstallQueueReorderValues
  onCancel: () => void
  onSave: (values: InstallQueueReorderValues) => void
}) => {
  const hasRunningOperation = target.queueName !== 'No active queue'
  const otherOperations = [
    ...(hasRunningOperation
      ? [
          {
            id: 'apply-config',
            name: 'Apply approved configuration',
            detail: 'Action · Running now',
            running: true,
          },
        ]
      : []),
    {
      id: 'verify-health',
      name: 'Verify service health',
      detail: 'Action · Queued',
      running: false,
    },
    {
      id: 'deploy-application',
      name: 'Deploy application',
      detail: 'Deploy · Queued',
      running: false,
    },
  ]
  const minimumPosition = hasRunningOperation ? 2 : 1
  const maximumPosition = otherOperations.length + 1
  const form = useForm({
    defaultValues: {
      queuePosition: reorder?.queuePosition ?? target.queuePosition,
    } as InstallQueueReorderValues,
    validators: {
      onMount: installQueueReorderSchema,
      onChange: installQueueReorderSchema,
    },
    onSubmit: ({ value }) => onSave(value),
  })
  const canSubmit = useStore(form.store, (state) => state.canSubmit)
  const queuePosition = useStore(
    form.store,
    (state) => state.values.queuePosition
  )

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 p-6">
      <div className="flex items-start gap-3">
        <Button
          variant="ghost"
          size="sm"
          aria-label="Back to queue preview"
          onClick={onCancel}
        >
          <Icon variant="ArrowLeftIcon" />
        </Button>
        <HeadingGroup className="gap-1">
          <Text as="h1" variant="h2" weight="stronger">
            Reorder install queue
          </Text>
          <Text theme="neutral">
            Choose where the required runbook should run on{' '}
            <Link href={`/org-1/installs/${target.id}`}>
              {target.installName}
            </Link>
            .
          </Text>
        </HeadingGroup>
      </div>

      <Card className="!gap-4 !p-4">
        <div className="flex items-start gap-3">
          <div className="flex size-9 shrink-0 items-center justify-center rounded-md bg-cool-grey-100 dark:bg-dark-grey-700">
            <Icon variant="BookOpenTextIcon" size={18} theme="neutral" />
          </div>
          <div className="flex min-w-0 flex-1 flex-col gap-1">
            <div className="flex flex-wrap items-center gap-2">
              <Text weight="strong">Rotate database credentials</Text>
              <Badge size="sm" theme="brand">
                <Icon variant="LockIcon" size={11} />
                Required
              </Badge>
            </div>
            <Text variant="subtext" theme="neutral">
              {target.queueName} · Reordering affects only {target.installName}
            </Text>
          </div>
        </div>
      </Card>

      <form
        autoComplete="off"
        noValidate
        onSubmit={(event) => event.preventDefault()}
        className="flex flex-col gap-3"
      >
        <div className="flex flex-col gap-1">
          <Text variant="base" weight="strong">
            Queue order
          </Text>
          <Text variant="subtext" theme="neutral">
            Move the runbook within this install queue. A running operation
            cannot be moved.
          </Text>
        </div>
        <form.Field name="queuePosition">
          {(field) => {
            const operations = [...otherOperations]
            operations.splice(
              Math.min(field.state.value - 1, operations.length),
              0,
              {
                id: 'rotate-credentials',
                name: 'Rotate database credentials',
                detail: 'Runbook · Required group operation',
                running: false,
              }
            )

            return (
              <div className="overflow-hidden rounded-lg border">
                {operations.map((operation, index) => {
                  const position = index + 1
                  const isTarget = operation.id === 'rotate-credentials'
                  return (
                    <div
                      key={operation.id}
                      className={`flex items-center gap-3 border-t p-4 first:border-t-0 ${
                        isTarget ? 'bg-primary-50 dark:bg-primary-950/20' : ''
                      }`}
                    >
                      <Icon
                        variant={
                          operation.running
                            ? 'LockIcon'
                            : operation.id === 'deploy-application'
                              ? 'RocketIcon'
                              : operation.id === 'rotate-credentials'
                                ? 'BookOpenTextIcon'
                                : 'LightningIcon'
                        }
                        size={16}
                        theme="neutral"
                      />
                      <Text family="mono" variant="subtext" theme="neutral">
                        {position.toString().padStart(2, '0')}
                      </Text>
                      <div className="flex min-w-0 flex-1 flex-col gap-1">
                        <div className="flex flex-wrap items-center gap-2">
                          <Text weight="strong">{operation.name}</Text>
                          {isTarget && (
                            <Badge size="sm" theme="brand">
                              <Icon variant="LockIcon" size={11} />
                              Required
                            </Badge>
                          )}
                        </div>
                        <Text variant="subtext" theme="neutral">
                          {operation.detail}
                        </Text>
                      </div>
                      {operation.running ? (
                        <Status status="info">Running</Status>
                      ) : isTarget ? (
                        <div className="flex items-center gap-1">
                          {field.state.value > minimumPosition && (
                            <Button
                              variant="ghost"
                              size="xs"
                              aria-label="Move runbook up"
                              onClick={() =>
                                field.handleChange(field.state.value - 1)
                              }
                            >
                              <Icon variant="ArrowUpIcon" />
                            </Button>
                          )}
                          {field.state.value < maximumPosition && (
                            <Button
                              variant="ghost"
                              size="xs"
                              aria-label="Move runbook down"
                              onClick={() =>
                                field.handleChange(field.state.value + 1)
                              }
                            >
                              <Icon variant="ArrowDownIcon" />
                            </Button>
                          )}
                        </div>
                      ) : (
                        <Status status="default">Queued</Status>
                      )}
                    </div>
                  )
                })}
              </div>
            )
          }}
        </form.Field>
      </form>

      <div className="flex items-start gap-2">
        <Icon variant="LockIcon" size={15} />
        <Text variant="subtext">
          The runbook remains required regardless of its queue position.
        </Text>
      </div>

      <div className="flex justify-end gap-2 border-t pt-4">
        <Button variant="secondary" onClick={onCancel}>
          Cancel
        </Button>
        <Button
          variant="primary"
          disabled={!canSubmit || queuePosition === target.queuePosition}
          tooltipProps={
            queuePosition === target.queuePosition
              ? { tipContent: 'Move the runbook to save a new order' }
              : undefined
          }
          onClick={() => form.handleSubmit()}
        >
          <Icon variant="CheckIcon" />
          Save order
        </Button>
      </div>
    </div>
  )
}

export const QueueRunbookForInstallGroup = () => {
  const [queued, setQueued] = useState(false)
  const [reorderingInstallId, setReorderingInstallId] = useState<string>()
  const [reorders, setReorders] = useState<
    Record<string, InstallQueueReorderValues>
  >({})
  const effectiveTargets = installGroupTargets.map((target) => {
    const reorder = reorders[target.id]
    return reorder
      ? {
          ...target,
          reordered: true,
          queuePosition: reorder.queuePosition,
        }
      : target
  })
  const reorderingTarget = installGroupTargets.find(
    (target) => target.id === reorderingInstallId
  )

  if (reorderingTarget) {
    return (
      <InstallQueueReorder
        target={reorderingTarget}
        reorder={reorders[reorderingTarget.id]}
        onCancel={() => setReorderingInstallId(undefined)}
        onSave={(values) => {
          setReorders((current) => ({
            ...current,
            [reorderingTarget.id]: values,
          }))
          setReorderingInstallId(undefined)
        }}
      />
    )
  }

  if (queued) {
    return (
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 p-6">
        <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
          <div className="flex items-start gap-3">
            <Button
              variant="ghost"
              size="sm"
              aria-label="Back to install group"
              onClick={() => setQueued(false)}
            >
              <Icon variant="ArrowLeftIcon" />
            </Button>
            <HeadingGroup className="gap-1">
              <div className="flex flex-wrap items-center gap-2">
                <Text as="h1" variant="h2" weight="stronger">
                  Rotate database credentials
                </Text>
                <Status status="info" variant="badge">
                  Queued
                </Status>
              </div>
              <Text theme="neutral">Group operation for Production fleet</Text>
            </HeadingGroup>
          </div>
          <Button variant="secondary" onClick={() => setQueued(false)}>
            View install group
          </Button>
        </div>

        <Banner theme="success">
          <div className="flex flex-col gap-1">
            <Text weight="strong">Runbook queued for 6 installs</Text>
            <Text>
              Added 5 required entries and linked 1 existing entry. Each install
              keeps its own queue ordering.
            </Text>
          </div>
        </Banner>

        <Card className="!gap-4 !p-4">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="flex size-9 items-center justify-center rounded-md bg-cool-grey-100 dark:bg-dark-grey-700">
                <Icon variant="TreeStructureIcon" size={18} theme="neutral" />
              </div>
              <div className="flex flex-col gap-0.5">
                <Text weight="strong">Group progress</Text>
                <Text variant="subtext" theme="neutral">
                  Membership snapshot · 6 installs
                </Text>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <Text variant="subtext" theme="neutral">
                0 of 6 complete
              </Text>
              <Badge size="sm" theme="brand">
                <Icon variant="LockIcon" size={11} />
                Required
              </Badge>
            </div>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-cool-grey-100 dark:bg-dark-grey-700" />
        </Card>

        <div className="flex flex-col gap-2">
          <Text variant="base" weight="strong">
            Install entries
          </Text>
          <Text variant="subtext" theme="neutral">
            Entries run independently when they reach the front of each install
            queue.
          </Text>
        </div>
        <Table
          columns={trackedInstallGroupColumns}
          data={effectiveTargets}
          enableSearch={false}
          enableSorting={false}
        />
      </div>
    )
  }

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 p-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
        <div className="flex items-start gap-3">
          <Button variant="ghost" size="sm" aria-label="Back to install group">
            <Icon variant="ArrowLeftIcon" />
          </Button>
          <HeadingGroup className="gap-1">
            <Text as="h1" variant="h2" weight="stronger">
              Queue runbook for install group
            </Text>
            <Text theme="neutral">
              Add one required operation to every install queue in Production
              fleet.
            </Text>
          </HeadingGroup>
        </div>
        <Button variant="primary" onClick={() => setQueued(true)}>
          <Icon variant="StackPlusIcon" />
          Queue for 6 installs
        </Button>
      </div>

      <Banner theme="info">
        <div className="flex flex-col gap-1">
          <Text weight="strong">Membership is captured on submission</Text>
          <Text variant="subtext">
            These 6 installs form the operation snapshot. Installs added to the
            group later are not included.
          </Text>
        </div>
      </Banner>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_20rem]">
        <div className="flex min-w-0 flex-col gap-6">
          <Card className="!gap-4 !p-4">
            <div className="flex items-start gap-3">
              <div className="flex size-9 shrink-0 items-center justify-center rounded-md bg-cool-grey-100 dark:bg-dark-grey-700">
                <Icon variant="BookOpenTextIcon" size={18} theme="neutral" />
              </div>
              <div className="flex min-w-0 flex-1 flex-col gap-1">
                <Text variant="subtext" theme="neutral">
                  Runbook
                </Text>
                <div className="flex flex-wrap items-center gap-2">
                  <Text variant="base" weight="strong">
                    Rotate database credentials
                  </Text>
                  <Badge size="sm" theme="brand">
                    <Icon variant="LockIcon" size={11} />
                    Required
                  </Badge>
                </div>
                <Text variant="subtext" theme="neutral">
                  Rotate credentials and verify dependent services.
                </Text>
              </div>
            </div>
            <div className="-mx-4">
              <Divider />
            </div>
            <div className="flex items-start gap-3">
              <div className="flex size-9 shrink-0 items-center justify-center rounded-md bg-cool-grey-100 dark:bg-dark-grey-700">
                <Icon variant="TreeStructureIcon" size={18} theme="neutral" />
              </div>
              <div className="flex flex-col gap-1">
                <Text variant="subtext" theme="neutral">
                  Install group
                </Text>
                <Text weight="strong">Production fleet</Text>
                <Text variant="subtext" theme="neutral">
                  6 matching installs · Snapshot taken when queued
                </Text>
              </div>
            </div>
          </Card>
        </div>

        <Card className="h-fit !gap-4 !p-4">
          <Text weight="strong">Submission summary</Text>
          <div className="flex flex-col gap-3">
            <div className="flex items-center justify-between gap-3">
              <Text variant="subtext" theme="neutral">
                Group members
              </Text>
              <Text weight="strong">6</Text>
            </div>
            <div className="flex items-center justify-between gap-3">
              <Text variant="subtext" theme="neutral">
                New queue entries
              </Text>
              <Text weight="strong">5</Text>
            </div>
            <div className="flex items-center justify-between gap-3">
              <Text variant="subtext" theme="neutral">
                Existing entry reused
              </Text>
              <Text weight="strong">1</Text>
            </div>
          </div>
          <div className="-mx-4">
            <Divider />
          </div>
          <div className="flex items-start gap-2">
            <Icon variant="LockIcon" size={15} theme="brand" />
            <Text variant="subtext" theme="neutral">
              All 6 entries are required and cannot be removed from their
              install queues.
            </Text>
          </div>
          <div className="flex items-start gap-2">
            <Icon
              variant="ArrowsInLineVerticalIcon"
              size={15}
              theme="neutral"
            />
            <Text variant="subtext" theme="neutral">
              Each entry waits for the operations already ahead of it.
            </Text>
          </div>
        </Card>
      </div>

      <div className="flex flex-col gap-2">
        <Text variant="base" weight="strong">
          Queue preview
        </Text>
        <Text variant="subtext" theme="neutral">
          Existing operations keep their positions. Duplicate entries are reused
          instead of added twice. Reorder an install to change where the
          required entry lands.
        </Text>
      </div>
      <Table
        columns={getInstallGroupColumns((target) =>
          setReorderingInstallId(target.id)
        )}
        data={effectiveTargets}
        enableSearch={false}
        enableSorting={false}
      />
    </div>
  )
}
