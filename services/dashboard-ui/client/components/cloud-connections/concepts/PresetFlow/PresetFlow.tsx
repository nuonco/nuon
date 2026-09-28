import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { cn } from '@/utils/classnames'
import {
  ACCOUNT_ID,
  PERMISSIONS_POLICY,
  REGION,
  ROLE_ARN,
  STACK_ACTIONS,
  STACK_RESOURCE,
  TRUST_POLICY,
  setupSnippet,
} from './mockData'

type TAccess = 'preset' | 'custom'
type TFormat = 'terraform' | 'cli' | 'cloudformation'
type TVerification =
  | 'idle'
  | 'verifying'
  | 'verified'
  | 'failed-trust'
  | 'failed-permissions'

interface IPresetFlow {
  initialStep?: number
  initialAccess?: TAccess
  initialFormat?: TFormat
  verification?: TVerification
  showTrustPolicy?: boolean
}

const FLOW_STEPS = [
  'Cloud and account',
  'Access preset',
  'Run in your cloud',
  'Verify',
]

const StepRail = ({
  step,
  setStep,
}: {
  step: number
  setStep: (step: number) => void
}) => (
  <nav
    aria-label="Cloud connection steps"
    className="flex gap-2 overflow-x-auto lg:flex-col"
  >
    {FLOW_STEPS.map((label, index) => {
      const number = index + 1
      const isCurrent = step === number
      const isComplete = number < step
      return (
        <Button
          key={label}
          variant="ghost"
          onClick={() => setStep(number)}
          aria-current={isCurrent ? 'step' : undefined}
          className={cn(
            'min-w-fit !justify-start !p-3 !shadow-none',
            isCurrent && '!bg-primary-50 dark:!bg-primary-950'
          )}
        >
          <span
            className={cn(
              'flex h-7 w-7 shrink-0 items-center justify-center rounded-full border text-sm',
              isCurrent && 'text-primary-600 dark:text-primary-400'
            )}
          >
            {isComplete ? (
              <Icon variant="CheckIcon" size={14} theme="success" />
            ) : (
              number
            )}
          </span>
          <Text
            as="span"
            variant="subtext"
            weight={isCurrent ? 'strong' : undefined}
          >
            {label}
          </Text>
        </Button>
      )
    })}
  </nav>
)

const CloudAndAccount = () => (
  <div className="flex max-w-2xl flex-col gap-6">
    <SectionHeader
      title="Choose the AWS account"
      description="This connection gives one Nuon identity access to one AWS account in this org."
    />
    <div className="grid gap-4 sm:grid-cols-2">
      <Input
        labelProps={{ labelText: 'AWS account ID' }}
        value={ACCOUNT_ID}
        readOnly
      />
      <Select
        labelProps={{ labelText: 'AWS region' }}
        value={REGION}
        onChange={() => {}}
        options={[
          { value: 'us-east-1', label: 'us-east-1' },
          { value: 'us-west-2', label: 'us-west-2' },
        ]}
      />
    </div>
    <Input
      labelProps={{ labelText: 'Connection name' }}
      value="Production stacks"
      onChange={() => {}}
      helperText="Use a name that identifies the account and its purpose."
    />
  </div>
)

const ActionList = () => (
  <Expand
    id="preset-flow-actions"
    heading={<Text weight="strong">View exact IAM actions and scope</Text>}
  >
    <div className="flex flex-col gap-4 border-t p-4">
      {STACK_ACTIONS.map((group) => (
        <div key={group.purpose} className="flex flex-col gap-2">
          <Text variant="subtext" weight="strong">
            {group.purpose}
          </Text>
          <Text
            family="mono"
            variant="subtext"
            theme="neutral"
            className="break-words"
          >
            {group.actions.join(' · ')}
          </Text>
        </div>
      ))}
      <div className="flex flex-col gap-1">
        <Text variant="subtext" weight="strong">
          Resource scope
        </Text>
        <Text
          family="mono"
          variant="subtext"
          theme="neutral"
          className="break-all"
        >
          {STACK_RESOURCE}
        </Text>
      </div>
    </div>
  </Expand>
)

const AccessPreset = ({
  access,
  setAccess,
  showTrustPolicy,
}: {
  access?: TAccess
  setAccess: (access: TAccess) => void
  showTrustPolicy: boolean
}) => (
  <div className="flex flex-col gap-6">
    <SectionHeader
      title="Choose what Nuon can access"
      description="Start with a Nuon preset or provide the permissions policy in AWS. The OIDC trust stays limited to this connection."
    />
    <div
      className={cn(
        'grid gap-6',
        access === 'custom'
          ? 'xl:grid-cols-[minmax(20rem,0.7fr)_minmax(0,1.3fr)]'
          : 'xl:grid-cols-[minmax(0,1.3fr)_minmax(20rem,0.7fr)]'
      )}
    >
      <Card
        className={cn(
          '!gap-4',
          access === 'preset' && 'ring-2 ring-primary-500'
        )}
      >
        <div className="flex items-start gap-4">
          <Icon variant="StackIcon" size={28} theme="brand" />
          <div className="flex min-w-0 flex-1 flex-col gap-2">
            <div className="flex flex-wrap items-center gap-2">
              <Text variant="base" weight="strong">
                Stacks
              </Text>
              <Badge theme="info">Recommended</Badge>
            </div>
            <Text>
              <strong>What you grant:</strong> Nuon can create, update, delete,
              and inspect the install's CloudFormation stack, including its
              outputs.
            </Text>
            <Text variant="subtext" theme="neutral">
              <strong>Why Nuon needs it:</strong> Nuon keeps the install
              infrastructure in sync without requiring an operator to apply
              every stack change by hand.
            </Text>
          </div>
        </div>
        <ActionList />
        <Banner theme="info">
          The rendered policy is a starting point. You can edit its actions or
          resource scope in AWS before attaching it.
        </Banner>
        <Button variant="secondary" onClick={() => setAccess('preset')}>
          {access === 'preset' ? 'Selected' : 'Select Stacks'}
        </Button>
      </Card>
      <Card
        className={cn(
          '!gap-4',
          access === 'custom' && 'ring-2 ring-primary-500'
        )}
      >
        <div className="flex items-start gap-3">
          <Icon variant="SlidersHorizontalIcon" size={24} theme="neutral" />
          <div className="flex flex-col gap-2">
            <Text variant="base" weight="strong">
              Customize access
            </Text>
            <Text>
              Nuon renders only the trust policy. Attach a permissions policy
              that matches the stack this connection will manage.
            </Text>
          </div>
        </div>
        <Text variant="subtext" theme="neutral">
          At minimum, the role needs permission to create or update the stack
          and read its status and outputs. Add delete access if Nuon should tear
          the stack down.
        </Text>
        <Button variant="secondary" onClick={() => setAccess('custom')}>
          {access === 'custom' ? 'Selected' : 'Select Custom'}
        </Button>
        {(access === 'custom' || showTrustPolicy) && (
          <div className="flex min-w-0 flex-col gap-2 border-t pt-4">
            <Text weight="strong">Trust policy</Text>
            <Text variant="subtext" theme="neutral">
              Keep the audience and subject unchanged so only this Nuon
              connection can assume the role.
            </Text>
            <CodeBlock language="json" showCopy>
              {TRUST_POLICY}
            </CodeBlock>
          </div>
        )}
      </Card>
    </div>
  </div>
)

const RunbookStep = ({
  number,
  title,
  description,
  status,
  children,
}: {
  number: number
  title: string
  description: string
  status: 'ready' | 'done' | 'failed'
  children: React.ReactNode
}) => (
  <Card className="!gap-0 !p-0 overflow-hidden">
    <div className="flex items-center gap-4 p-5">
      <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border">
        {status === 'done' ? (
          <Icon variant="CheckIcon" theme="success" />
        ) : status === 'failed' ? (
          <Icon variant="WarningIcon" theme="error" />
        ) : (
          number
        )}
      </span>
      <div className="flex min-w-0 flex-1 flex-col gap-1">
        <Text weight="strong">{title}</Text>
        <Text variant="subtext" theme="neutral">
          {description}
        </Text>
      </div>
      <Badge
        theme={
          status === 'failed'
            ? 'error'
            : status === 'done'
              ? 'success'
              : 'neutral'
        }
      >
        {status === 'failed'
          ? 'Needs attention'
          : status === 'done'
            ? 'Done'
            : 'Ready'}
      </Badge>
    </div>
    <div className="flex min-w-0 flex-col gap-4 border-t p-5">{children}</div>
  </Card>
)

const RunInCloud = ({
  access,
  format,
  setFormat,
  failure,
}: {
  access: TAccess
  format: TFormat
  setFormat: (format: TFormat) => void
  failure?: 'trust' | 'permissions'
}) => {
  const language =
    format === 'terraform'
      ? 'hcl'
      : format === 'cloudformation'
        ? 'yaml'
        : 'bash'
  const setupSteps = [
    {
      title: 'Create the OIDC provider',
      description: 'Add the Nuon issuer to this AWS account once.',
      part: 'oidc' as const,
    },
    {
      title: 'Create the role with the trust policy',
      description:
        'Use the exact audience and connection subject rendered by Nuon.',
      part: 'role' as const,
    },
    ...(access === 'preset'
      ? [
          {
            title: 'Attach the Stacks permissions policy',
            description:
              'Use the rendered policy as-is or edit its access in AWS before attaching it.',
            part: 'permissions' as const,
          },
        ]
      : []),
  ]
  return (
    <div className="flex flex-col gap-6">
      <SectionHeader
        title="Run the setup in AWS"
        description="Complete the steps in order, then paste the role ARN. Nuon never writes IAM resources into the account."
        actions={
          <ToggleButton<TFormat>
            label="Setup format"
            value={format}
            onChange={setFormat}
            options={[
              { value: 'terraform', label: 'Terraform' },
              { value: 'cli', label: 'AWS CLI' },
              { value: 'cloudformation', label: 'CloudFormation' },
            ]}
          />
        }
      />
      {access === 'custom' && (
        <Banner theme="info">
          Custom access omits the permissions step. Attach your permissions
          policy to the role in AWS.
        </Banner>
      )}
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-4">
        {setupSteps.map((setupStep, index) => {
          const number = index + 1
          const isFailed =
            failure === 'trust'
              ? number === 2
              : failure === 'permissions' && setupStep.part === 'permissions'
          return (
            <RunbookStep
              key={setupStep.part}
              number={number}
              title={setupStep.title}
              description={setupStep.description}
              status={isFailed ? 'failed' : 'ready'}
            >
              <Text variant="subtext" weight="strong">
                {format === 'terraform'
                  ? 'Terraform resource'
                  : format === 'cloudformation'
                    ? 'CloudFormation resource'
                    : 'AWS CLI command'}
              </Text>
              <CodeBlock language={language} showCopy>
                {setupSnippet(format, setupStep.part)}
              </CodeBlock>
              {setupStep.part === 'role' && (
                <div className="flex min-w-0 flex-col gap-2">
                  <Text variant="subtext" weight="strong">
                    Trust policy · nuon-trust.json
                  </Text>
                  <CodeBlock language="json" showCopy>
                    {TRUST_POLICY}
                  </CodeBlock>
                </div>
              )}
              {setupStep.part === 'permissions' && (
                <div className="flex min-w-0 flex-col gap-2">
                  <Text variant="subtext" weight="strong">
                    Permissions policy · nuon-permissions.json
                  </Text>
                  <CodeBlock language="json" showCopy>
                    {PERMISSIONS_POLICY}
                  </CodeBlock>
                </div>
              )}
            </RunbookStep>
          )
        })}
        <RunbookStep
          number={setupSteps.length + 1}
          title="Paste the role ARN"
          description="Copy the role ARN from the completed AWS setup."
          status="ready"
        >
          <Input
            labelProps={{ labelText: 'IAM role ARN' }}
            value={ROLE_ARN}
            onChange={() => {}}
          />
        </RunbookStep>
      </div>
    </div>
  )
}

const VerificationCheck = ({
  status,
  children,
}: {
  status: 'pending' | 'checking' | 'passed' | 'failed'
  children: React.ReactNode
}) => (
  <div className="flex items-center gap-3 py-3 border-b last:border-b-0">
    {status === 'checking' ? (
      <Loading size={18} />
    ) : (
      <Icon
        variant={
          status === 'failed'
            ? 'WarningCircleIcon'
            : status === 'passed'
              ? 'CheckCircleIcon'
              : 'ClockIcon'
        }
        theme={
          status === 'failed'
            ? 'error'
            : status === 'passed'
              ? 'success'
              : 'neutral'
        }
        weight={status === 'passed' ? 'fill' : 'regular'}
      />
    )}
    <Text>{children}</Text>
  </div>
)

const Verify = ({
  access,
  verification,
  goToStep,
}: {
  access: TAccess
  verification: TVerification
  goToStep: (step: number) => void
}) => {
  const isWorking = verification === 'verifying'
  const isVerified = verification === 'verified'
  const trustFailed = verification === 'failed-trust'
  const permissionsFailed = verification === 'failed-permissions'
  return (
    <div className="flex max-w-3xl flex-col gap-6">
      <SectionHeader
        title="Verify the connection"
        description={
          access === 'custom'
            ? 'Nuon checks that this connection can assume the role and that a different Nuon connection cannot.'
            : 'Nuon checks the OIDC identity, rejects a foreign Nuon connection, and runs a read-only CloudFormation probe.'
        }
      />
      {isWorking && (
        <Banner theme="info">
          <div className="flex flex-col gap-1">
            <Text weight="strong">Verifying connection</Text>
            <Text variant="subtext">
              Testing the OIDC exchange and account access.
            </Text>
          </div>
        </Banner>
      )}
      {isVerified && (
        <Banner theme="success">
          <div className="flex flex-col gap-1">
            <Text weight="strong">Connection verified</Text>
            <Text variant="subtext">
              The role is limited to this connection and the selected access is
              available.
            </Text>
          </div>
        </Banner>
      )}
      {trustFailed && (
        <Banner theme="error">
          <div className="flex flex-col items-start gap-2">
            <div className="flex flex-col gap-1">
              <Text weight="strong">OIDC identity check failed</Text>
              <Text variant="subtext">
                AWS rejected this connection's subject. Fix the trust policy in
                step 2 of the AWS runbook.
              </Text>
            </div>
            <Button variant="secondary" onClick={() => goToStep(3)}>
              Fix runbook step 2
            </Button>
          </div>
        </Banner>
      )}
      {permissionsFailed && (
        <Banner theme="error">
          <div className="flex flex-col items-start gap-2">
            <div className="flex flex-col gap-1">
              <Text weight="strong">CloudFormation probe failed</Text>
              <Text variant="subtext">
                AWS denied cloudformation:DescribeStacks. Fix the permissions
                policy in step 3 of the AWS runbook.
              </Text>
            </div>
            <Button variant="secondary" onClick={() => goToStep(3)}>
              Fix runbook step 3
            </Button>
          </div>
        </Banner>
      )}
      <Card className="!gap-0">
        <Text weight="strong">Verification checks</Text>
        <VerificationCheck
          status={
            trustFailed
              ? 'failed'
              : isWorking
                ? 'checking'
                : isVerified || permissionsFailed
                  ? 'passed'
                  : 'pending'
          }
        >
          Exchange an OIDC token for this connection
        </VerificationCheck>
        <VerificationCheck
          status={isVerified || permissionsFailed ? 'passed' : 'pending'}
        >
          Reject a foreign Nuon connection
        </VerificationCheck>
        {access === 'preset' && (
          <VerificationCheck
            status={
              permissionsFailed ? 'failed' : isVerified ? 'passed' : 'pending'
            }
          >
            Read the install stack with cloudformation:DescribeStacks
          </VerificationCheck>
        )}
      </Card>
      {!isVerified && !isWorking && (
        <Button variant="primary">Verify connection</Button>
      )}
      {isVerified && <Button variant="primary">Create connection</Button>}
    </div>
  )
}

export const PresetFlow = ({
  initialStep = 1,
  initialAccess,
  initialFormat = 'terraform',
  verification = 'idle',
  showTrustPolicy = false,
}: IPresetFlow) => {
  const [step, setStep] = useState(initialStep)
  const [access, setAccess] = useState<TAccess | undefined>(initialAccess)
  const [format, setFormat] = useState<TFormat>(initialFormat)
  const selectedAccess = access ?? 'preset'

  return (
    <PageSection className="min-h-screen bg-white dark:bg-dark-grey-900">
      <SectionHeader
        variant="page"
        title="Create cloud connection"
        description="Grant one Nuon identity controlled access to one cloud account."
        status={<Badge theme="info">AWS · OIDC</Badge>}
      />
      <div className="grid flex-1 gap-8 lg:grid-cols-[15rem_minmax(0,1fr)]">
        <StepRail step={step} setStep={setStep} />
        <main className="min-w-0">
          {step === 1 && <CloudAndAccount />}
          {step === 2 && (
            <AccessPreset
              access={access}
              setAccess={setAccess}
              showTrustPolicy={showTrustPolicy}
            />
          )}
          {step === 3 && (
            <RunInCloud
              access={selectedAccess}
              format={format}
              setFormat={setFormat}
              failure={
                verification === 'failed-trust'
                  ? 'trust'
                  : verification === 'failed-permissions'
                    ? 'permissions'
                    : undefined
              }
            />
          )}
          {step === 4 && (
            <Verify
              access={selectedAccess}
              verification={verification}
              goToStep={setStep}
            />
          )}
        </main>
      </div>
      {step < 4 && (
        <div className="flex justify-end gap-2 border-t pt-4">
          {step > 1 && (
            <Button
              variant="secondary"
              onClick={() => setStep((current) => current - 1)}
            >
              Back
            </Button>
          )}
          <Button
            variant="primary"
            disabled={step === 2 && !access}
            tooltipProps={
              step === 2 && !access
                ? {
                    tipContent:
                      'Cannot continue — choose Stacks or Custom access',
                  }
                : undefined
            }
            onClick={() => setStep((current) => Math.min(4, current + 1))}
          >
            {step === 3 ? 'Review verification' : 'Continue'}
          </Button>
        </div>
      )}
    </PageSection>
  )
}
