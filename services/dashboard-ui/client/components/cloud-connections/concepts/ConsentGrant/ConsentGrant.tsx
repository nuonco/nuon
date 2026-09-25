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
import { ACCOUNT_ID, CAPABILITIES, CUSTOM_POLICY, scriptFor } from './mockData'

type TResult = 'idle' | 'verified' | 'failed'
interface IConsentGrant {
  initialStacks?: boolean
  initialRepos?: string[]
  customOpen?: boolean
  result?: TResult
}

const Capability = ({
  name,
  enabled,
  onChange,
  children,
}: {
  name: keyof typeof CAPABILITIES
  enabled: boolean
  onChange: (value: boolean) => void
  children: React.ReactNode
}) => {
  const capability = CAPABILITIES[name]
  return (
    <div className="flex flex-col gap-4 border-b py-5 last:border-b-0">
      <div className="flex items-start gap-4">
        <Icon
          variant={name === 'stacks' ? 'StackIcon' : 'ShippingContainerIcon'}
          size={24}
          theme={enabled ? 'brand' : 'neutral'}
        />
        <div className="flex min-w-0 flex-1 flex-col gap-2">
          <Text variant="base" weight="strong">
            {capability.title}
          </Text>
          <Text>
            <strong>What:</strong> {capability.what}
          </Text>
          <Text variant="subtext" theme="neutral">
            <strong>Why:</strong> {capability.why}
          </Text>
        </div>
        <Toggle checked={enabled} onChange={onChange} />
      </div>
      {enabled && (
        <div className="ml-10 flex flex-col gap-4">
          {children}
          <Expand
            id={`consent-${name}-actions`}
            heading={
              <Text variant="subtext" weight="strong">
                Permissions this grants
              </Text>
            }
          >
            <Text
              family="mono"
              variant="subtext"
              theme="neutral"
              className="border-t p-3 break-words"
            >
              {capability.actions}
            </Text>
          </Expand>
        </div>
      )}
    </div>
  )
}

export const ConsentGrant = ({
  initialStacks = true,
  initialRepos = ['*'],
  customOpen = false,
  result = 'idle',
}: IConsentGrant) => {
  const [stacks, setStacks] = useState(initialStacks)
  const [images, setImages] = useState(true)
  const [repos, setRepos] = useState(initialRepos)
  const [region, setRegion] = useState('us-east-1')
  const [prefix, setPrefix] = useState('nuon-*')
  const [advanced, setAdvanced] = useState(customOpen)
  const repoValue = repos.join(', ')
  const setRepoValue = (value: string) =>
    setRepos(
      value
        .split(',')
        .map((repo) => repo.trim())
        .filter(Boolean)
    )
  const scripts = {
    terraform: (
      <CodeBlock language="hcl" showCopy>
        {scriptFor('terraform', stacks, images, repos, region, prefix)}
      </CodeBlock>
    ),
    cli: (
      <CodeBlock language="bash" showCopy>
        {scriptFor('cli', stacks, images, repos, region, prefix)}
      </CodeBlock>
    ),
    cloudformation: (
      <div className="flex flex-col gap-3">
        <Button variant="secondary">Open quick-create in AWS</Button>
        <CodeBlock language="yaml" showCopy>
          {scriptFor('cloudformation', stacks, images, repos, region, prefix)}
        </CodeBlock>
      </div>
    ),
  }

  return (
    <PageSection className="min-h-screen bg-white dark:bg-dark-grey-900">
      <SectionHeader
        variant="page"
        title={`Nuon is asking for access to AWS account ${ACCOUNT_ID}`}
        description="Choose exactly what this Nuon connection may do. The setup on the right updates with every change."
        status={<Badge theme="info">OIDC · exact subject</Badge>}
      />
      <div className="grid flex-1 gap-6 xl:grid-cols-[minmax(30rem,1fr)_minmax(34rem,0.9fr)]">
        <Card className="!gap-0">
          <Capability name="stacks" enabled={stacks} onChange={setStacks}>
            <Input
              labelProps={{ labelText: 'Install stack name prefix' }}
              value={prefix}
              onChange={(event) => setPrefix(event.currentTarget.value)}
              helperText="Only stacks with this prefix are in scope."
            />
          </Capability>
          <Capability name="images" enabled={images} onChange={setImages}>
            <div className="grid gap-4 sm:grid-cols-2">
              <Input
                labelProps={{ labelText: 'ECR repositories' }}
                value={repoValue}
                onChange={(event) => setRepoValue(event.currentTarget.value)}
                helperText="Comma-separated. Use * for all repositories."
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
            </div>
          </Capability>
          <Expand
            id="consent-advanced"
            isOpen={advanced}
            onExpandedChange={setAdvanced}
            heading={
              <div className="flex items-center gap-2">
                <Icon variant="SlidersHorizontalIcon" />
                <Text weight="strong">Advanced: edit the policy yourself</Text>
              </div>
            }
          >
            <div className="flex flex-col gap-3 border-t p-4">
              <Banner theme="info">
                The permissions policy can change. The rendered OIDC trust
                policy must stay unchanged.
              </Banner>
              <Textarea
                labelProps={{ labelText: 'Custom permissions policy' }}
                value={CUSTOM_POLICY}
                onChange={() => {}}
                minRows={12}
                className="font-mono text-xs"
              />
            </div>
          </Expand>
        </Card>
        <aside className="min-w-0">
          <Card className="!gap-5 xl:sticky xl:top-6">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <Text variant="base" weight="strong">
                Run this in AWS
              </Text>
              <Text variant="subtext" theme="neutral">
                Live setup
              </Text>
            </div>
            <Tabs
              tabLabels={{ cli: 'AWS CLI', cloudformation: 'CloudFormation' }}
              tabs={scripts}
            />
            <Input
              labelProps={{ labelText: 'IAM role ARN' }}
              value={`arn:aws:iam::${ACCOUNT_ID}:role/nuon-cloud-connection`}
              onChange={() => {}}
            />
            {result === 'verified' && (
              <Banner theme="success">
                <div>
                  <Text weight="strong">Connection verified</Text>
                  <Text variant="subtext">
                    OIDC, foreign-subject, and capability probes passed.
                  </Text>
                </div>
              </Banner>
            )}
            {result === 'failed' && (
              <Banner theme="error">
                <div>
                  <Text weight="strong">Foreign-subject probe failed</Text>
                  <Text variant="subtext">
                    Restrict api.nuon.co:sub to the rendered subject, then
                    verify again.
                  </Text>
                </div>
              </Banner>
            )}
            <Button variant="primary">Verify connection</Button>
          </Card>
        </aside>
      </div>
    </PageSection>
  )
}
