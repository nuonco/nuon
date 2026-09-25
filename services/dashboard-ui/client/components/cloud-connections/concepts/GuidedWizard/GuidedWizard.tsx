import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { Textarea } from '@/components/common/form/Textarea'
import { Toggle } from '@/components/common/form/Toggle'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { cn } from '@/utils/classnames'
import {
  ACCOUNT_ID,
  IMAGE_ACTIONS,
  STACK_ACTIONS,
  SUBJECT,
  TRUST_POLICY,
  cliSnippet,
  cloudFormationSnippet,
  permissionsPolicy,
  terraformSnippet,
} from './mockData'

type TVerifyState = 'idle' | 'verifying' | 'verified' | 'failed'

interface IGuidedWizard {
  initialStep?: number
  verifyState?: TVerifyState
}

const STEPS = [
  'Cloud & account',
  'What Nuon will do',
  'Review & customize',
  'Run it in AWS',
  'Verify',
]

const Permissions = ({
  id,
  actions,
}: {
  id: string
  actions: readonly (readonly [string, string])[]
}) => (
  <Expand
    id={id}
    heading={<Text weight="strong">Permissions this grants</Text>}
  >
    <div className="flex flex-col gap-3 border-t p-4">
      {actions.map(([plain, raw]) => (
        <div key={raw} className="flex flex-col gap-1">
          <Text variant="subtext">{plain}</Text>
          <Text
            family="mono"
            variant="subtext"
            theme="neutral"
            className="break-words"
          >
            {raw}
          </Text>
        </div>
      ))}
    </div>
  </Expand>
)

export const GuidedWizard = ({
  initialStep = 1,
  verifyState = 'idle',
}: IGuidedWizard) => {
  const [step, setStep] = useState(initialStep)
  const [images, setImages] = useState(true)
  const [repos, setRepos] = useState('*')
  const [region, setRegion] = useState('us-east-1')
  const [stackPrefix, setStackPrefix] = useState('nuon-*')
  const [custom, setCustom] = useState(false)

  const page = (() => {
    if (step === 1)
      return (
        <div className="flex flex-col gap-6 max-w-2xl">
          <SectionHeader
            title="Choose the cloud account"
            description="This connection gives one Nuon identity access to one cloud account."
          />
          <Select
            labelProps={{ labelText: 'Cloud' }}
            value="aws"
            onChange={() => {}}
            options={[
              { value: 'aws', label: 'Amazon Web Services' },
              { value: 'azure', label: 'Microsoft Azure' },
              { value: 'gcp', label: 'Google Cloud' },
            ]}
          />
          <Input
            labelProps={{ labelText: 'AWS account ID' }}
            value={ACCOUNT_ID}
            readOnly
            helperText="Nuon will create and manage install infrastructure in this account."
          />
          <Input
            labelProps={{ labelText: 'Connection name' }}
            value="Production AWS"
            readOnly
          />
        </div>
      )
    if (step === 2)
      return (
        <div className="flex flex-col gap-6">
          <SectionHeader
            title="Choose what Nuon can do"
            description="The default covers install infrastructure and private image builds. Turn off anything this connection does not need."
          />
          <div className="grid gap-4 xl:grid-cols-2">
            <Card className="!p-0 overflow-hidden">
              <div className="flex gap-4 p-6">
                <Icon variant="StackIcon" size={24} theme="brand" />
                <div className="flex flex-col gap-2">
                  <Text variant="base" weight="strong">
                    Install stacks
                  </Text>
                  <Text>
                    Nuon creates, updates, and deletes each install bootstrap
                    stack.
                  </Text>
                  <Text variant="subtext" theme="neutral">
                    <strong>Why:</strong> Customers should not have to create
                    runner roles, networking, or other install resources by
                    hand.
                  </Text>
                </div>
                <Toggle checked onChange={() => {}} />
              </div>
              <Permissions
                id="wizard-stack-permissions"
                actions={STACK_ACTIONS}
              />
            </Card>
            <Card className="!p-0 overflow-hidden">
              <div className="flex gap-4 p-6">
                <Icon variant="ShippingContainerIcon" size={24} theme="brand" />
                <div className="flex flex-col gap-2">
                  <Text variant="base" weight="strong">
                    Pull images
                  </Text>
                  <Text>
                    Nuon's build runner reads private images from ECR.
                  </Text>
                  <Text variant="subtext" theme="neutral">
                    <strong>Why:</strong> Nuon needs the source image to build
                    and publish it into each install registry.
                  </Text>
                </div>
                <Toggle checked={images} onChange={setImages} />
              </div>
              <Permissions
                id="wizard-image-permissions"
                actions={IMAGE_ACTIONS}
              />
            </Card>
          </div>
        </div>
      )
    if (step === 3)
      return (
        <div className="flex flex-col gap-6">
          <SectionHeader
            title="Review and customize access"
            description="The default policy is ready to use. Narrow its scope here or supply a permissions policy."
            status={<Badge theme="info">Default is recommended</Badge>}
          />
          <div className="grid gap-6 xl:grid-cols-[minmax(20rem,0.8fr)_minmax(34rem,1.2fr)]">
            <Card className="!gap-5">
              <Text weight="strong">Customize the default</Text>
              <Input
                labelProps={{ labelText: 'ECR repositories' }}
                value={repos}
                onChange={(event) => setRepos(event.currentTarget.value)}
                helperText="Use * for all repositories or enter a repository name."
              />
              <Select
                labelProps={{ labelText: 'AWS region' }}
                value={region}
                onChange={setRegion}
                options={[
                  { value: 'us-east-1', label: 'us-east-1' },
                  { value: 'us-west-2', label: 'us-west-2' },
                ]}
              />
              <Input
                labelProps={{ labelText: 'Install stack name prefix' }}
                value={stackPrefix}
                onChange={(event) => setStackPrefix(event.currentTarget.value)}
                helperText="Only stacks with this prefix are in scope."
              />
              <Toggle
                checked={custom}
                onChange={setCustom}
                label="Write my own policy"
                description="Replace the default permissions policy. The trust policy stays fixed."
              />
            </Card>
            <div className="flex min-w-0 flex-col gap-5">
              <div className="flex flex-col gap-2">
                <Text weight="strong">Permissions policy</Text>
                {custom ? (
                  <Textarea
                    labelProps={{ labelText: 'Custom permissions policy' }}
                    value={permissionsPolicy(
                      images,
                      stackPrefix,
                      repos,
                      region
                    )}
                    onChange={() => {}}
                    minRows={18}
                    className="font-mono text-xs"
                  />
                ) : (
                  <CodeBlock language="json" showCopy>
                    {permissionsPolicy(images, stackPrefix, repos, region)}
                  </CodeBlock>
                )}
              </div>
              <div className="flex flex-col gap-2">
                <Text weight="strong">Trust policy</Text>
                <Text variant="subtext" theme="neutral">
                  Read-only. The exact OIDC subject below means only this Nuon
                  org and connection can assume the role.
                </Text>
                <CodeBlock language="json" showCopy>
                  {TRUST_POLICY}
                </CodeBlock>
              </div>
            </div>
          </div>
        </div>
      )
    if (step === 4)
      return (
        <div className="flex flex-col gap-6">
          <SectionHeader
            title="Run the setup in AWS"
            description="Choose one format. It creates the OIDC provider, role, trust policy, and permissions policy."
          />
          <Tabs
            tabLabels={{ cli: 'AWS CLI', cloudformation: 'CloudFormation' }}
            tabs={{
              terraform: (
                <CodeBlock language="hcl" showCopy>
                  {terraformSnippet(repos, region, stackPrefix)}
                </CodeBlock>
              ),
              cli: (
                <CodeBlock language="bash" showCopy>
                  {cliSnippet(region)}
                </CodeBlock>
              ),
              cloudformation: (
                <div className="flex flex-col gap-3">
                  <Button variant="secondary">Open quick-create in AWS</Button>
                  <CodeBlock language="yaml" showCopy>
                    {cloudFormationSnippet(stackPrefix)}
                  </CodeBlock>
                </div>
              ),
            }}
          />
          <Card>
            <Input
              labelProps={{ labelText: 'IAM role ARN' }}
              placeholder={`arn:aws:iam::${ACCOUNT_ID}:role/nuon-cloud-connection`}
              helperText="Paste the role ARN after the setup finishes."
            />
          </Card>
        </div>
      )
    return (
      <div className="flex flex-col gap-6 max-w-3xl">
        <SectionHeader
          title="Verify the connection"
          description="Nuon proves the role works, rejects a foreign subject, and checks each selected capability."
        />
        {verifyState === 'verifying' && (
          <Banner theme="info">
            <div className="flex flex-col gap-1">
              <Text weight="strong">Verifying connection</Text>
              <Text variant="subtext">
                Testing the OIDC exchange and capability permissions.
              </Text>
            </div>
          </Banner>
        )}
        {verifyState === 'verified' && (
          <Banner theme="success">
            <div className="flex flex-col gap-2">
              <Text weight="strong">Connection verified</Text>
              <Text variant="subtext">
                The exact subject can assume the role. A foreign subject was
                denied.
              </Text>
              <div className="flex gap-2">
                <Badge theme="success">Install stacks</Badge>
                <Badge theme="success">Pull images</Badge>
              </div>
            </div>
          </Banner>
        )}
        {verifyState === 'failed' && (
          <Banner theme="error">
            <div className="flex flex-col gap-1">
              <Text weight="strong">Capability probe failed</Text>
              <Text variant="subtext">
                ECR denied ecr:BatchGetImage for repository acme/payments. Add
                the repository to the policy, then verify again.
              </Text>
            </div>
          </Banner>
        )}
        <Card className="!gap-4">
          <Text weight="strong">Checks</Text>
          {[
            'Positive OIDC exchange',
            'Foreign-subject denial',
            'Install stack permissions',
            'ECR image permissions',
          ].map((check, index) => (
            <div key={check} className="flex items-center gap-2">
              <Icon
                variant={
                  verifyState === 'failed' && index === 3
                    ? 'WarningOctagonIcon'
                    : verifyState === 'verifying'
                      ? 'ClockIcon'
                      : 'CheckCircleIcon'
                }
                theme={
                  verifyState === 'failed' && index === 3
                    ? 'error'
                    : verifyState === 'verifying'
                      ? 'neutral'
                      : 'success'
                }
              />
              <Text>{check}</Text>
            </div>
          ))}
        </Card>
      </div>
    )
  })()

  return (
    <PageSection className="min-h-screen bg-white dark:bg-dark-grey-900">
      <SectionHeader
        variant="page"
        title="Create cloud connection"
        description={`Grant Nuon controlled access to AWS account ${ACCOUNT_ID}.`}
      />
      <div className="grid flex-1 gap-8 lg:grid-cols-[15rem_minmax(0,1fr)]">
        <nav
          aria-label="Connection steps"
          className="flex flex-row gap-2 overflow-x-auto lg:flex-col"
        >
          {STEPS.map((label, index) => {
            const number = index + 1
            return (
              <button
                type="button"
                key={label}
                onClick={() => setStep(number)}
                className={cn(
                  'flex min-w-fit items-center gap-3 rounded-md p-3 text-left cursor-pointer',
                  step === number
                    ? 'bg-primary-50 dark:bg-primary-950'
                    : 'hover:bg-cool-grey-500/8'
                )}
              >
                <span
                  className={cn(
                    'flex h-7 w-7 items-center justify-center rounded-full border text-sm',
                    step === number &&
                      'border-primary-500 text-primary-600 dark:text-primary-400'
                  )}
                >
                  {number}
                </span>
                <Text
                  variant="subtext"
                  weight={step === number ? 'strong' : 'default'}
                >
                  {label}
                </Text>
              </button>
            )
          })}
        </nav>
        <main className="min-w-0">{page}</main>
      </div>
      <div className="flex justify-end gap-2 border-t pt-4">
        <Button
          variant="secondary"
          disabled={step === 1}
          tooltipProps={
            step === 1
              ? { tipContent: 'Cannot go back — this is the first step' }
              : undefined
          }
          onClick={() => setStep((value) => Math.max(1, value - 1))}
        >
          Back
        </Button>
        <Button
          variant="primary"
          onClick={() => setStep((value) => Math.min(5, value + 1))}
        >
          {step === 4
            ? 'Review verification'
            : step === 5
              ? 'Verify connection'
              : 'Continue'}
        </Button>
      </div>
      <Text variant="subtext" theme="neutral" className="sr-only">
        OIDC subject {SUBJECT}
      </Text>
    </PageSection>
  )
}
