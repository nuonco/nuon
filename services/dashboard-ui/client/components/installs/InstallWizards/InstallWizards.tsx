import { useState, type ReactNode } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { FormInput } from '@/components/common/form/FormInput'
import { FormRadioGroup } from '@/components/common/form/FormRadioGroup'
import { Input } from '@/components/common/form/Input'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { cn } from '@/utils/classnames'
import { installSetupSchema, type TInstallSetupValues } from './schema'

export type TStackOwnership = 'customer' | 'nuon'
export type TInstallSetupStep = 'source' | 'details' | 'stack' | 'provision'
export type TInstallDeprovisionStep =
  | 'confirm'
  | 'teardown'
  | 'stack'
  | 'complete'

export interface IInstallFlowProgressItem {
  id: string
  label: string
  status: string
  description?: string
}

interface IWizardStep<TStep extends string> {
  id: TStep
  label: string
}

interface IInstallWizardPage<TStep extends string> {
  title: string
  description: string
  steps: IWizardStep<TStep>[]
  currentStep: TStep
  status?: string
  children: ReactNode
  onStepSelect: (step: TStep) => void
  onBack?: () => void
  onNext?: () => void
  nextLabel: string
  nextDisabled?: boolean
  nextDisabledReason?: string
  nextVariant?: 'primary' | 'danger'
}

const InstallWizardPage = <TStep extends string>({
  title,
  description,
  steps,
  currentStep,
  status,
  children,
  onStepSelect,
  onBack,
  onNext,
  nextLabel,
  nextDisabled = false,
  nextDisabledReason,
  nextVariant = 'primary',
}: IInstallWizardPage<TStep>) => {
  const currentIndex = Math.max(
    0,
    steps.findIndex((step) => step.id === currentStep)
  )

  return (
    <PageLayout>
      <SectionHeader
        variant="page"
        title={title}
        description={description}
        status={status ? <Status status={status} variant="badge" /> : undefined}
      />
      <PageContent>
        <PageSection>
          <div className="mx-auto flex w-full max-w-4xl flex-col gap-6">
            <ol
              aria-label={`${title} steps`}
              className="grid grid-cols-1 gap-2 border-b pb-4 sm:grid-cols-2 lg:grid-cols-4"
            >
              {steps.map((step, index) => {
                const selected = step.id === currentStep
                const available = index <= currentIndex

                return (
                  <li key={step.id}>
                    <Button
                      variant={selected ? 'secondary' : 'ghost'}
                      size="sm"
                      className="w-full justify-start"
                      aria-current={selected ? 'step' : undefined}
                      disabled={!available}
                      tooltipProps={
                        !available
                          ? { tipContent: 'Complete the current step first' }
                          : undefined
                      }
                      onClick={() => onStepSelect(step.id)}
                    >
                      <span
                        className={cn(
                          'mr-2 flex size-5 items-center justify-center rounded-full border font-mono text-xs',
                          selected &&
                            'border-primary-600 bg-primary-600 text-white'
                        )}
                      >
                        {index + 1}
                      </span>
                      {step.label}
                    </Button>
                  </li>
                )
              })}
            </ol>

            <Card className="min-h-96">
              <div className="flex flex-1 flex-col gap-6">{children}</div>
              <div className="flex flex-wrap items-center justify-between gap-3 border-t pt-4">
                <Button type="button" onClick={onBack} disabled={!onBack}>
                  Back
                </Button>
                <Button
                  type="button"
                  variant={nextVariant}
                  onClick={onNext}
                  disabled={nextDisabled || !onNext}
                  tooltipProps={
                    nextDisabled && nextDisabledReason
                      ? { tipContent: nextDisabledReason }
                      : undefined
                  }
                >
                  {nextLabel}
                </Button>
              </div>
            </Card>
          </div>
        </PageSection>
      </PageContent>
    </PageLayout>
  )
}

const BannerCopy = ({
  heading,
  children,
}: {
  heading: string
  children: ReactNode
}) => (
  <div className="flex flex-col gap-1">
    <Text weight="strong">{heading}</Text>
    <Text variant="body">{children}</Text>
  </div>
)

const ProgressList = ({ items }: { items: IInstallFlowProgressItem[] }) => (
  <ol className="flex flex-col divide-y">
    {items.map((item) => (
      <li
        key={item.id}
        className="flex flex-wrap items-center justify-between gap-3 py-4 first:pt-0 last:pb-0"
      >
        <div className="flex flex-col gap-1">
          <Text weight="strong">{item.label}</Text>
          {item.description ? (
            <Text variant="subtext" theme="neutral">
              {item.description}
            </Text>
          ) : null}
        </div>
        <Status status={item.status} />
      </li>
    ))}
  </ol>
)

const setupSteps: IWizardStep<TInstallSetupStep>[] = [
  { id: 'source', label: 'Choose source' },
  { id: 'details', label: 'Install details' },
  { id: 'stack', label: 'Manage stack' },
  { id: 'provision', label: 'Provision' },
]

export interface IInstallSetupWizard {
  appName?: string
  branchName?: string
  error?: string
  initialInstallName?: string
  initialRegion?: string
  initialRoleArn?: string
  initialStackOwnership?: TStackOwnership
  initialStep?: TInstallSetupStep
  onComplete?: () => void
  progress?: IInstallFlowProgressItem[]
}

export const InstallSetupWizard = ({
  appName = 'payments',
  branchName = 'main',
  error,
  initialInstallName = 'production',
  initialRegion = 'us-west-2',
  initialRoleArn = '',
  initialStackOwnership = 'customer',
  initialStep = 'source',
  onComplete,
  progress = [],
}: IInstallSetupWizard) => {
  const initialIndex = Math.max(
    0,
    setupSteps.findIndex((step) => step.id === initialStep)
  )
  const [stepIndex, setStepIndex] = useState(initialIndex)
  const step = setupSteps[stepIndex]!
  const form = useForm({
    defaultValues: {
      installName: initialInstallName,
      region: initialRegion,
      roleArn: initialRoleArn,
      stackOwnership: initialStackOwnership,
    } as TInstallSetupValues,
    validators: {
      onMount: installSetupSchema,
      onChange: installSetupSchema,
    },
    onSubmit: () => setStepIndex(3),
  })
  const values = useStore(form.store, (state) => state.values)
  const detailsReady =
    values.installName.trim().length > 0 && values.region.trim().length > 0
  const stackReady =
    values.stackOwnership === 'customer' || values.roleArn.trim().length > 0
  const provisionComplete =
    progress.length > 0 && progress.every((item) => item.status === 'success')

  const nextLabel =
    step.id === 'source'
      ? 'Enter install details'
      : step.id === 'details'
        ? 'Choose stack management'
        : step.id === 'stack'
          ? 'Provision install'
          : 'View install'

  const nextDisabled =
    (step.id === 'details' && !detailsReady) ||
    (step.id === 'stack' && !stackReady) ||
    (step.id === 'provision' && !provisionComplete)

  const nextDisabledReason =
    step.id === 'details'
      ? 'Cannot continue — enter an install name and AWS region'
      : step.id === 'stack'
        ? 'Cannot provision — connect an IAM role'
        : step.id === 'provision'
          ? 'Cannot view install — provisioning is not complete'
          : undefined

  const handleNext = () => {
    if (step.id === 'stack') {
      form.handleSubmit()
      return
    }
    if (step.id === 'provision') {
      onComplete?.()
      return
    }
    setStepIndex((index) => Math.min(index + 1, setupSteps.length - 1))
  }

  return (
    <form
      className="flex min-h-0 flex-1"
      autoComplete="off"
      noValidate
      onSubmit={(event) => event.preventDefault()}
    >
      <InstallWizardPage
        title="Set up an install"
        description="Create an install and keep its first provision in one guided flow."
        steps={setupSteps}
        currentStep={step.id}
        status={step.id === 'provision' ? 'provisioning' : undefined}
        onStepSelect={(stepId) =>
          setStepIndex(setupSteps.findIndex((item) => item.id === stepId))
        }
        onBack={
          stepIndex > 0 ? () => setStepIndex((index) => index - 1) : undefined
        }
        onNext={handleNext}
        nextLabel={nextLabel}
        nextDisabled={nextDisabled}
        nextDisabledReason={nextDisabledReason}
      >
        {step.id === 'source' ? (
          <>
            <SectionHeader
              title="Choose the app source"
              description="Select the branch and app config this install should follow."
            />
            <Card className="!gap-4 !p-4">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <Text weight="strong">{appName}</Text>
                <Status status="active">Config ready</Status>
              </div>
              <Text family="mono" variant="subtext" theme="neutral" nowrap>
                {branchName}
              </Text>
              <Text variant="subtext" theme="neutral">
                New branch runs will update this install through its deployment
                plan.
              </Text>
            </Card>
          </>
        ) : null}

        {step.id === 'details' ? (
          <>
            <SectionHeader
              title="Enter install details"
              description="Set the cloud location and required app inputs."
            />
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <form.Field name="installName">
                {(field) => (
                  <FormInput
                    field={field}
                    id="install-name"
                    labelProps={{ labelText: 'Install name' }}
                  />
                )}
              </form.Field>
              <form.Field name="region">
                {(field) => (
                  <FormInput
                    field={field}
                    id="install-region"
                    labelProps={{ labelText: 'AWS region' }}
                  />
                )}
              </form.Field>
            </div>
          </>
        ) : null}

        {step.id === 'stack' ? (
          <>
            <SectionHeader
              title="Choose who manages the stack"
              description="Nuon can create and destroy the install stack after an IAM role is connected."
            />
            <form.Field name="stackOwnership">
              {(field) => (
                <FormRadioGroup
                  field={field}
                  label="Stack management"
                  options={[
                    {
                      value: 'customer',
                      label: (
                        <span className="flex flex-col gap-1">
                          <Text weight="strong">
                            Customer manages the stack
                          </Text>
                          <Text variant="subtext" theme="neutral">
                            Create and destroy the CloudFormation stack from the
                            customer AWS account.
                          </Text>
                        </span>
                      ),
                    },
                    {
                      value: 'nuon',
                      label: (
                        <span className="flex flex-col gap-1">
                          <Text weight="strong">Nuon manages the stack</Text>
                          <Text variant="subtext" theme="neutral">
                            Connect an IAM role so Nuon can create and destroy
                            the stack.
                          </Text>
                        </span>
                      ),
                    },
                  ]}
                />
              )}
            </form.Field>

            {values.stackOwnership === 'nuon' ? (
              <>
                <form.Field name="roleArn">
                  {(field) => (
                    <FormInput
                      field={field}
                      id="install-role-arn"
                      labelProps={{ labelText: 'IAM role ARN' }}
                      placeholder="arn:aws:iam::123456789012:role/nuon-stack-manager"
                    />
                  )}
                </form.Field>
                {stackReady ? (
                  <Banner theme="success">
                    <BannerCopy heading="IAM role connected">
                      Nuon will create the stack before provisioning the runner.
                    </BannerCopy>
                  </Banner>
                ) : null}
              </>
            ) : (
              <Banner theme="info">
                <BannerCopy heading="Create the stack in AWS">
                  Download the CloudFormation template, create the stack in the
                  customer account, then return here to verify it.
                </BannerCopy>
              </Banner>
            )}
          </>
        ) : null}

        {step.id === 'provision' ? (
          <>
            <SectionHeader
              title="Provision the install"
              description="Keep this page open while the stack and install services come online."
            />
            {error ? (
              <Banner theme="error">
                <BannerCopy heading="Install provision failed">
                  {error}
                </BannerCopy>
              </Banner>
            ) : null}
            <ProgressList items={progress} />
          </>
        ) : null}
      </InstallWizardPage>
    </form>
  )
}

const deprovisionSteps: IWizardStep<TInstallDeprovisionStep>[] = [
  { id: 'confirm', label: 'Confirm' },
  { id: 'teardown', label: 'Teardown services' },
  { id: 'stack', label: 'Destroy stack' },
  { id: 'complete', label: 'Complete' },
]

export interface IInstallDeprovisionWizard {
  error?: string
  initialStackOwnership?: TStackOwnership
  initialStep?: TInstallDeprovisionStep
  installName?: string
  onComplete?: () => void
  progress?: IInstallFlowProgressItem[]
  stackDestroying?: boolean
}

export const InstallDeprovisionWizard = ({
  error,
  initialStackOwnership = 'customer',
  initialStep = 'confirm',
  installName = 'production',
  onComplete,
  progress = [],
  stackDestroying = false,
}: IInstallDeprovisionWizard) => {
  const initialIndex = Math.max(
    0,
    deprovisionSteps.findIndex((step) => step.id === initialStep)
  )
  const [stepIndex, setStepIndex] = useState(initialIndex)
  const [confirmation, setConfirmation] = useState('')
  const step = deprovisionSteps[stepIndex]!
  const confirmationValid = confirmation === installName

  const nextLabel =
    step.id === 'confirm'
      ? 'Deprovision install'
      : step.id === 'teardown'
        ? 'Destroy stack'
        : step.id === 'stack'
          ? initialStackOwnership === 'nuon'
            ? stackDestroying
              ? 'Destroying stack'
              : 'Destroy stack'
            : 'Mark stack destroyed'
          : 'View activity'

  const nextDisabled =
    (step.id === 'confirm' && !confirmationValid) ||
    (step.id === 'teardown' &&
      !(
        progress.length > 0 &&
        progress.every((item) => item.status === 'success')
      )) ||
    stackDestroying

  const nextDisabledReason =
    step.id === 'teardown'
      ? 'Cannot destroy stack — install teardown is not complete'
      : stackDestroying
        ? 'Stack destruction is in progress'
        : undefined

  const handleNext = () => {
    if (step.id === 'complete') {
      onComplete?.()
      return
    }
    setStepIndex((index) => Math.min(index + 1, deprovisionSteps.length - 1))
  }

  return (
    <InstallWizardPage
      title={`Deprovision ${installName}`}
      description="Teardown install services and remove the cloud stack without losing track of progress."
      steps={deprovisionSteps}
      currentStep={step.id}
      status={
        step.id === 'teardown' || stackDestroying ? 'deprovisioning' : undefined
      }
      onStepSelect={(stepId) =>
        setStepIndex(deprovisionSteps.findIndex((item) => item.id === stepId))
      }
      onBack={
        stepIndex > 0 ? () => setStepIndex((index) => index - 1) : undefined
      }
      onNext={handleNext}
      nextLabel={nextLabel}
      nextDisabled={nextDisabled}
      nextDisabledReason={nextDisabledReason}
      nextVariant={
        step.id === 'confirm' || step.id === 'stack' ? 'danger' : 'primary'
      }
    >
      {step.id === 'confirm' ? (
        <>
          <SectionHeader
            title="Deprovision this install"
            description="This removes the install services and cloud stack. This action cannot be undone."
          />
          <Banner theme="warn">
            <BannerCopy heading="Infrastructure will be removed">
              Components and the sandbox will be torn down before the install
              stack is destroyed.
            </BannerCopy>
          </Banner>
          <Input
            id="deprovision-confirmation"
            autoComplete="off"
            value={confirmation}
            onChange={(event) => setConfirmation(event.target.value)}
            placeholder={installName}
            labelProps={{
              labelText: `Type ${installName} to confirm`,
            }}
          />
        </>
      ) : null}

      {step.id === 'teardown' ? (
        <>
          <SectionHeader
            title="Teardown install services"
            description="Keep this page open while components and the sandbox are removed."
          />
          {error ? (
            <Banner theme="error">
              <BannerCopy heading="Install teardown failed">{error}</BannerCopy>
            </Banner>
          ) : null}
          <ProgressList items={progress} />
        </>
      ) : null}

      {step.id === 'stack' && initialStackOwnership === 'nuon' ? (
        <>
          <SectionHeader
            title="Destroy the install stack"
            description="Nuon will use the connected IAM role to remove the CloudFormation stack."
          />
          <Banner theme="info">
            <BannerCopy heading="Ready to destroy">
              Install services are gone. Destroying the stack removes the runner
              and remaining AWS resources.
            </BannerCopy>
          </Banner>
          <Status status={stackDestroying ? 'executing' : 'pending'}>
            {stackDestroying ? 'Destroying stack' : 'Waiting to destroy stack'}
          </Status>
        </>
      ) : null}

      {step.id === 'stack' && initialStackOwnership === 'customer' ? (
        <>
          <SectionHeader
            title="Destroy the install stack"
            description="Remove the CloudFormation stack from the customer AWS account."
          />
          <Banner theme="warn">
            <BannerCopy heading="Manual action required">
              Complete these steps in AWS, then return here to mark the stack
              destroyed.
            </BannerCopy>
          </Banner>
          <ol className="flex list-decimal flex-col gap-3 pl-5 text-sm">
            <li>Open the AWS CloudFormation console.</li>
            <li>
              Find the stack associated with{' '}
              <span className="font-mono whitespace-nowrap">{installName}</span>.
            </li>
            <li>Select the stack and choose Delete.</li>
            <li>Wait for deletion to finish.</li>
          </ol>
        </>
      ) : null}

      {step.id === 'complete' ? (
        <>
          <SectionHeader
            title="Install deprovisioned"
            description="Install services and cloud resources have been removed."
          />
          <Status status="deprovisioned" />
          <Text>
            {installName} remains in activity history with its final status.
          </Text>
        </>
      ) : null}
    </InstallWizardPage>
  )
}
