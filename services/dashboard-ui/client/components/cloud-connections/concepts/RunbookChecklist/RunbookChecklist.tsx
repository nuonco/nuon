import { useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { ToggleButton } from '@/components/common/ToggleButton'
import { Input } from '@/components/common/form/Input'
import { Textarea } from '@/components/common/form/Textarea'
import { PageSection } from '@/components/layout/PageSection'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { cn } from '@/utils/classnames'
import { ACCOUNT_ID, POLICY, SUBJECT, snippets } from './mockData'

type TFormat = keyof typeof snippets
interface IRunbookChecklist {
  completedThrough?: number
  failedStep?: number
  allDone?: boolean
}
const STEP_NAMES = [
  'Create the trust',
  'Create the role',
  'Attach permissions',
  'Tell Nuon the role ARN',
  'Verify',
]

export const RunbookChecklist = ({
  completedThrough = 0,
  failedStep,
  allDone = false,
}: IRunbookChecklist) => {
  const [format, setFormat] = useState<TFormat>('cli')
  const [scopeOpen, setScopeOpen] = useState(false)
  const [policyOpen, setPolicyOpen] = useState(false)
  const [done, setDone] = useState(
    () =>
      new Set(Array.from({ length: completedThrough }, (_, index) => index + 1))
  )
  const toggleDone = (step: number) =>
    setDone((current) => {
      const next = new Set(current)
      next.has(step) ? next.delete(step) : next.add(step)
      return next
    })

  return (
    <PageSection className="min-h-screen bg-white dark:bg-dark-grey-900">
      <SectionHeader
        variant="page"
        title="Connect AWS account"
        description="Follow the runbook in order. Nuon verifies the result before using the role."
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
      <Card className="!gap-4">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="flex max-w-4xl flex-col gap-2">
            <Text variant="base" weight="strong">
              You are granting Nuon access to AWS account {ACCOUNT_ID}
            </Text>
            <Text variant="subtext" theme="neutral">
              Only subject{' '}
              <Text as="span" family="mono" variant="subtext">
                {SUBJECT}
              </Text>{' '}
              can assume this role.
            </Text>
            <div className="flex flex-wrap gap-2">
              <Badge theme="info">Install stacks</Badge>
              <Badge theme="info">Pull images · 2 repositories</Badge>
              <Badge theme="neutral">us-east-1</Badge>
            </div>
          </div>
          <Button
            variant="secondary"
            onClick={() => setScopeOpen((value) => !value)}
          >
            Change what is granted
          </Button>
        </div>
        {scopeOpen && (
          <div className="grid gap-4 border-t pt-4 sm:grid-cols-3">
            <Input
              labelProps={{ labelText: 'ECR repositories' }}
              value="acme/api, acme/worker"
              onChange={() => {}}
            />
            <Input
              labelProps={{ labelText: 'AWS region' }}
              value="us-east-1"
              onChange={() => {}}
            />
            <Input
              labelProps={{ labelText: 'Install stack name prefix' }}
              value="nuon-"
              onChange={() => {}}
            />
          </div>
        )}
      </Card>
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-4">
        {STEP_NAMES.map((name, index) => {
          const number = index + 1
          const isDone = allDone || done.has(number)
          const isFailed = failedStep === number
          return (
            <Card
              key={name}
              className={cn(
                '!p-0 overflow-hidden',
                isFailed && '!border-red-500'
              )}
            >
              <div className="flex items-center gap-4 p-5">
                <button
                  type="button"
                  aria-label={`${isDone ? 'Mark incomplete' : 'Mark complete'}: ${name}`}
                  onClick={() => toggleDone(number)}
                  className="cursor-pointer"
                >
                  <span
                    className={cn(
                      'flex h-8 w-8 items-center justify-center rounded-full border',
                      isDone &&
                        'bg-green-50 text-green-800 dark:bg-green-950 dark:text-green-500',
                      isFailed &&
                        'bg-red-50 text-red-800 dark:bg-red-950 dark:text-red-500'
                    )}
                  >
                    {isDone ? (
                      <Icon variant="CheckIcon" />
                    ) : isFailed ? (
                      <Icon variant="WarningIcon" />
                    ) : (
                      number
                    )}
                  </span>
                </button>
                <div className="flex flex-1 flex-col gap-1">
                  <Text weight="strong">{name}</Text>
                  <Text variant="subtext" theme="neutral">
                    {number === 1
                      ? 'Add the Nuon issuer once in this AWS account.'
                      : number === 2
                        ? 'Create a role whose trust policy fixes the exact OIDC subject.'
                        : number === 3
                          ? 'Attach the default policy or edit the capability scope.'
                          : number === 4
                            ? 'Paste the role ARN into Nuon.'
                            : 'Test OIDC isolation and every selected capability.'}
                  </Text>
                </div>
                {isDone && <Badge theme="success">Done</Badge>}
              </div>
              <div className="flex flex-col gap-4 border-t p-5">
                <CodeBlock
                  language={
                    format === 'terraform'
                      ? 'hcl'
                      : format === 'cloudformation'
                        ? 'yaml'
                        : 'bash'
                  }
                  showCopy
                >
                  {snippets[format][index]}
                </CodeBlock>
                {number === 3 && (
                  <Expand
                    id="runbook-policy"
                    isOpen={policyOpen}
                    onExpandedChange={setPolicyOpen}
                    heading={<Text weight="strong">Edit permissions</Text>}
                  >
                    <div className="flex flex-col gap-3 border-t pt-4">
                      <Text variant="subtext" theme="neutral">
                        Narrow the capability policy here. The trust policy in
                        step 2 stays unchanged.
                      </Text>
                      <Textarea
                        labelProps={{ labelText: 'Permissions policy' }}
                        value={POLICY}
                        onChange={() => {}}
                        minRows={12}
                        className="font-mono text-xs"
                      />
                    </div>
                  </Expand>
                )}
                {number === 4 && (
                  <Input
                    labelProps={{ labelText: 'IAM role ARN' }}
                    value={`arn:aws:iam::${ACCOUNT_ID}:role/nuon-cloud-connection`}
                    onChange={() => {}}
                  />
                )}
                {number === 5 && isFailed && (
                  <Banner theme="error">
                    <div>
                      <Text weight="strong">
                        Permissions probe failed at step 3
                      </Text>
                      <Text variant="subtext">
                        ecr:BatchGetImage is missing for acme/worker. Update the
                        policy in step 3, then verify again.
                      </Text>
                    </div>
                  </Banner>
                )}
                {number === 5 && allDone && (
                  <Banner theme="success">
                    <div>
                      <Text weight="strong">Connection verified</Text>
                      <Text variant="subtext">
                        The exact subject succeeded, a foreign subject was
                        denied, and both capability probes passed.
                      </Text>
                    </div>
                  </Banner>
                )}
                {number === 5 && !allDone && (
                  <Button variant="primary">Verify connection</Button>
                )}
              </div>
            </Card>
          )
        })}
      </div>
    </PageSection>
  )
}
