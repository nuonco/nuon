import { useMemo, useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { Badge } from '../../atoms/Badge'
import { Button } from '../../atoms/Button'
import { Icon } from '../../atoms/Icon'
import { Input } from '../../atoms/Input'
import { Radio } from '../../atoms/Radio'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import { Disclosure } from '../../molecules/Disclosure'
import { FormInput } from '../../molecules/FormInput'
import { FormSelect } from '../../molecules/FormSelect'
import { FormSwitch } from '../../molecules/FormSwitch'
import { CloudRegion } from '../../molecules/CloudRegion'
import { AppSelect, type IAppSelectItem } from '../../organisms/AppSelect'
import type { IWizardDescriptor } from '../../../utils/wizard'
import { cloudRegionsFor } from '../../../utils/cloud-regions'
import type { TCloudPlatform } from '@/types'
import { Wizard } from './Wizard'

export default {
  title: 'lite/templates/Wizard/Install setup',
}

const APPS: IAppSelectItem[] = [
  {
    id: 'app98e2wpzdxwoey393edtqj45',
    name: 'Payments API',
    platform: 'aws',
    source: 'acme/payments',
    updatedLabel: 'synced 8 minutes ago',
  },
  {
    id: 'app7fplr1up5atx5zpxotbabm',
    name: 'Support portal',
    platform: 'aws',
    source: 'acme/support-portal',
    updatedLabel: 'synced yesterday',
  },
  {
    id: 'appk933tcyzji01s7us3aeo3x',
    name: 'Analytics warehouse',
    platform: 'azure',
    source: 'acme/analytics',
    updatedLabel: 'synced 3 days ago',
    readiness: 'no-components',
  },
  {
    id: 'appm41s7us3aeo3xk933tcyz',
    name: 'Realtime ingest',
    platform: 'gcp',
    source: 'acme/realtime-ingest',
    updatedLabel: 'synced 25 minutes ago',
  },
]

const REGION_COPY: Record<
  TCloudPlatform,
  { label: string; placeholder: string; error: string }
> = {
  aws: {
    label: 'AWS region',
    placeholder: 'Choose an AWS region',
    error: 'Choose an AWS region',
  },
  azure: {
    label: 'Azure location',
    placeholder: 'Choose an Azure location',
    error: 'Choose an Azure location',
  },
  gcp: {
    label: 'GCP region',
    placeholder: 'Choose a GCP region',
    error: 'Choose a GCP region',
  },
  unknown: {
    label: 'Region',
    placeholder: 'Choose a region',
    error: 'Choose a region',
  },
}

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

type TEnrollmentGroupKind = 'labels' | 'all' | 'install-ids' | 'wildcard'

interface IEnrollmentGroup {
  id: string
  name: string
  kind: TEnrollmentGroupKind
  labels?: Record<string, string>
}

interface IEnrollmentBranch {
  id: string
  name: string
  groups: IEnrollmentGroup[]
}

const APP_BRANCHES: Record<string, IEnrollmentBranch[]> = {
  app98e2wpzdxwoey393edtqj45: [
    {
      id: 'branch_payments_main',
      name: 'main',
      groups: [
        {
          id: 'group_payments_production',
          name: 'Production',
          kind: 'labels',
          labels: { env: 'production', tier: 'critical' },
        },
        {
          id: 'group_payments_preview',
          name: 'Preview',
          kind: 'wildcard',
          labels: { env: 'preview', pull_request: '*' },
        },
      ],
    },
    {
      id: 'branch_payments_release',
      name: 'release',
      groups: [
        {
          id: 'group_payments_staging',
          name: 'Staging',
          kind: 'labels',
          labels: { env: 'staging' },
        },
        {
          id: 'group_payments_all',
          name: 'Every install',
          kind: 'all',
        },
      ],
    },
  ],
  app7fplr1up5atx5zpxotbabm: [
    {
      id: 'branch_support_main',
      name: 'main',
      groups: [
        {
          id: 'group_support_production',
          name: 'Production',
          kind: 'labels',
          labels: { env: 'production' },
        },
      ],
    },
    {
      id: 'branch_support_next',
      name: 'next',
      groups: [],
    },
  ],
  appk933tcyzji01s7us3aeo3x: [
    {
      id: 'branch_analytics_main',
      name: 'main',
      groups: [
        {
          id: 'group_analytics_pinned',
          name: 'Pinned installs',
          kind: 'install-ids',
        },
      ],
    },
  ],
  appm41s7us3aeo3xk933tcyz: [],
}

const installDetailsSchema = (platform: TCloudPlatform) =>
  z.object({
    name: z.string().trim().min(1, 'Enter an install name'),
    region: z.string().min(1, REGION_COPY[platform].error),
    awsAccountId: z
      .string()
      .regex(/^[0-9]{12}$/, 'Enter a 12-digit AWS account ID'),
  })

const inputsSchema = z.object({
  hostname: z.string().trim().min(1, 'Enter a hostname'),
  replicas: z.string().regex(/^[0-9]+$/, 'Enter a whole number'),
})

interface IInstallSetupValues {
  appId: string
  name: string
  region: string
  awsAccountId: string
  autoApprove: boolean
  stackOnly: boolean
  labels: Array<{ key: string; value: string }>
  hostname: string
  replicas: string
  metricsEnabled: boolean
  branchId: string
  installGroupId: string
}

interface IInstallSetupProgress {
  app: boolean
  details: boolean
  inputs: boolean
  provisioned: boolean
}

interface IInstallSetupState extends IInstallSetupProgress {
  values: IInstallSetupValues
}

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

const StepActions = ({
  readOnly,
  disabled,
  onContinue,
}: {
  readOnly: boolean
  disabled?: boolean
  onContinue: () => void
}) =>
  readOnly ? null : (
    <div className="flex justify-end">
      <Button variant="primary" disabled={disabled} onClick={onContinue}>
        Continue
      </Button>
    </div>
  )

const enrollmentDescription = (group: IEnrollmentGroup) => {
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
  branches: IEnrollmentBranch[]
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
              branch.groups.map((group) => {
                const selectable = group.kind === 'labels'
                return (
                  <div key={group.id} className="rounded-lg bg-surface-01 p-3">
                    <Radio
                      name="installGroupId"
                      value={group.id}
                      checked={
                        selectedBranchId === branch.id &&
                        selectedGroupId === group.id
                      }
                      disabled={readOnly || !selectable}
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
                      description={enrollmentDescription(group)}
                    />
                  </div>
                )
              })
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

const InstallSetupStory = () => {
  const [progress, setProgress] = useState<IInstallSetupProgress>({
    app: false,
    details: false,
    inputs: false,
    provisioned: false,
  })
  const form = useForm({
    defaultValues: {
      appId: '',
      name: '',
      region: '',
      awsAccountId: '',
      autoApprove: false,
      stackOnly: false,
      labels: [{ key: '', value: '' }],
      hostname: '',
      replicas: '2',
      metricsEnabled: true,
      branchId: '',
      installGroupId: '',
    } satisfies IInstallSetupValues,
  })
  const values = useStore(form.store, (state) => state.values)
  const state: IInstallSetupState = { ...progress, values }
  const selectedApp = APPS.find((app) => app.id === values.appId)
  const platform = selectedApp?.platform ?? 'unknown'
  const detailsValid = installDetailsSchema(platform).safeParse(values).success
  const inputsValid = inputsSchema.safeParse(values).success
  const selectedBranches = APP_BRANCHES[values.appId] ?? []
  const regionField = REGION_COPY[platform]

  const complete = (key: keyof IInstallSetupProgress) => {
    setProgress((current) => ({ ...current, [key]: true }))
  }

  const descriptor = useMemo<IWizardDescriptor<IInstallSetupState>>(
    () => ({
      steps: [
        {
          id: 'app',
          label: 'App',
          complete: (current) => current.app && Boolean(current.values.appId),
          render: ({ readOnly }) => (
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
                    apps={APPS}
                    disabled={readOnly}
                    onValueChange={() => {
                      form.setFieldValue('branchId', '')
                      form.setFieldValue('installGroupId', '')
                      form.setFieldValue('region', '')
                    }}
                  />
                )}
              </form.Field>
              {values.appId ? (
                <form.Field name="installGroupId">
                  {(field) => (
                    <BranchEnrollment
                      branches={selectedBranches}
                      selectedBranchId={values.branchId}
                      selectedGroupId={field.state.value}
                      readOnly={readOnly}
                      onSelect={(branchId, groupId) => {
                        form.setFieldValue('branchId', branchId)
                        field.handleChange(groupId)
                      }}
                      onClear={() => {
                        form.setFieldValue('branchId', '')
                        field.handleChange('')
                      }}
                    />
                  )}
                </form.Field>
              ) : null}
              <StepActions
                readOnly={readOnly}
                disabled={!values.appId}
                onContinue={() => complete('app')}
              />
            </div>
          ),
        },
        {
          id: 'details',
          label: 'Install details',
          complete: (current) =>
            current.details &&
            installDetailsSchema(
              APPS.find((app) => app.id === current.values.appId)?.platform ??
                'unknown'
            ).safeParse(current.values).success,
          render: ({ readOnly }) => (
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
                      field={field}
                      label="Install name"
                      placeholder="acme-production"
                      disabled={readOnly}
                    />
                  )}
                </form.Field>
                <form.Field
                  name="region"
                  validators={{
                    onBlur: z.string().min(1, regionField.error),
                  }}
                >
                  {(field) => (
                    <FormSelect
                      field={field}
                      label={regionField.label}
                      options={regionOptions(platform)}
                      placeholder={regionField.placeholder}
                      searchable
                      disabled={readOnly}
                    />
                  )}
                </form.Field>
              </div>
              <form.Field
                name="awsAccountId"
                validators={{
                  onBlur: z
                    .string()
                    .regex(/^[0-9]{12}$/, 'Enter a 12-digit AWS account ID'),
                }}
              >
                {(field) => (
                  <FormInput
                    field={field}
                    label="AWS account ID"
                    description="The AWS account this install deploys into. It cannot be changed later."
                    placeholder="123456789012"
                    inputMode="numeric"
                    disabled={readOnly}
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
              <StepActions
                readOnly={readOnly}
                disabled={!detailsValid}
                onContinue={() => complete('details')}
              />
            </div>
          ),
        },
        {
          id: 'labels-inputs',
          label: 'Labels and inputs',
          complete: (current) =>
            current.inputs && inputsSchema.safeParse(current.values).success,
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
                    {labelsField.state.value.length ? (
                      labelsField.state.value.map((_, index) => (
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
                      ))
                    ) : (
                      <Text variant="caption" color="tertiary">
                        No labels yet
                      </Text>
                    )}
                    {readOnly ? null : (
                      <Button
                        className="self-start"
                        onClick={() =>
                          labelsField.pushValue({ key: '', value: '' })
                        }
                      >
                        Add label
                      </Button>
                    )}
                  </div>
                )}
              </form.Field>
              <div className="flex flex-col gap-4 border-t pt-5">
                <Text weight="medium">App inputs</Text>
                <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
                  <form.Field
                    name="hostname"
                    validators={{
                      onBlur: z.string().trim().min(1, 'Enter a hostname'),
                    }}
                  >
                    {(field) => (
                      <FormInput
                        field={field}
                        label="Public hostname"
                        placeholder="payments.example.com"
                        disabled={readOnly}
                      />
                    )}
                  </form.Field>
                  <form.Field
                    name="replicas"
                    validators={{
                      onBlur: z
                        .string()
                        .regex(/^[0-9]+$/, 'Enter a whole number'),
                    }}
                  >
                    {(field) => (
                      <FormInput
                        field={field}
                        label="Replica count"
                        type="number"
                        min="1"
                        disabled={readOnly}
                      />
                    )}
                  </form.Field>
                </div>
                <form.Field name="metricsEnabled">
                  {(field) => (
                    <FormSwitch
                      field={field}
                      label="Enable metrics"
                      description="Expose application metrics to the configured monitoring integration."
                      disabled={readOnly}
                    />
                  )}
                </form.Field>
              </div>
              <StepActions
                readOnly={readOnly}
                disabled={!inputsValid}
                onContinue={() => complete('inputs')}
              />
            </div>
          ),
        },
        {
          id: 'provision',
          label: 'Provision',
          complete: (current) => current.provisioned,
          render: ({ state: current, readOnly }) => (
            <div className="flex flex-col gap-5">
              <StepHeading
                title="Provision install"
                description="Follow the install workflow until provisioning finishes."
              />
              <div className="flex min-h-44 flex-col items-center justify-center gap-3 rounded-lg bg-surface-01 p-6 text-center">
                <Status
                  status={current.provisioned ? 'success' : 'provisioning'}
                  label={
                    current.provisioned
                      ? 'Provision complete'
                      : 'Provision workflow placeholder'
                  }
                />
                <Text variant="caption" color="secondary">
                  Workflow steps, retries, failures, and panels will be added
                  here.
                </Text>
              </div>
              {readOnly || current.provisioned ? null : (
                <div className="flex justify-end">
                  <Button
                    variant="primary"
                    onClick={() => complete('provisioned')}
                  >
                    Complete mock provision
                  </Button>
                </div>
              )}
            </div>
          ),
        },
      ],
    }),
    [detailsValid, form, inputsValid, platform, regionField, values]
  )

  const reset = () => {
    form.reset()
    setProgress({
      app: false,
      details: false,
      inputs: false,
      provisioned: false,
    })
  }

  return (
    <form
      className="mx-auto flex w-full max-w-4xl flex-col gap-4"
      autoComplete="off"
      noValidate
      onSubmit={(event) => event.preventDefault()}
    >
      <div className="flex justify-end">
        <Button variant="ghost" onClick={reset}>
          Reset mock data
        </Button>
      </div>
      <Wizard descriptor={descriptor} state={state} />
    </form>
  )
}

export const Overview = () => (
  <ComponentDocs
    name="Install setup wizard"
    tier="template"
    summary="An interactive four-step fixture for designing the Lite install setup flow."
    use={[
      'Enter mock values and continue through app and optional install-group selection, install details, labels and inputs, and provision.',
    ]}
    avoid={[
      'Do not treat this fixture as API wiring or the final provision workflow.',
      'Do not add production dashboard components to Lite.',
    ]}
    rules={[
      'Every input is backed by TanStack Form and validated with Zod where required.',
      'Completed steps remain reachable and render their controls read-only.',
      'Selecting an app reveals its branches and install groups in the same step.',
      'Install-group enrollment is optional and never blocks Continue.',
      'Only concrete label-selector groups can be joined during setup; automatic, install-ID and wildcard groups remain visible but disabled.',
      'Each step owns the single primary action that advances it, so the flow never shows two primaries at once.',
    ]}
    props={[]}
  />
)

export const Interactive = () => <InstallSetupStory />
