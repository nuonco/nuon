import { useEffect } from 'react'
import { useLocation } from 'react-router'
import { Banner } from '@/components/common/Banner'
import { CodeBlock } from '@/components/common/CodeBlock'
import { LabeledValue } from '@/components/common/LabeledValue'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { Tabs } from '@/components/common/Tabs'
import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import type { TCloudConnection, TCloudConnectionSetup } from '@/types'
import {
  ConnectionPolicies,
  failedRunbookStep,
  roleName,
} from '../ConnectionPolicies'

export const RunInCloud = ({
  connection,
  setup,
}: {
  connection: TCloudConnection
  setup: TCloudConnectionSetup
}) => {
  const { hash } = useLocation()
  useEffect(() => {
    if (hash)
      document.getElementById(hash.slice(1))?.scrollIntoView({ block: 'start' })
  }, [hash, setup])
  const snippet = (format: 'cli' | 'terraform' | 'cloudformation') =>
    setup[format]?.replaceAll('<role-name>', roleName(connection.principal)) ||
    ''
  return (
    <div className="flex min-w-0 flex-col gap-6">
      <SectionHeader
        title="Run the setup in AWS"
        description="Complete the steps in order. Nuon never writes IAM resources into the account."
      />
      {connection.status === 'error' && (
        <Banner theme="error">
          Runbook step {failedRunbookStep(connection)} needs attention.{' '}
          {connection.status_message}
        </Banner>
      )}
      {connection.preset === 'custom' && (
        <Banner theme="info">
          Custom access omits the permissions step. Attach your permissions
          policy to the role in AWS.
        </Banner>
      )}
      <LabeledValue label="Role ARN">
        <ClickToCopy>
          <Text family="mono" variant="subtext" className="break-all">
            {connection.principal}
          </Text>
        </ClickToCopy>
      </LabeledValue>
      <ol className="list-decimal pl-5 space-y-2">
        <li id="runbook-step-1" className="scroll-mt-6">
          <Text as="div" weight="strong">
            Create the OIDC provider
          </Text>
          <Text as="div" variant="subtext" theme="neutral">
            Add the Nuon issuer to this AWS account once. If it already exists,
            reuse it.
          </Text>
        </li>
        <li>
          <Text as="div" weight="strong">
            Create the role with the trust policy
          </Text>
          <Text as="div" variant="subtext" theme="neutral">
            Use the exact audience and connection subject rendered by Nuon.
          </Text>
        </li>
        {connection.preset === 'stacks' && (
          <li>
            <Text as="div" weight="strong">
              Attach the permissions policy
            </Text>
            <Text as="div" variant="subtext" theme="neutral">
              Review the IAM actions and scope below before attaching it.
            </Text>
          </li>
        )}
      </ol>
      <Tabs
        initActiveTab="cli"
        tabLabels={{
          cli: 'AWS CLI',
          terraform: 'Terraform',
          cloudformation: 'CloudFormation',
        }}
        tabs={{
          cli: (
            <CodeBlock language="bash" showCopy wrapLongLines>
              {snippet('cli')}
            </CodeBlock>
          ),
          terraform: (
            <CodeBlock language="hcl" showCopy>
              {snippet('terraform')}
            </CodeBlock>
          ),
          cloudformation: (
            <div className="flex flex-col gap-3">
              <Text variant="subtext" theme="neutral">
                Add these resources under the Resources section of a
                CloudFormation template.
              </Text>
              <CodeBlock language="yaml" showCopy>
                {snippet('cloudformation')}
              </CodeBlock>
            </div>
          ),
        }}
      />
      <ConnectionPolicies connection={connection} />
    </div>
  )
}
