import { useEffect, useMemo, useRef, useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import type { TAPIError, TCloudPlatform, TInstall } from '@/types'
import { Badge } from '../../atoms/Badge'
import { Banner } from '../../atoms/Banner'
import { Button } from '../../atoms/Button'
import { Icon } from '../../atoms/Icon'
import { Input } from '../../atoms/Input'
import { Radio } from '../../atoms/Radio'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import { CloudRegion } from '../../molecules/CloudRegion'
import { Disclosure } from '../../molecules/Disclosure'
import { FormErrorBanner } from '../../molecules/FormErrorBanner'
import { FormInput } from '../../molecules/FormInput'
import { FormSelect } from '../../molecules/FormSelect'
import { FormSwitch } from '../../molecules/FormSwitch'
import { AppSelect, type IAppSelectItem } from '../AppSelect'
import { Wizard } from '../../templates/Wizard'
import { cloudRegionsFor } from '../../../utils/cloud-regions'
import {
  isDuplicateInstallNameError,
  type ICreateInstallValues,
} from '../../../utils/create-install'
import { clearDraft, loadDraft, saveDraft } from '../../../utils/draft'
import type { IWizardDescriptor } from '../../../utils/wizard'

export type TInstallGroupKind = 'labels' | 'all' | 'install-ids' | 'wildcard'

export interface IInstallSetupGroup {
  id: string
  name: string
  kind: TInstallGroupKind
  labels?: Record<string, string>
}

export interface IInstallSetupBranch {
  id: string
  name: string
  groups: IInstallSetupGroup[]
}

export interface IInstallSetupInput {
  name: string
  label: string
  description?: string
  required?: boolean
  type: 'text' | 'number' | 'password' | 'boolean'
  defaultValue: string | boolean
}

export interface IInstallSetup {
  apps: IAppSelectItem[]
  branches: IInstallSetupBranch[]
  inputs: IInstallSetupInput[]
  platform: TCloudPlatform
  configurationKey?: string
  configurationLoading?: boolean
  configurationReady?: boolean
  appsLoading?: boolean
  requireTargetAccount?: boolean
  pending?: boolean
  error?: Error | TAPIError
  install?: TInstall
  persistence?: {
    orgId: string
    wizard: string
  }
  onAppChange: (appId: string) => void
  onBranchChange: (branchId?: string) => void
  onClearError?: () => void
  onDiscard?: () => void
  onSubmit: (values: ICreateInstallValues) => void
}

interface ISetupValues extends ICreateInstallValues {
  appId: string
}

interface ISetupState {
  app: boolean
  details: boolean
  created: boolean
  provisioned: boolean
  values: ISetupValues
}

interface IInstallSetupDraft {
  values: ISetupValues
  progress: {
    app: boolean
    details: boolean
  }
}

const LOCATION_COPY: Record<
  TCloudPlatform,
  {
    label: string
    placeholder: string
    accountLabel: string
    accountPlaceholder: string
  }
> = {
  aws: {
    label: 'AWS region',
    placeholder: 'Choose an AWS region',
    accountLabel: 'AWS account ID',
    accountPlaceholder: '123456789012',
  },
  azure: {
    label: 'Azure location',
    placeholder: 'Choose an Azure location',
    accountLabel: 'Azure subscription ID',
    accountPlaceholder: '00000000-0000-0000-0000-000000000000',
  },
  gcp: {
    label: 'GCP region',
    placeholder: 'Choose a GCP region',
    accountLabel: 'GCP project ID',
    accountPlaceholder: 'example-production',
  },
  unknown: {
    label: 'Region',
    placeholder: 'Choose a region',
    accountLabel: 'Cloud account ID',
    accountPlaceholder: '',
  },
}

const accountSchema = (
  platform: TCloudPlatform,
  requireTargetAccount: boolean
) => {
  let account = z.string()
  if (platform === 'aws') {
    account = requireTargetAccount
      ? z.string().regex(/^[0-9]{12}$/, 'Enter a 12-digit AWS account ID')
      : z
          .string()
          .refine(
            (value) => !value || /^[0-9]{12}$/.test(value),
            'Enter a 12-digit AWS account ID'
          )
  } else if (requireTargetAccount) {
    account = z
      .string()
      .trim()
      .min(1, `Enter ${LOCATION_COPY[platform].accountLabel.toLowerCase()}`)
  }

  return account
}

const detailSchema = (
  platform: TCloudPlatform,
  requireTargetAccount: boolean
) =>
  z.object({
    name: z.string().trim().min(1, 'Enter an install name'),
    location: z
      .string()
      .min(1, `Choose ${LOCATION_COPY[platform].label.toLowerCase()}`),
    accountId: accountSchema(platform, requireTargetAccount),
  })

const StepHeading = ({
  title,
  description,
}: {
  title: string
  description: string
}) => (
  <div className="flex flex-col gap-1">
    <Text as="h2" variant="heading">
      {title}
    </Text>
    <Text color="secondary">{description}</Text>
  </div>
)

const StepAction = ({
  readOnly,
  disabled,
  loading,
  label = 'Continue',
  onClick,
}: {
  readOnly: boolean
  disabled?: boolean
  loading?: boolean
  label?: string
  onClick: () => void
}) =>
  readOnly ? null : (
    <div className="flex justify-end">
      <Button
        variant="primary"
        disabled={disabled}
        loading={loading}
        onClick={onClick}
      >
        {label}
      </Button>
    </div>
  )

const groupDescription = (group: IInstallSetupGroup) => {
  if (group.kind === 'all') {
    return 'This group already includes every install automatically.'
  }
  if (group.kind === 'install-ids') {
    return 'This group is managed with explicit install IDs.'
  }
  if (group.kind === 'wildcard') {
    return 'Wildcard selectors cannot be joined by applying labels.'
  }
  const labels = Object.entries(group.labels ?? {})
    .map(([key, value]) => `${key}=${value}`)
    .join(', ')
  return labels ? `Joining applies ${labels}.` : undefined
}

const BranchEnrollment = ({
  branches,
  selectedBranchId,
  selectedGroupId,
  readOnly,
  onSelect,
  onClear,
}: {
  branches: IInstallSetupBranch[]
  selectedBranchId: string
  selectedGroupId: string
  readOnly: boolean
  onSelect: (branchId: string, groupId: string) => void
  onClear: () => void
}) => (
  <div className="flex flex-col gap-3 border-t pt-5">
    <div className="flex flex-wrap items-start justify-between gap-3">
      <span className="flex flex-col gap-1">
        <span className="flex items-center gap-2">
          <Text weight="medium">App branch enrollment</Text>
          <Badge>Optional</Badge>
        </span>
        <Text variant="caption" color="tertiary">
          Join an install group now, or enroll this install later.
        </Text>
      </span>
      {!readOnly && selectedGroupId ? (
        <Button size="sm" variant="ghost" onClick={onClear}>
          Clear enrollment
        </Button>
      ) : null}
    </div>

    {branches.length ? (
      <div className="flex flex-col gap-2">
        {branches.map((branch) => (
          <Disclosure
            key={branch.id}
            title={branch.name}
            icon={<Icon variant="GitBranchIcon" size={16} />}
            status={
              <Badge>
                {branch.groups.length}{' '}
                {branch.groups.length === 1 ? 'group' : 'groups'}
              </Badge>
            }
            defaultOpen={
              branches.length === 1 || selectedBranchId === branch.id
            }
            className="rounded-lg border border-divider"
            headerClassName="px-3"
            contentClassName="flex flex-col gap-2 border-t border-divider p-3"
          >
            {branch.groups.length ? (
              branch.groups.map((group) => (
                <div key={group.id} className="rounded-lg bg-surface-01 p-3">
                  <Radio
                    name="installGroupId"
                    value={group.id}
                    checked={
                      selectedBranchId === branch.id &&
                      selectedGroupId === group.id
                    }
                    disabled={readOnly || group.kind !== 'labels'}
                    onChange={() => onSelect(branch.id, group.id)}
                    label={
                      <span className="flex flex-wrap items-center gap-2">
                        <span>{group.name}</span>
                        {Object.entries(group.labels ?? {}).map(
                          ([key, value]) => (
                            <Badge
                              key={key}
                              variant="code"
                              labelKey={key}
                              labelValue={value}
                            />
                          )
                        )}
                      </span>
                    }
                    description={groupDescription(group)}
                  />
                </div>
              ))
            ) : (
              <Text variant="caption" color="tertiary">
                No install groups in this branch.
              </Text>
            )}
          </Disclosure>
        ))}
      </div>
    ) : (
      <Text variant="caption" color="tertiary">
        This app has no app branches. You can continue without enrollment.
      </Text>
    )}
  </div>
)

const regionOptions = (platform: TCloudPlatform) =>
  cloudRegionsFor(platform).map((region) => ({
    value: region.value,
    textValue: `${region.text} ${region.value}`,
    label: (
      <CloudRegion
        platform={platform}
        region={platform === 'azure' ? undefined : region.value}
        location={platform === 'azure' ? region.value : undefined}
      />
    ),
    description: region.helpText ?? region.value,
  }))

export const InstallSetup = ({
  apps,
  branches,
  inputs,
  platform,
  configurationKey,
  configurationLoading = false,
  configurationReady = false,
  appsLoading = false,
  requireTargetAccount = false,
  pending = false,
  error,
  install,
  persistence,
  onAppChange,
  onBranchChange,
  onClearError,
  onDiscard,
  onSubmit,
}: IInstallSetup) => {
  const [initialDraft] = useState(() =>
    persistence
      ? loadDraft<IInstallSetupDraft>(persistence.orgId, persistence.wizard)
          ?.values
      : undefined
  )
  const [progress, setProgress] = useState(
    initialDraft?.progress ?? { app: false, details: false }
  )
  const form = useForm({
    defaultValues:
      initialDraft?.values ??
      ({
        appId: '',
        name: '',
        location: '',
        accountId: '',
        autoApprove: false as boolean,
        stackOnly: false as boolean,
        labels: [{ key: '', value: '' }],
        inputs: [],
        branchId: '',
        installGroupId: '',
      } satisfies ISetupValues),
  })
  const values = useStore(form.store, (state) => state.values)
  const hydrated = useRef(false)
  const nameInputRef = useRef<HTMLInputElement>(null)
  const [takenName, setTakenName] = useState<string>()
  const [nameFocusRequest, setNameFocusRequest] = useState(0)

  useEffect(() => {
    if (hydrated.current) return
    hydrated.current = true
    if (values.appId) onAppChange(values.appId)
    if (values.branchId) onBranchChange(values.branchId)
  }, [onAppChange, onBranchChange, values.appId, values.branchId])

  useEffect(() => {
    if (!persistence || install) return
    saveDraft<IInstallSetupDraft>(persistence.orgId, persistence.wizard, {
      values,
      progress,
    })
  }, [install, persistence, progress, values])

  useEffect(() => {
    if (!persistence || !install) return
    clearDraft(persistence.orgId, persistence.wizard)
  }, [install, persistence])

  useEffect(() => {
    if (!nameFocusRequest) return
    nameInputRef.current?.focus()
    nameInputRef.current?.select()
  }, [nameFocusRequest])

  useEffect(() => {
    if (!configurationKey) return
    const current = new Map(
      form.state.values.inputs.map((input) => [input.name, input.value])
    )
    form.setFieldValue(
      'inputs',
      inputs.map((input) => ({
        name: input.name,
        value: current.get(input.name) ?? input.defaultValue,
      }))
    )
  }, [configurationKey, form, inputs])

  const detailsValid = detailSchema(platform, requireTargetAccount).safeParse(
    values
  ).success
  const inputsValid = inputs.every((input, index) => {
    if (!input.required || input.type === 'boolean') return true
    return String(values.inputs[index]?.value ?? '').trim().length > 0
  })
  const phase = install?.lifecycle_phase?.phase
  const selectedGroup = branches
    .find((branch) => branch.id === values.branchId)
    ?.groups.find((group) => group.id === values.installGroupId)
  const selectedGroupLabels =
    selectedGroup?.kind === 'labels'
      ? Object.entries(selectedGroup.labels ?? {})
      : []
  const duplicateNameError = isDuplicateInstallNameError(error)
  const nameTaken = !!takenName && values.name.trim() === takenName
  const state: ISetupState = {
    ...progress,
    created: !!install,
    provisioned: !!install && !!phase && phase !== 'provisioning',
    values,
  }

  const descriptor = useMemo<IWizardDescriptor<ISetupState>>(
    () => ({
      steps: [
        {
          id: 'app',
          label: 'App',
          complete: (current) =>
            current.created || (current.app && Boolean(current.values.appId)),
          editable: (current) => !current.created,
          render: ({ readOnly, editing }) => (
            <div className="flex flex-col gap-5">
              <StepHeading
                title="Select app"
                description="Choose the app this install is created from."
              />
              <form.Field
                name="appId"
                validators={{ onBlur: z.string().min(1, 'Choose an app') }}
              >
                {(field) => (
                  <AppSelect
                    field={field}
                    apps={apps}
                    loading={appsLoading}
                    disabled={readOnly}
                    onValueChange={(appId) => {
                      const readiness = apps.find(
                        (app) => app.id === appId
                      )?.readiness
                      form.setFieldValue('branchId', '')
                      form.setFieldValue('installGroupId', '')
                      form.setFieldValue('location', '')
                      form.setFieldValue('inputs', [])
                      form.setFieldValue(
                        'stackOnly',
                        readiness === 'no-components' ||
                          readiness === 'no-component-builds'
                      )
                      onBranchChange(undefined)
                      onAppChange(appId)
                    }}
                  />
                )}
              </form.Field>
              {values.appId ? (
                <form.Field name="installGroupId">
                  {(field) => (
                    <BranchEnrollment
                      branches={branches}
                      selectedBranchId={values.branchId}
                      selectedGroupId={field.state.value}
                      readOnly={readOnly}
                      onSelect={(branchId, groupId) => {
                        form.setFieldValue('branchId', branchId)
                        form.setFieldValue('inputs', [])
                        field.handleChange(groupId)
                        onBranchChange(branchId)
                      }}
                      onClear={() => {
                        form.setFieldValue('branchId', '')
                        form.setFieldValue('inputs', [])
                        field.handleChange('')
                        onBranchChange(undefined)
                      }}
                    />
                  )}
                </form.Field>
              ) : null}
              <StepAction
                readOnly={readOnly || editing}
                disabled={!values.appId}
                onClick={() =>
                  setProgress((current) => ({ ...current, app: true }))
                }
              />
            </div>
          ),
        },
        {
          id: 'details',
          label: 'Install details',
          complete: (current) =>
            current.created ||
            (current.details &&
              detailSchema(platform, requireTargetAccount).safeParse(
                current.values
              ).success),
          editable: (current) => !current.created,
          render: ({ readOnly, editing }) => {
            const copy = LOCATION_COPY[platform]
            return (
              <div className="flex flex-col gap-5">
                <StepHeading
                  title="Configure install"
                  description="Set the install identity, destination, approval behavior, and initial provisioning scope."
                />
                <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
                  <form.Field
                    name="name"
                    validators={{
                      onBlur: z.string().trim().min(1, 'Enter an install name'),
                    }}
                  >
                    {(field) => (
                      <FormInput
                        ref={nameInputRef}
                        field={field}
                        label="Install name"
                        placeholder="acme-production"
                        error={
                          nameTaken
                            ? 'An install with this name already exists'
                            : undefined
                        }
                        disabled={readOnly}
                      />
                    )}
                  </form.Field>
                  <form.Field
                    name="location"
                    validators={{
                      onBlur: z
                        .string()
                        .min(1, `Choose ${copy.label.toLowerCase()}`),
                    }}
                  >
                    {(field) => (
                      <FormSelect
                        field={field}
                        label={copy.label}
                        options={regionOptions(platform)}
                        placeholder={copy.placeholder}
                        searchable
                        disabled={readOnly}
                      />
                    )}
                  </form.Field>
                </div>
                <form.Field
                  name="accountId"
                  validators={{
                    onBlur: accountSchema(platform, requireTargetAccount),
                  }}
                >
                  {(field) => (
                    <FormInput
                      field={field}
                      label={copy.accountLabel}
                      description="The target cloud account for this install. It cannot be changed later."
                      placeholder={copy.accountPlaceholder}
                      disabled={readOnly}
                      optional={!requireTargetAccount}
                    />
                  )}
                </form.Field>
                <div className="flex flex-col gap-4">
                  <form.Field name="autoApprove">
                    {(field) => (
                      <FormSwitch
                        field={field}
                        label="Auto-approve changes"
                        description="Automatically approve and apply future changes without manual confirmation."
                        disabled={readOnly}
                      />
                    )}
                  </form.Field>
                  <form.Field name="stackOnly">
                    {(field) => (
                      <FormSwitch
                        field={field}
                        label="Provision stack and runner only"
                        description="Leave the sandbox and components unprovisioned for now."
                        disabled={readOnly}
                      />
                    )}
                  </form.Field>
                </div>
                <StepAction
                  readOnly={readOnly || editing}
                  disabled={!detailsValid || nameTaken}
                  onClick={() =>
                    setProgress((current) => ({
                      ...current,
                      details: true,
                    }))
                  }
                />
              </div>
            )
          },
        },
        {
          id: 'labels-inputs',
          label: 'Labels and inputs',
          complete: (current) => current.created,
          editable: (current) => !current.created,
          render: ({ readOnly }) => (
            <div className="flex flex-col gap-6">
              <StepHeading
                title="Add labels and inputs"
                description="Labels organize this install. Inputs configure the selected app."
              />
              <form.Field name="labels" mode="array">
                {(labelsField) => (
                  <div className="flex flex-col gap-3">
                    <Text weight="medium">Labels</Text>
                    {selectedGroupLabels.length ? (
                      <div className="flex flex-col gap-2">
                        <Text variant="caption" color="tertiary">
                          Added by the {selectedGroup?.name} install group.
                        </Text>
                        {selectedGroupLabels.map(([key, value]) => (
                          <div
                            key={key}
                            className="grid grid-cols-[1fr_1fr_auto] items-center gap-2"
                          >
                            <Input
                              aria-label={`Install group label ${key} key`}
                              value={key}
                              disabled
                            />
                            <Input
                              aria-label={`Install group label ${key} value`}
                              value={value}
                              disabled
                            />
                            <Badge>Install group</Badge>
                          </div>
                        ))}
                      </div>
                    ) : null}
                    {labelsField.state.value.map((_, index) => (
                      <div
                        key={index}
                        className="grid grid-cols-[1fr_1fr_auto] gap-2"
                      >
                        <form.Field name={`labels[${index}].key`}>
                          {(field) => (
                            <Input
                              aria-label={`Label ${index + 1} key`}
                              placeholder="Key"
                              value={field.state.value}
                              onChange={(event) =>
                                field.handleChange(event.target.value)
                              }
                              disabled={readOnly}
                            />
                          )}
                        </form.Field>
                        <form.Field name={`labels[${index}].value`}>
                          {(field) => (
                            <Input
                              aria-label={`Label ${index + 1} value`}
                              placeholder="Value"
                              value={field.state.value}
                              onChange={(event) =>
                                field.handleChange(event.target.value)
                              }
                              disabled={readOnly}
                            />
                          )}
                        </form.Field>
                        <Button
                          size="sm"
                          variant="ghost"
                          iconOnly
                          aria-label={`Remove label ${index + 1}`}
                          disabled={readOnly}
                          onClick={() => labelsField.removeValue(index)}
                        >
                          <Icon variant="TrashIcon" size={16} />
                        </Button>
                      </div>
                    ))}
                    {!readOnly ? (
                      <Button
                        className="self-start"
                        onClick={() =>
                          labelsField.pushValue({ key: '', value: '' })
                        }
                      >
                        Add label
                      </Button>
                    ) : null}
                  </div>
                )}
              </form.Field>

              <div className="flex flex-col gap-4 border-t pt-5">
                <Text weight="medium">App inputs</Text>
                {configurationLoading ? (
                  <>
                    <Text loading loadingWidth={18} />
                    <Text loading loadingWidth={24} />
                  </>
                ) : !configurationReady ? (
                  <Banner theme="error" heading="App configuration unavailable">
                    No active app configuration is available for this selection.
                  </Banner>
                ) : inputs.length ? (
                  inputs.map((input, index) => (
                    <form.Field
                      key={input.name}
                      name={`inputs[${index}].value`}
                      validators={{
                        onBlur:
                          input.required && input.type !== 'boolean'
                            ? z
                                .union([z.string(), z.boolean()])
                                .refine(
                                  (value) => String(value).trim().length > 0,
                                  `Enter ${input.label.toLowerCase()}`
                                )
                            : undefined,
                      }}
                    >
                      {(field) =>
                        input.type === 'boolean' ? (
                          <FormSwitch
                            field={field}
                            label={input.label}
                            description={input.description}
                            disabled={readOnly}
                          />
                        ) : (
                          <FormInput
                            field={field}
                            label={input.label}
                            description={input.description}
                            type={input.type}
                            optional={!input.required}
                            disabled={readOnly}
                          />
                        )
                      }
                    </form.Field>
                  ))
                ) : (
                  <Text variant="caption" color="tertiary">
                    This app configuration has no vendor inputs.
                  </Text>
                )}
              </div>
              {duplicateNameError ? (
                <Banner
                  theme="error"
                  heading="Install name already in use"
                  actions={
                    <Button
                      variant="secondary"
                      onClick={() => {
                        setTakenName(values.name.trim())
                        setNameFocusRequest((count) => count + 1)
                        setProgress((current) => ({
                          ...current,
                          details: false,
                        }))
                        onClearError?.()
                      }}
                    >
                      Edit install name
                    </Button>
                  }
                >
                  An install with this name already exists. Choose a different
                  name and try again.
                </Banner>
              ) : (
                <FormErrorBanner
                  error={error}
                  fallback="Unable to create the install"
                />
              )}
              <StepAction
                readOnly={readOnly}
                label="Create install"
                loading={pending}
                disabled={
                  !configurationReady || !inputsValid || pending || !!install
                }
                onClick={() => onSubmit(values)}
              />
            </div>
          ),
        },
        {
          id: 'provision',
          label: 'Provision',
          complete: (current) => current.provisioned,
          render: ({ state: current }) => (
            <div className="flex flex-col gap-5">
              <StepHeading
                title="Provision install"
                description="Follow the install workflow until provisioning finishes."
              />
              <div className="flex min-h-44 flex-col items-center justify-center gap-3 rounded-lg bg-surface-01 p-6 text-center">
                <Status
                  status={
                    current.provisioned
                      ? 'success'
                      : (install?.runner_status ?? phase ?? 'provisioning')
                  }
                  label={
                    current.provisioned
                      ? 'Provision complete'
                      : 'Provisioning install'
                  }
                />
                <Text variant="caption" color="secondary">
                  {install?.runner_status_description ??
                    'The provision workflow is running.'}
                </Text>
              </div>
            </div>
          ),
        },
      ],
    }),
    [
      apps,
      appsLoading,
      branches,
      configurationLoading,
      configurationReady,
      detailsValid,
      duplicateNameError,
      error,
      form,
      inputs,
      inputsValid,
      install,
      onAppChange,
      onBranchChange,
      onClearError,
      onSubmit,
      pending,
      platform,
      nameTaken,
      requireTargetAccount,
      selectedGroup,
      selectedGroupLabels,
      values,
    ]
  )

  return (
    <form
      className="mx-auto flex w-full max-w-4xl flex-col gap-4"
      autoComplete="off"
      noValidate
      onSubmit={(event) => event.preventDefault()}
    >
      <Wizard
        descriptor={descriptor}
        state={state}
        discardAction={
          onDiscard && !install
            ? {
                children: 'Discard setup',
                onClick: () => {
                  if (persistence) {
                    clearDraft(persistence.orgId, persistence.wizard)
                  }
                  onDiscard()
                },
              }
            : undefined
        }
      />
    </form>
  )
}
