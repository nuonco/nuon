import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Loading } from '@/components/common/Loading'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { ToggleButton } from '@/components/common/ToggleButton'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { Textarea } from '@/components/common/form/Textarea'
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
  initialTrustPolicy?: string
  startTrustEditing?: boolean
}

const FLOW_STEPS = [
  'Cloud and account',
  'Access',
  'Run in your cloud',
  'Verify',
]

const FieldLabel = ({ label, tip }: { label: string; tip: string }) => (
  <span className="flex items-center gap-1.5">
    {label}
    <Tooltip
      position="top"
      tipContent={tip}
      tipContentClassName="max-w-72 whitespace-normal"
    >
      <span
        aria-label={`${label} information`}
        className="inline-flex text-cool-grey-500 dark:text-cool-grey-400"
        tabIndex={0}
      >
        <Icon variant="InfoIcon" size={14} />
      </span>
    </Tooltip>
  </span>
)

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
      description="Nuon assumes an IAM role in this account via OIDC. You decide what that role can do."
    />
    <div className="grid gap-4 sm:grid-cols-2">
      <Input
        id="preset-flow-account-id"
        labelProps={{
          labelText: (
            <FieldLabel
              label="AWS account ID"
              tip="The 12-digit ID of the AWS account you want to give Nuon access to."
            />
          ),
        }}
        value={ACCOUNT_ID}
        readOnly
      />
      <Select
        id="preset-flow-region"
        labelProps={{
          labelText: (
            <FieldLabel
              label="AWS region"
              tip="Where Nuon will create install stacks by default; you can override per install."
            />
          ),
        }}
        value={REGION}
        onChange={() => {}}
        options={[
          { value: 'us-east-1', label: 'us-east-1' },
          { value: 'us-west-2', label: 'us-west-2' },
        ]}
      />
    </div>
    <Input
      id="preset-flow-name"
      labelProps={{
        labelText: (
          <FieldLabel
            label="Connection name"
            tip="Shown wherever this connection is picked, e.g. when creating an install."
          />
        ),
      }}
      value="Production stacks"
      onChange={() => {}}
      helperText="Use a name that identifies the account and its purpose."
    />
  </div>
)

const hasRequiredTrustConditions = (policy: string) => {
  try {
    const parsed = JSON.parse(policy)
    const statements = Array.isArray(parsed.Statement)
      ? parsed.Statement
      : [parsed.Statement]
    return statements.some((statement: any) => {
      const conditions = statement?.Condition?.StringEquals
      return (
        conditions?.['api.nuon.co:aud'] === 'sts.amazonaws.com' &&
        conditions?.['api.nuon.co:sub'] ===
          'org:org_01JEXAMPLE:connection:cc_01JEXAMPLE'
      )
    })
  } catch {
    return false
  }
}

const TrustPolicyEditor = ({
  trustPolicy,
  setTrustPolicy,
  initiallyEditing = false,
}: {
  trustPolicy: string
  setTrustPolicy: (policy: string) => void
  initiallyEditing?: boolean
}) => {
  const [isEditing, setIsEditing] = useState(initiallyEditing)
  const isValidForConnection = hasRequiredTrustConditions(trustPolicy)
  return (
    <div className="flex min-w-0 flex-col gap-3">
      <div className="flex items-center justify-between gap-3">
        <Text weight="strong">Trust policy</Text>
        <Button
          size="sm"
          variant="secondary"
          onClick={() => setIsEditing((current) => !current)}
        >
          {isEditing ? 'Finish editing' : 'Edit'}
        </Button>
      </div>
      <Text variant="subtext" theme="neutral">
        Start with the rendered policy. The audience and subject conditions must
        stay unchanged for verification to pass.
      </Text>
      {isEditing ? (
        <>
          <Textarea
            id="preset-flow-trust-policy"
            labelProps={{ labelText: 'Editable trust policy' }}
            value={trustPolicy}
            onChange={(event) => setTrustPolicy(event.currentTarget.value)}
            rows={16}
            className="font-mono text-xs"
          />
          {!isValidForConnection && (
            <Banner theme="warn">
              Restore the rendered audience and subject conditions before
              verifying this connection.
            </Banner>
          )}
          <Button
            className="w-fit"
            variant="secondary"
            onClick={() => setTrustPolicy(TRUST_POLICY)}
          >
            Reset to rendered
          </Button>
        </>
      ) : (
        <CodeBlock language="json" showCopy>
          {trustPolicy}
        </CodeBlock>
      )}
    </div>
  )
}

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
  trustPolicy,
  setTrustPolicy,
  startTrustEditing,
}: {
  access?: TAccess
  setAccess: (access: TAccess) => void
  showTrustPolicy: boolean
  trustPolicy: string
  setTrustPolicy: (policy: string) => void
  startTrustEditing: boolean
}) => (
  <div className="flex flex-col gap-6">
    <SectionHeader
      title="Choose what Nuon can access"
      description="Choose managed stack access or provide the permissions policy in AWS. The OIDC trust stays limited to this connection."
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
                Manage install stacks
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
              Nuon renders the trust policy so this connection can assume the
              role. You attach whatever permissions policy you want.
            </Text>
          </div>
        </div>
        <Text variant="subtext" theme="neutral">
          The Manage install stacks policy is a good starting point if you want
          to trim it.
        </Text>
        <Button variant="secondary" onClick={() => setAccess('custom')}>
          {access === 'custom' ? 'Selected' : 'Select Custom'}
        </Button>
        {(access === 'custom' || showTrustPolicy) && (
          <div className="border-t pt-4">
            <TrustPolicyEditor
              trustPolicy={trustPolicy}
              setTrustPolicy={setTrustPolicy}
              initiallyEditing={startTrustEditing}
            />
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
  action,
  children,
}: {
  number: number
  title: string
  description: string
  status: 'ready' | 'done' | 'failed'
  action?: React.ReactNode
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
      {action}
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
  trustPolicy,
  setTrustPolicy,
  roleArn,
  setRoleArn,
  awsTrustURL,
  startTrustEditing,
}: {
  access: TAccess
  format: TFormat
  setFormat: (format: TFormat) => void
  failure?: 'trust' | 'permissions'
  trustPolicy: string
  setTrustPolicy: (policy: string) => void
  roleArn: string
  setRoleArn: (arn: string) => void
  awsTrustURL: string
  startTrustEditing: boolean
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
              { value: 'cli', label: 'AWS CLI' },
              { value: 'terraform', label: 'Terraform' },
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
      {format === 'cloudformation' && (
        <Banner theme="info">
          <div className="flex flex-col gap-1">
            <Text weight="strong">Quick create needs a hosted template</Text>
            <Text variant="subtext">
              A one-click AWS quick-create link can replace these copy steps
              once Nuon hosts this CloudFormation template at a stable URL.
            </Text>
          </div>
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
              action={
                setupStep.part === 'role' ? (
                  <Link href={awsTrustURL} isExternal>
                    Open in AWS console
                  </Link>
                ) : undefined
              }
            >
              <Text variant="subtext" weight="strong">
                {format === 'terraform'
                  ? 'Terraform resource'
                  : format === 'cloudformation'
                    ? 'CloudFormation resource'
                    : 'AWS CLI command'}
              </Text>
              <CodeBlock language={language} showCopy>
                {setupSnippet(format, setupStep.part, trustPolicy)}
              </CodeBlock>
              {setupStep.part === 'role' && (
                <TrustPolicyEditor
                  trustPolicy={trustPolicy}
                  setTrustPolicy={setTrustPolicy}
                  initiallyEditing={startTrustEditing}
                />
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
            id="preset-flow-role-arn"
            labelProps={{
              labelText: (
                <FieldLabel
                  label="IAM role ARN"
                  tip="The ARN of the role you created in step 2."
                />
              ),
            }}
            value={roleArn}
            onChange={(event) => setRoleArn(event.currentTarget.value)}
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
  awsTrustURL,
}: {
  access: TAccess
  verification: TVerification
  goToStep: (step: number) => void
  awsTrustURL: string
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
            ? 'Nuon checks that the role trusts this connection and rejects other Nuon connections.'
            : 'Nuon checks the role trust, rejects other Nuon connections, and runs a read-only CloudFormation probe.'
        }
      />
      {isWorking && (
        <Banner theme="info" className="!text-blue-800 dark:!text-blue-300">
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
              <Text weight="strong">Role trust check failed</Text>
              <Text variant="subtext">
                AWS rejected this connection's subject. Fix the trust policy in
                step 2 of the AWS runbook.
              </Text>
            </div>
            <Button variant="secondary" onClick={() => goToStep(3)}>
              Fix runbook step 2
            </Button>
            <Link href={awsTrustURL} isExternal>
              Open in AWS console
            </Link>
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
          Reject another Nuon connection
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
  initialFormat = 'cli',
  verification = 'idle',
  showTrustPolicy = false,
  initialTrustPolicy = TRUST_POLICY,
  startTrustEditing = false,
}: IPresetFlow) => {
  const [step, setStep] = useState(initialStep)
  const [access, setAccess] = useState<TAccess | undefined>(initialAccess)
  const [format, setFormat] = useState<TFormat>(initialFormat)
  const [trustPolicy, setTrustPolicy] = useState(initialTrustPolicy)
  const [roleArn, setRoleArn] = useState(ROLE_ARN)
  const selectedAccess = access ?? 'preset'
  const roleName = roleArn.match(/role\/(.+)$/)?.[1] || 'nuon-cloud-connection'
  const awsTrustURL = `https://console.aws.amazon.com/iam/home#/roles/details/${encodeURIComponent(roleName)}?section=trust`

  return (
    <PageSection className="min-h-screen bg-white dark:bg-dark-grey-900">
      <SectionHeader
        variant="page"
        title="Create cloud connection"
        description="Let Nuon manage resources in your AWS account through a role you control."
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
              trustPolicy={trustPolicy}
              setTrustPolicy={setTrustPolicy}
              startTrustEditing={startTrustEditing}
            />
          )}
          {step === 3 && (
            <RunInCloud
              access={selectedAccess}
              format={format}
              setFormat={setFormat}
              trustPolicy={trustPolicy}
              setTrustPolicy={setTrustPolicy}
              roleArn={roleArn}
              setRoleArn={setRoleArn}
              awsTrustURL={awsTrustURL}
              startTrustEditing={startTrustEditing}
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
              awsTrustURL={awsTrustURL}
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
