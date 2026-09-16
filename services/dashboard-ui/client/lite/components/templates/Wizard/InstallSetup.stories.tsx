import { useMemo, useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { z } from 'zod'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { Button } from '../../atoms/Button'
import { Icon } from '../../atoms/Icon'
import { Input } from '../../atoms/Input'
import { Status } from '../../atoms/Status'
import { Text } from '../../atoms/Text'
import { FormInput } from '../../molecules/FormInput'
import { FormRadioGroup } from '../../molecules/FormRadioGroup'
import { FormSelect } from '../../molecules/FormSelect'
import { FormSwitch } from '../../molecules/FormSwitch'
import { AppSelect, type IAppSelectItem } from '../../organisms/AppSelect'
import type { IWizardDescriptor } from '../../../utils/wizard'
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

const AWS_REGIONS = [
  {
    value: 'us-east-1',
    label: 'US East (N. Virginia)',
    description: 'us-east-1',
  },
  {
    value: 'us-west-2',
    label: 'US West (Oregon)',
    description: 'us-west-2',
  },
  {
    value: 'eu-west-1',
    label: 'Europe (Ireland)',
    description: 'eu-west-1',
  },
]

const BRANCHES = [
  {
    value: 'branch_main',
    label: 'main · Production',
    description: 'Applies env=production and tier=critical',
  },
  {
    value: 'branch_release',
    label: 'release · Staging',
    description: 'Applies env=staging',
  },
  {
    value: 'none',
    label: 'No app branch',
    description: 'Create the install without a deployment plan assignment.',
  },
]

const installDetailsSchema = z.object({
  name: z.string().trim().min(1, 'Enter an install name'),
  region: z.string().min(1, 'Choose an AWS region'),
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
}

interface IInstallSetupProgress {
  app: boolean
  details: boolean
  inputs: boolean
  branch: boolean
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

const InstallSetupStory = () => {
  const [progress, setProgress] = useState<IInstallSetupProgress>({
    app: false,
    details: false,
    inputs: false,
    branch: false,
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
    } satisfies IInstallSetupValues,
  })
  const values = useStore(form.store, (state) => state.values)
  const state: IInstallSetupState = { ...progress, values }
  const detailsValid = installDetailsSchema.safeParse(values).success
  const inputsValid = inputsSchema.safeParse(values).success

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
                  <AppSelect field={field} apps={APPS} disabled={readOnly} />
                )}
              </form.Field>
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
            installDetailsSchema.safeParse(current.values).success,
          render: ({ readOnly }) => (
            <div className="flex flex-col gap-5">
              <StepHeading
                title="Configure install"
                description="Set the install identity, AWS destination, approval behavior, and initial provisioning scope."
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
                    onBlur: z.string().min(1, 'Choose an AWS region'),
                  }}
                >
                  {(field) => (
                    <FormSelect
                      field={field}
                      label="AWS region"
                      options={AWS_REGIONS}
                      placeholder="Choose an AWS region"
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
          id: 'app-branch',
          label: 'App branch',
          complete: (current) =>
            current.branch && Boolean(current.values.branchId),
          render: ({ readOnly }) => (
            <div className="flex flex-col gap-5">
              <StepHeading
                title="Assign to app branch"
                description="Optionally add this install to an app branch deployment plan."
              />
              <form.Field name="branchId">
                {(field) => (
                  <FormRadioGroup
                    field={field}
                    label="App branch"
                    options={BRANCHES}
                    disabled={readOnly}
                  />
                )}
              </form.Field>
              <StepActions
                readOnly={readOnly}
                disabled={!values.branchId}
                onContinue={() => complete('branch')}
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
    [detailsValid, form, inputsValid, values]
  )

  const reset = () => {
    form.reset()
    setProgress({
      app: false,
      details: false,
      inputs: false,
      branch: false,
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
    summary="An interactive five-step fixture for designing the Lite install setup flow."
    use={[
      'Enter mock values and continue through app selection, install details, labels and inputs, app branch assignment, and provision.',
    ]}
    avoid={[
      'Do not treat this fixture as API wiring or the final provision workflow.',
      'Do not add production dashboard components to Lite.',
    ]}
    rules={[
      'Every input is backed by TanStack Form and validated with Zod where required.',
      'Completed steps remain reachable and render their controls read-only.',
      'No app branch is an explicit valid choice because assignment is optional.',
      'Each step owns the single primary action that advances it, so the flow never shows two primaries at once.',
    ]}
    props={[]}
  />
)

export const Interactive = () => <InstallSetupStory />
