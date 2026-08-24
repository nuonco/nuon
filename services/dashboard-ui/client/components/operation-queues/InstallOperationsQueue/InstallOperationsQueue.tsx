import { useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { Badge, type TBadgeTheme } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { Divider } from '@/components/common/Divider'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Menu } from '@/components/common/Menu'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { FormRadioGroup } from '@/components/common/form/FormRadioGroup'
import { cn } from '@/utils/classnames'
import { admissionRulesSchema, type AdmissionRulesValues } from './schema'

export type TOperationQueueStatus =
  | 'completed'
  | 'draft'
  | 'out-of-band'
  | 'running'

export type TOperationQueueEntryStatus =
  | 'blocked'
  | 'failed'
  | 'in-progress'
  | 'queued'
  | 'success'

export interface IOperationQueueEntry {
  id: string
  name: string
  type: 'action' | 'deploy' | 'runbook'
  status: TOperationQueueEntryStatus
  detail: string
  required?: boolean
  blockedBy?: string
}

export interface IQueueOperationOption {
  id: string
  name: string
  type: IOperationQueueEntry['type']
  detail: string
}

export interface IOperationQueueRule {
  id: string
  name: string
  description: string
}

export interface IInstallOperationsQueue {
  installName: string
  queueName: string
  status: TOperationQueueStatus
  entries: IOperationQueueEntry[]
  rules: IOperationQueueRule[]
  onAddOperation?: () => void
  onConfigureRules?: () => void
  onFinishOutsideOperation?: () => void
  onRunOutsideQueue?: () => void
  onStart?: () => void
}

export interface IAddQueueOperationStudio {
  installName: string
  queueName: string
  entries: IOperationQueueEntry[]
  options: IQueueOperationOption[]
  selectedOptionId?: string
  required: boolean
  onAdd: () => void
  onCancel: () => void
  onRequiredChange: (required: boolean) => void
  onSelectOption: (optionId: string) => void
}

export type TScheduledOperationsPolicy =
  AdmissionRulesValues['scheduledOperationsPolicy']
export type TQueueAdmissionPolicy = AdmissionRulesValues['admissionPolicy']

export type TQueueExemptionOperationType =
  | 'action'
  | 'component deploy'
  | 'runbook'
  | 'sandbox operation'

export interface IQueueAdmissionExemptionOption {
  id: string
  operationType: TQueueExemptionOperationType
  labelKey: string
  labelValue: string
  matchingOperations: string[]
}

export interface IQueueAdmissionRulesStudio {
  installName: string
  queueName: string
  admissionPolicy: TQueueAdmissionPolicy
  scheduledOperationsPolicy: TScheduledOperationsPolicy
  exemptionOptions: IQueueAdmissionExemptionOption[]
  selectedExemptionIds: string[]
  onCancel: () => void
  onSave: (rules: {
    admissionPolicy: TQueueAdmissionPolicy
    scheduledOperationsPolicy: TScheduledOperationsPolicy
    selectedExemptionIds: string[]
  }) => void
}

const statusTheme: Record<TOperationQueueStatus, TBadgeTheme> = {
  completed: 'success',
  draft: 'neutral',
  'out-of-band': 'warn',
  running: 'info',
}

const entryIcons: Record<IOperationQueueEntry['type'], TIconVariant> = {
  action: 'LightningIcon',
  deploy: 'RocketIcon',
  runbook: 'BookOpenTextIcon',
}

const entryStatusLabels: Record<TOperationQueueEntryStatus, string> = {
  blocked: 'Blocked',
  failed: 'Failed',
  'in-progress': 'Running',
  queued: 'Queued',
  success: 'Complete',
}

const queueStatusLabels: Record<TOperationQueueStatus, string> = {
  completed: 'Completed',
  draft: 'Draft',
  'out-of-band': 'Outside operation active',
  running: 'Running',
}

const OperationRow = ({
  entry,
  position,
}: {
  entry: IOperationQueueEntry
  position: number
}) => (
  <div className="flex items-start gap-4 py-4 first:pt-0 last:pb-0">
    <div className="flex w-6 shrink-0 justify-center pt-2">
      <Text family="mono" variant="subtext" theme="neutral">
        {position.toString().padStart(2, '0')}
      </Text>
    </div>
    <div className="flex size-9 shrink-0 items-center justify-center rounded-md bg-cool-grey-100 dark:bg-dark-grey-700">
      <Icon variant={entryIcons[entry.type]} size={18} theme="neutral" />
    </div>
    <div className="flex min-w-0 flex-1 flex-col gap-1">
      <div className="flex flex-wrap items-center gap-2">
        <Text weight="strong">{entry.name}</Text>
        <Badge size="xs" variant="code" theme="default">
          {entry.type}
        </Badge>
        {entry.required && (
          <Badge size="xs" theme="brand">
            <Icon variant="LockIcon" size={11} />
            Required
          </Badge>
        )}
      </div>
      <Text variant="subtext" theme="neutral">
        {entry.detail}
      </Text>
      {entry.blockedBy && (
        <Text flex variant="subtext" theme="warn">
          <Icon variant="LockIcon" size={13} />
          Waiting for {entry.blockedBy}
        </Text>
      )}
    </div>
    <Status status={entry.status}>{entryStatusLabels[entry.status]}</Status>
  </div>
)

const QueueActions = ({
  status,
  onAddOperation,
  onFinishOutsideOperation,
  onRunOutsideQueue,
  onStart,
}: Pick<
  IInstallOperationsQueue,
  | 'status'
  | 'onAddOperation'
  | 'onFinishOutsideOperation'
  | 'onRunOutsideQueue'
  | 'onStart'
>) => {
  if (status === 'out-of-band') {
    return (
      <Button variant="primary" onClick={onFinishOutsideOperation}>
        <Icon variant="ShieldCheckIcon" />
        Finish outside operation
      </Button>
    )
  }

  if (status === 'completed') return null

  return (
    <div className="flex flex-wrap items-center justify-end gap-2">
      {status === 'draft' && (
        <Button variant="primary" onClick={onStart}>
          <Icon variant="PlayIcon" />
          Start queue
        </Button>
      )}
      <Button
        variant={status === 'running' ? 'primary' : 'secondary'}
        onClick={onAddOperation}
      >
        <Icon variant="PlusIcon" />
        Add operation
      </Button>
      {status === 'running' && (
        <Button variant="secondary" onClick={onRunOutsideQueue}>
          <Icon variant="ShieldIcon" />
          Run outside queue
        </Button>
      )}
    </div>
  )
}

export const QueueAdmissionRulesStudio = ({
  installName,
  queueName,
  admissionPolicy,
  scheduledOperationsPolicy,
  exemptionOptions,
  selectedExemptionIds,
  onCancel,
  onSave,
}: IQueueAdmissionRulesStudio) => {
  const [currentExemptionIds, setCurrentExemptionIds] =
    useState(selectedExemptionIds)
  const form = useForm({
    defaultValues: {
      admissionPolicy,
      scheduledOperationsPolicy,
    } as AdmissionRulesValues,
    validators: {
      onMount: admissionRulesSchema,
      onChange: admissionRulesSchema,
    },
    onSubmit: ({ value }) =>
      onSave({
        admissionPolicy: value.admissionPolicy,
        scheduledOperationsPolicy: value.scheduledOperationsPolicy,
        selectedExemptionIds: currentExemptionIds,
      }),
  })
  const currentScheduledPolicy = useStore(
    form.store,
    (state) => state.values.scheduledOperationsPolicy
  )
  const currentAdmissionPolicy = useStore(
    form.store,
    (state) => state.values.admissionPolicy
  )
  const isStrict = currentAdmissionPolicy === 'strict'
  const canSubmit = useStore(form.store, (state) => state.canSubmit)
  const selectedExemptions = exemptionOptions.filter((exemption) =>
    currentExemptionIds.includes(exemption.id)
  )
  const availableExemptions = exemptionOptions.filter(
    (exemption) => !currentExemptionIds.includes(exemption.id)
  )

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-col gap-6 p-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
        <div className="flex items-start gap-3">
          <Button
            variant="ghost"
            size="sm"
            aria-label="Back to operations queue"
            onClick={onCancel}
          >
            <Icon variant="ArrowLeftIcon" />
          </Button>
          <div className="flex flex-col gap-1">
            <Text as="h1" variant="h2" weight="stronger">
              Admission rules
            </Text>
            <Text theme="neutral">
              Control which operations may start outside {queueName} on{' '}
              {installName}.
            </Text>
            <Badge size="sm" theme="neutral">
              <Icon variant="ArrowsInLineVerticalIcon" size={13} />
              {queueName}
            </Badge>
          </div>
        </div>
        <Button
          variant="primary"
          disabled={!canSubmit}
          onClick={() => form.handleSubmit()}
        >
          <Icon variant="CheckIcon" />
          Save rules
        </Button>
      </div>

      <Card className="!gap-4 !p-4">
        <div className="flex flex-col gap-1">
          <Text variant="base" weight="strong">
            Default behavior
          </Text>
          <Text variant="subtext" theme="neutral">
            Choose how operations that were not submitted through this queue are
            handled.
          </Text>
        </div>

        <div className="-mx-4">
          <Divider />
        </div>

        <form
          autoComplete="off"
          noValidate
          onSubmit={(event) => event.preventDefault()}
          className="flex flex-col gap-4"
        >
          <div className="grid gap-4 sm:grid-cols-[minmax(0,1fr)_minmax(18rem,1fr)]">
            <div className="flex flex-col gap-1">
              <Text weight="strong">Operations outside the queue</Text>
              <Text variant="subtext" theme="neutral">
                Manual, scheduled, and API-triggered workflows.
              </Text>
            </div>
            <form.Field name="admissionPolicy">
              {(field) => (
                <FormRadioGroup
                  field={field}
                  options={[
                    {
                      value: 'enqueue',
                      label: (
                        <span className="flex flex-col gap-0.5">
                          <Text weight="strong">Add to queue</Text>
                          <Text variant="subtext" theme="neutral">
                            Admit the operation at the end of the queue.
                          </Text>
                        </span>
                      ),
                    },
                    {
                      value: 'strict',
                      label: (
                        <span className="flex flex-col gap-0.5">
                          <Text flex weight="strong">
                            <Icon variant="LockIcon" size={14} />
                            Block all
                          </Text>
                          <Text variant="subtext" theme="neutral">
                            Only operations already in this queue may start.
                          </Text>
                        </span>
                      ),
                    },
                  ]}
                />
              )}
            </form.Field>
          </div>

          <div className="-mx-4">
            <Divider />
          </div>

          <div className="grid gap-4 sm:grid-cols-[minmax(0,1fr)_minmax(18rem,1fr)]">
            <div className="flex flex-col gap-1">
              <Text weight="strong">Scheduled operations</Text>
              <Text variant="subtext" theme="neutral">
                Cron-triggered drifts, actions, and runbooks.
              </Text>
            </div>
            {isStrict ? (
              <div className="flex flex-col items-start gap-1 sm:items-end">
                <Text flex weight="strong" theme="warn">
                  <Icon variant="LockIcon" size={14} />
                  Blocked
                </Text>
                <Text
                  variant="subtext"
                  theme="neutral"
                  className="sm:text-right"
                >
                  Scheduled workflows cannot start outside this queue.
                </Text>
              </div>
            ) : (
              <form.Field name="scheduledOperationsPolicy">
                {(field) => (
                  <FormRadioGroup
                    field={field}
                    options={[
                      {
                        value: 'enqueue',
                        label: (
                          <span className="flex flex-col gap-0.5">
                            <Text weight="strong">Add to queue</Text>
                            <Text variant="subtext" theme="neutral">
                              Wait for its queue position.
                            </Text>
                          </span>
                        ),
                      },
                      {
                        value: 'direct',
                        label: (
                          <span className="flex flex-col gap-0.5">
                            <Text weight="strong">Run directly</Text>
                            <Text variant="subtext" theme="neutral">
                              Skip queue admission unless a required entry
                              exists.
                            </Text>
                          </span>
                        ),
                      },
                    ]}
                  />
                )}
              </form.Field>
            )}
          </div>
        </form>
      </Card>

      <Card className="!gap-4 !p-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="flex flex-col gap-1">
            <Text variant="base" weight="strong">
              Operation exemptions
            </Text>
            <Text variant="subtext" theme="neutral">
              {isStrict
                ? 'Strict queue mode disables every operation exemption.'
                : 'Operations matching any exemption can run directly from manual, scheduled, or API requests.'}
            </Text>
          </div>
          <Dropdown
            key={availableExemptions.map((exemption) => exemption.id).join('-')}
            id="add-operation-exemption"
            alignment="right"
            variant="secondary"
            buttonText="Add exemption"
            disabled={isStrict || !availableExemptions.length}
            tooltipProps={
              isStrict
                ? { tipContent: 'Exemptions are disabled in strict queue mode' }
                : availableExemptions.length
                  ? undefined
                  : { tipContent: 'All available exemptions are in use' }
            }
          >
            <Menu className="w-80">
              <Text>Available exemptions</Text>
              {availableExemptions.map((exemption) => (
                <Button
                  key={exemption.id}
                  isMenuButton
                  onClick={() =>
                    setCurrentExemptionIds((current) => [
                      ...current,
                      exemption.id,
                    ])
                  }
                >
                  <span className="flex min-w-0 items-center gap-2">
                    <Badge size="xs" theme="neutral">
                      {exemption.operationType}
                    </Badge>
                    <Badge variant="code" size="sm" theme="default">
                      {exemption.labelKey}={exemption.labelValue}
                    </Badge>
                  </span>
                  <Icon variant="PlusIcon" size={14} />
                </Button>
              ))}
            </Menu>
          </Dropdown>
        </div>

        <div className="-mx-4">
          <Divider />
        </div>

        {isStrict ? (
          <div className="flex flex-col items-center gap-1 py-4 text-center">
            <Icon variant="LockIcon" size={20} theme="warn" />
            <Text weight="strong">All exemptions blocked</Text>
            <Text variant="subtext" theme="neutral">
              Manual, scheduled, and API operations must already be in this
              queue.
            </Text>
          </div>
        ) : selectedExemptions.length ? (
          <div className="divide-y">
            {selectedExemptions.map((exemption) => (
              <div
                key={exemption.id}
                className="grid gap-3 py-4 first:pt-0 last:pb-0 sm:grid-cols-[minmax(0,1fr)_auto_auto] sm:items-center"
              >
                <div className="flex min-w-0 flex-col gap-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge size="xs" theme="neutral">
                      {exemption.operationType}
                    </Badge>
                    <Badge variant="code" size="sm" theme="default">
                      {exemption.labelKey}={exemption.labelValue}
                    </Badge>
                  </div>
                  <Text variant="subtext" theme="neutral">
                    Matches {exemption.matchingOperations.length}{' '}
                    {exemption.matchingOperations.length === 1
                      ? 'operation'
                      : 'operations'}
                    : {exemption.matchingOperations.join(', ')}
                  </Text>
                </div>
                <div className="flex flex-wrap items-center gap-1">
                  <Text variant="subtext" theme="neutral" className="mr-1">
                    Sources
                  </Text>
                  {['Manual', 'Scheduled', 'API'].map((source) => (
                    <Badge key={source} size="xs" theme="neutral">
                      {source}
                    </Badge>
                  ))}
                </div>
                <Button
                  variant="ghost"
                  size="xs"
                  aria-label={`Remove ${exemption.operationType} ${exemption.labelKey}=${exemption.labelValue} exemption`}
                  onClick={() =>
                    setCurrentExemptionIds((current) =>
                      current.filter((id) => id !== exemption.id)
                    )
                  }
                >
                  <Icon variant="TrashIcon" size={14} />
                </Button>
              </div>
            ))}
          </div>
        ) : (
          <div className="flex flex-col items-center gap-1 py-4 text-center">
            <Text weight="strong">No operation exemptions</Text>
            <Text variant="subtext" theme="neutral">
              Operations follow the default behavior above.
            </Text>
          </div>
        )}
      </Card>

      <Card className="!gap-4 !p-4 bg-cool-grey-50 dark:bg-dark-grey-700">
        <div className="flex items-center gap-2">
          <Icon variant="InfoIcon" size={18} theme="brand" />
          <Text variant="base" weight="strong">
            Impact
          </Text>
        </div>
        <div className="flex flex-col gap-2">
          {isStrict ? (
            <Text flex variant="subtext" weight="strong">
              <Icon variant="LockIcon" theme="warn" size={15} />
              Direct, scheduled, API, and label-exempt operations are blocked
              unless they are already in this queue.
            </Text>
          ) : (
            <>
              <Text flex variant="subtext">
                <Icon variant="CheckCircleIcon" theme="success" size={15} />
                Scheduled operations{' '}
                {currentScheduledPolicy === 'direct'
                  ? 'run directly.'
                  : 'are added to the queue.'}
              </Text>
              <Text flex variant="subtext">
                <Icon variant="TagIcon" theme="neutral" size={15} />
                {selectedExemptions.length}{' '}
                {selectedExemptions.length === 1
                  ? 'operation exemption may'
                  : 'operation exemptions may'}{' '}
                run directly.
              </Text>
            </>
          )}
          <Text flex variant="subtext" weight="strong">
            <Icon variant="LockIcon" theme="warn" size={15} />
            Required queue entries always use the queue and override these
            settings.
          </Text>
          <Text flex variant="subtext" theme="neutral">
            <Icon variant="ShieldIcon" size={15} />
            Authorized outside-queue runs are permissioned and audited
            separately.
          </Text>
        </div>
      </Card>
    </div>
  )
}

export const AddQueueOperationStudio = ({
  installName,
  queueName,
  entries,
  options,
  selectedOptionId,
  required,
  onAdd,
  onCancel,
  onRequiredChange,
  onSelectOption,
}: IAddQueueOperationStudio) => {
  const selectedOption = options.find(
    (option) => option.id === selectedOptionId
  )
  const predecessor = entries.at(-1)
  const position = entries.length + 1

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 p-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
        <div className="flex items-start gap-3">
          <Button
            variant="ghost"
            size="sm"
            aria-label="Back to operations queue"
            onClick={onCancel}
          >
            <Icon variant="ArrowLeftIcon" />
          </Button>
          <div className="flex flex-col gap-1">
            <Text as="h1" variant="h2" weight="stronger">
              Add operation
            </Text>
            <Text theme="neutral">
              {queueName} for {installName}
            </Text>
          </div>
        </div>
        <Button
          variant="primary"
          disabled={!selectedOption}
          tooltipProps={
            selectedOption
              ? undefined
              : { tipContent: 'Cannot add — choose an operation first' }
          }
          onClick={onAdd}
        >
          <Icon variant="PlusIcon" />
          Add to queue
        </Button>
      </div>

      <Banner theme="neutral">
        Adding an operation does not run it immediately. It joins the end of the
        queue and waits for the operations before it.
      </Banner>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,3fr)_minmax(19rem,2fr)]">
        <Card className="!gap-4 !p-4">
          <div className="flex items-center justify-between gap-3 pl-5">
            <div className="flex flex-col gap-1">
              <Text variant="base" weight="strong">
                Queue notebook
              </Text>
              <Text variant="subtext" theme="neutral">
                Choose and configure the next operation.
              </Text>
            </div>
            <Badge size="sm" theme="neutral">
              {position} operations
            </Badge>
          </div>

          <div className="flex flex-col gap-2 pl-5">
            {entries.map((entry, index) => (
              <div
                key={entry.id}
                className="flex items-center gap-3 rounded-lg border bg-cool-grey-50/70 px-3 py-2.5 dark:bg-dark-grey-800"
              >
                <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-blue-500/10 text-xs font-semibold text-blue-700 dark:text-blue-400">
                  {index + 1}
                </span>
                <Icon
                  variant={entryIcons[entry.type]}
                  size={16}
                  className="shrink-0 text-cool-grey-500 dark:text-cool-grey-400"
                />
                <div className="flex min-w-0 flex-col">
                  <Text weight="strong" className="truncate">
                    {entry.name}
                  </Text>
                  <Text variant="subtext" theme="neutral" className="truncate">
                    {entry.type}
                  </Text>
                </div>
                <Status className="ml-auto" status={entry.status}>
                  {entryStatusLabels[entry.status]}
                </Status>
              </div>
            ))}

            <div className="rounded-lg border bg-white shadow-sm ring-1 ring-blue-500/30 dark:bg-dark-grey-800">
              <div className="flex items-center gap-2 border-b px-3 py-2">
                <Icon
                  variant={
                    selectedOption
                      ? entryIcons[selectedOption.type]
                      : 'PlusCircleIcon'
                  }
                  size={14}
                />
                <Text variant="subtext" weight="strong">
                  Operation {position} ·{' '}
                  {selectedOption ? selectedOption.type : 'Choose operation'}
                </Text>
                {selectedOption && (
                  <Button
                    className="ml-auto"
                    variant="ghost"
                    size="xs"
                    onClick={() => onSelectOption('')}
                  >
                    Change operation
                  </Button>
                )}
              </div>

              <div className="flex flex-col gap-4 p-3">
                {selectedOption ? (
                  <>
                    <div className="flex items-start gap-3">
                      <div className="flex size-9 shrink-0 items-center justify-center rounded-md bg-cool-grey-100 dark:bg-dark-grey-700">
                        <Icon
                          variant={entryIcons[selectedOption.type]}
                          size={18}
                          theme="neutral"
                        />
                      </div>
                      <div className="flex min-w-0 flex-col gap-1">
                        <Text weight="strong">{selectedOption.name}</Text>
                        <Text variant="subtext" theme="neutral">
                          {selectedOption.detail}
                        </Text>
                        <Text flex variant="subtext" theme="neutral">
                          <Icon variant="ArrowsInLineVerticalIcon" size={13} />
                          Runs after {predecessor?.name ?? 'the queue starts'}
                        </Text>
                      </div>
                    </div>

                    <Divider />

                    <div className="flex flex-col gap-2">
                      <Text variant="subtext" weight="strong">
                        Execution requirement
                      </Text>
                      <div className="grid gap-2 sm:grid-cols-2">
                        <Button
                          variant="secondary"
                          aria-pressed={!required}
                          className={cn(
                            '!h-auto w-full !items-start !justify-start !whitespace-normal !p-3 text-left',
                            !required && 'bg-blue-500/5 ring-1 ring-blue-500/20'
                          )}
                          onClick={() => onRequiredChange(false)}
                        >
                          <Icon variant="ArrowsInLineVerticalIcon" size={17} />
                          <span className="flex flex-col gap-0.5">
                            <Text weight="strong">Standard</Text>
                            <Text variant="subtext" theme="neutral">
                              Can be removed before it starts.
                            </Text>
                          </span>
                        </Button>
                        <Button
                          variant="secondary"
                          aria-pressed={required}
                          className={cn(
                            '!h-auto w-full !items-start !justify-start !whitespace-normal !p-3 text-left',
                            required && 'bg-blue-500/5 ring-1 ring-blue-500/20'
                          )}
                          onClick={() => onRequiredChange(true)}
                        >
                          <Icon variant="LockIcon" size={17} />
                          <span className="flex flex-col gap-0.5">
                            <Text weight="strong">Required</Text>
                            <Text variant="subtext" theme="neutral">
                              Must complete and cannot be discarded.
                            </Text>
                          </span>
                        </Button>
                      </div>
                    </div>
                  </>
                ) : (
                  <div className="flex flex-col gap-3">
                    <div className="flex flex-col gap-1">
                      <Text weight="strong">Choose an operation</Text>
                      <Text variant="subtext" theme="neutral">
                        Add a runbook or another install operation to this
                        queue.
                      </Text>
                    </div>
                    <div className="grid gap-2">
                      {options.map((option) => (
                        <Button
                          key={option.id}
                          variant="secondary"
                          className="!h-auto w-full !justify-start !whitespace-normal !p-3 text-left"
                          onClick={() => onSelectOption(option.id)}
                        >
                          <div className="flex size-9 shrink-0 items-center justify-center rounded-md bg-cool-grey-100 dark:bg-dark-grey-700">
                            <Icon
                              variant={entryIcons[option.type]}
                              size={18}
                              theme="neutral"
                            />
                          </div>
                          <span className="flex min-w-0 flex-col gap-0.5">
                            <Text weight="strong">{option.name}</Text>
                            <Text variant="subtext" theme="neutral">
                              {option.detail}
                            </Text>
                          </span>
                          <Badge
                            className="ml-auto"
                            size="xs"
                            variant="code"
                            theme="default"
                          >
                            {option.type}
                          </Badge>
                        </Button>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </Card>

        <div className="flex flex-col gap-6 lg:sticky lg:top-4">
          <Card className="!gap-4 !p-4">
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <Icon variant="EyeIcon" size={17} />
                <Text variant="base" weight="strong">
                  Queue preview
                </Text>
              </div>
              <Badge size="sm" theme="neutral">
                One at a time
              </Badge>
            </div>
            <Divider />
            <div className="flex flex-col gap-3">
              {entries.map((entry, index) => (
                <div key={entry.id} className="flex items-center gap-3">
                  <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-cool-grey-100 text-xs font-semibold dark:bg-dark-grey-700">
                    {index + 1}
                  </span>
                  <Text className="min-w-0 flex-1 truncate">{entry.name}</Text>
                  {entry.required && <Icon variant="LockIcon" size={13} />}
                </div>
              ))}
              <div
                className={cn(
                  'flex items-center gap-3 rounded-lg border border-dashed p-3',
                  selectedOption && 'bg-blue-500/5 ring-1 ring-blue-500/20'
                )}
              >
                <span className="flex size-6 shrink-0 items-center justify-center rounded-full bg-blue-500/10 text-xs font-semibold text-blue-700 dark:text-blue-400">
                  {position}
                </span>
                <div className="flex min-w-0 flex-1 flex-col">
                  <Text weight="strong" className="truncate">
                    {selectedOption?.name ?? 'Choose an operation'}
                  </Text>
                  <Text variant="subtext" theme="neutral">
                    {selectedOption
                      ? 'Waiting to be added'
                      : 'New queue position'}
                  </Text>
                </div>
                {required && selectedOption && (
                  <Badge size="xs" theme="brand">
                    <Icon variant="LockIcon" size={11} />
                    Required
                  </Badge>
                )}
              </div>
            </div>
          </Card>

          <Card className="!gap-4 !p-4">
            <div className="flex items-center gap-2">
              <Icon variant="ShieldCheckIcon" theme="brand" size={18} />
              <Text weight="strong">Admission behavior</Text>
            </div>
            <Text variant="subtext" theme="neutral">
              Once added, direct attempts to run this operation are routed
              through the queue. It starts only when earlier operations have
              completed.
            </Text>
          </Card>
        </div>
      </div>
    </div>
  )
}

export const InstallOperationsQueue = ({
  installName,
  queueName,
  status,
  entries,
  rules,
  onAddOperation,
  onConfigureRules,
  onFinishOutsideOperation,
  onRunOutsideQueue,
  onStart,
}: IInstallOperationsQueue) => {
  const completeCount = entries.filter(
    (entry) => entry.status === 'success'
  ).length
  const activeEntry = entries.find((entry) => entry.status === 'in-progress')

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 p-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-start">
        <div className="flex flex-col gap-1">
          <div className="flex flex-wrap items-center gap-3">
            <Text as="h1" variant="h2" weight="stronger">
              Operations queue
            </Text>
            <Badge size="md" theme={statusTheme[status]}>
              {queueStatusLabels[status]}
            </Badge>
          </div>
          <Text theme="neutral">
            {queueName} for {installName}
          </Text>
        </div>
        <QueueActions
          status={status}
          onAddOperation={onAddOperation}
          onFinishOutsideOperation={onFinishOutsideOperation}
          onRunOutsideQueue={onRunOutsideQueue}
          onStart={onStart}
        />
      </div>

      {status === 'out-of-band' && (
        <Banner theme="warn">
          <div className="flex flex-col gap-1">
            <Text weight="strong">Operation running outside queue</Text>
            <Text variant="subtext">
              Normal queue dispatch is held automatically while this audited
              operation bypasses ordering rules. Dispatch resumes when it ends.
            </Text>
          </div>
        </Banner>
      )}

      <div className="grid gap-6 lg:grid-cols-[minmax(0,2fr)_minmax(18rem,1fr)]">
        <Card className="!gap-4 !p-4">
          <div className="flex items-center justify-between gap-4">
            <div className="flex flex-col gap-1">
              <Text variant="base" weight="strong">
                Queue operations
              </Text>
              <Text variant="subtext" theme="neutral">
                {completeCount} of {entries.length} complete
                {activeEntry ? ` · Running ${activeEntry.name}` : ''}
              </Text>
            </div>
            <Badge size="sm" theme="neutral">
              <Icon variant="ArrowsInLineVerticalIcon" size={13} />
              One at a time
            </Badge>
          </div>
          <Divider />
          <div className="divide-y">
            {entries.map((entry, index) => (
              <OperationRow key={entry.id} entry={entry} position={index + 1} />
            ))}
          </div>
        </Card>

        <div className="flex flex-col gap-6">
          <Card className="!gap-4 !p-4">
            <div className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-2">
                <Icon variant="ShieldCheckIcon" theme="brand" size={20} />
                <Text variant="base" weight="strong">
                  Queue rules
                </Text>
              </div>
              {onConfigureRules && (
                <Button variant="ghost" size="xs" onClick={onConfigureRules}>
                  Configure rules
                </Button>
              )}
            </div>
            <Divider />
            <div className="flex flex-col gap-4">
              {rules.map((rule) => (
                <div key={rule.id} className="flex items-start gap-3">
                  <Icon
                    className="mt-0.5 shrink-0"
                    variant="CheckCircleIcon"
                    theme="success"
                    size={17}
                  />
                  <div className="flex min-w-0 flex-col gap-0.5">
                    <Text weight="strong">{rule.name}</Text>
                    <Text variant="subtext" theme="neutral">
                      {rule.description}
                    </Text>
                  </div>
                </div>
              ))}
            </div>
          </Card>

          <Card className="!gap-4 !p-4">
            <div className="flex items-center gap-2">
              <Icon variant="ShieldIcon" theme="warn" size={20} />
              <Text variant="base" weight="strong">
                Run outside queue
              </Text>
            </div>
            <Text variant="subtext" theme="neutral">
              Run one urgent operation outside ordering rules. Normal dispatch
              is held automatically, and the operation and reason are recorded.
            </Text>
            <Text flex variant="subtext" theme="neutral">
              <Icon variant="LockIcon" size={13} />
              Permission and reason required
            </Text>
          </Card>
        </div>
      </div>
    </div>
  )
}
