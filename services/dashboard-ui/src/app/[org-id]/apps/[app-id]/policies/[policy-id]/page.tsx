import type { Metadata } from 'next'
import Link from 'next/link'
import { notFound } from 'next/navigation'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { CodeBlock } from '@/components/common/CodeBlock'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getAppPolicy, getComponents, getOrg } from '@/lib'
import type { TPageProps } from '@/types'

type TPolicyPageProps = TPageProps<'org-id' | 'app-id' | 'policy-id'>

function formatPolicyType(type: string): string {
  return type
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

function getCodeLanguage(engine: string): 'yaml' | 'rego' {
  return engine === 'opa' ? 'rego' : 'yaml'
}

export async function generateMetadata({
  params,
}: TPolicyPageProps): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['policy-id']: policyId,
  } = await params
  const [{ data: app }, { data: policy }] = await Promise.all([
    getApp({ appId, orgId }),
    getAppPolicy({ appId, orgId, policyId }),
  ])

  const policyName = policy?.name || 'Policy'

  return {
    title: `${policyName} | ${app?.name} | Nuon`,
  }
}

export default async function PolicyDetailPage({ params }: TPolicyPageProps) {
  const {
    ['org-id']: orgId,
    ['app-id']: appId,
    ['policy-id']: policyId,
  } = await params

  const [
    { data: app },
    { data: org },
    { data: policy, error, status },
    { data: components },
  ] = await Promise.all([
    getApp({ appId, orgId }),
    getOrg({ orgId }),
    getAppPolicy({ appId, orgId, policyId }),
    getComponents({ appId, orgId }),
  ])

  if (error || !policy) {
    if (status === 404) {
      notFound()
    }
    notFound()
  }

  const componentNameToId: Record<string, string> = {}
  components?.forEach((c) => {
    if (c.name && c.id) {
      componentNameToId[c.name] = c.id
    }
  })

  const policyComponents = policy.components || []
  const isAllComponents =
    !policyComponents ||
    policyComponents.length === 0 ||
    (policyComponents.length === 1 && policyComponents[0] === '*')

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${orgId}`, text: org?.name },
          { path: `/${orgId}/apps`, text: 'Apps' },
          { path: `/${orgId}/apps/${appId}`, text: app?.name },
          { path: `/${orgId}/apps/${appId}/policies`, text: 'Policies' },
          {
            path: `/${orgId}/apps/${appId}/policies/${policyId}`,
            text: policy.name,
          },
        ]}
      />

      <div className="flex items-center justify-between">
        <HeadingGroup>
          <BackLink className="mb-4" />
          <span className="flex items-center gap-3">
            <Icon variant="ShieldCheck" size={24} />
            <Text variant="base" weight="strong">
              {policy.name}
            </Text>
          </span>
          <ID>{policy.id}</ID>
        </HeadingGroup>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <Badge theme="default" size="md">
          {formatPolicyType(policy.type || '')}
        </Badge>
        <Badge theme={policy.engine === 'kyverno' ? 'brand' : 'info'} size="md">
          {(policy.engine || '').toUpperCase()}
        </Badge>
        {policy.created_at && (
          <Text variant="subtext" className="flex items-center gap-1">
            Created <Time time={policy.created_at} format="relative" />
          </Text>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mt-4">
        <div className="lg:col-span-2">
          <Card className="!p-0 overflow-hidden">
            <div className="flex items-center justify-between p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
              <Text weight="strong" variant="body">
                Policy Definition
              </Text>
              <ClickToCopyButton
                textToCopy={policy.contents || ''}
                className="w-fit"
              />
            </div>
            <div className="p-4">
              <CodeBlock
                language={getCodeLanguage(policy.engine || '')}
                className="max-h-[600px]"
                showLineNumbers
              >
                {policy.contents || ''}
              </CodeBlock>
            </div>
          </Card>
        </div>

        <div className="lg:col-span-1">
          <Card className="!p-0 overflow-hidden">
            <div className="flex items-center justify-between p-4 border-b border-cool-grey-200 dark:border-dark-grey-600">
              <Text weight="strong" variant="body">
                Applicable Components
              </Text>
            </div>
            <div className="p-4">
              {isAllComponents ? (
                <div className="flex items-center gap-2 text-grey-600 dark:text-grey-400">
                  <Icon variant="Stack" size={16} />
                  <Text variant="body" className="italic">
                    All components
                  </Text>
                </div>
              ) : (
                <div className="flex flex-col gap-2">
                  {policyComponents.map((comp) => {
                    const componentId = componentNameToId[comp]
                    return componentId ? (
                      <Link
                        key={comp}
                        href={`/${orgId}/apps/${appId}/components/${componentId}`}
                        className="flex items-center gap-2 rounded px-3 py-2 text-sm border border-cool-grey-200 dark:border-dark-grey-600 hover:bg-grey-50 dark:hover:bg-dark-grey-800 transition-colors"
                      >
                        <Icon variant="Cube" size={14} />
                        <Text variant="body">{comp}</Text>
                        <Icon
                          variant="ArrowSquareOut"
                          size={12}
                          className="ml-auto text-grey-400"
                        />
                      </Link>
                    ) : (
                      <div
                        key={comp}
                        className="flex items-center gap-2 rounded px-3 py-2 text-sm border border-cool-grey-200 dark:border-dark-grey-600"
                      >
                        <Icon variant="Cube" size={14} />
                        <Text variant="body">{comp}</Text>
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          </Card>
        </div>
      </div>
    </PageSection>
  )
}
