import { Link, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { CodeBlock } from '@/components/common/CodeBlock'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getAppPolicy, getComponents } from '@/lib'

function formatPolicyType(type: string): string {
  return type
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

function getCodeLanguage(engine: string): string {
  return engine === 'opa' ? 'rego' : 'yaml'
}

export const PolicyDetail = () => {
  const { policyId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()

  const { data: policyResult } = useQuery({
    queryKey: ['app-policy', org?.id, app?.id, policyId],
    queryFn: () =>
      getAppPolicy({ orgId: org.id, appId: app.id, policyId: policyId! }),
    enabled: !!org?.id && !!app?.id && !!policyId,
  })

  const { data: componentsResult } = useQuery({
    queryKey: ['app-components', org?.id, app?.id],
    queryFn: () => getComponents({ orgId: org.id, appId: app.id }),
    enabled: !!org?.id && !!app?.id,
  })

  const policy = policyResult
  const components = componentsResult?.data ?? []

  const componentNameToId: Record<string, string> = {}
  components.forEach((c) => {
    if (c.name && c.id) {
      componentNameToId[c.name] = c.id
    }
  })

  const isSandboxPolicy = policy?.type === 'sandbox'
  const policyComponents = policy?.components ?? []
  const isAllComponents =
    !isSandboxPolicy &&
    (policyComponents.length === 0 ||
      (policyComponents.length === 1 && policyComponents[0] === '*'))

  return (
    <PageSection>
      <PageTitle title={`${policy?.name ?? 'Policy'} | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/policies`, text: 'Policies' },
          {
            path: `/${org?.id}/apps/${app?.id}/policies/${policyId}`,
            text: policy?.name,
          },
        ]}
      />

      <div className="flex items-center justify-between">
        <HeadingGroup>
          <BackLink className="mb-4" />
          <Text variant="base" weight="strong">
            {policy?.name}
          </Text>
          {policy?.id && <ID>{policy.id}</ID>}
        </HeadingGroup>
      </div>

      <div className="flex items-start gap-10">
        {policy?.type && (
          <LabeledValue label="Type">
            <Text variant="subtext">{formatPolicyType(policy.type)}</Text>
          </LabeledValue>
        )}
        {policy?.engine && (
          <LabeledValue label="Engine">
            <Text variant="subtext">{policy.engine}</Text>
          </LabeledValue>
        )}
        {policy?.created_at && (
          <LabeledValue label="Created">
            <Time variant="subtext" time={policy.created_at} format="relative" />
          </LabeledValue>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <Card className="!p-5 !gap-4">
            <div className="flex items-center justify-between">
              <Text weight="strong">Policy Definition</Text>
              <ClickToCopyButton
                textToCopy={policy?.contents ?? ''}
                className="w-fit"
              />
            </div>
            <CodeBlock
              language={getCodeLanguage(policy?.engine ?? '')}
              showLineNumbers
              className="!max-h-none"
            >
              {policy?.contents ?? ''}
            </CodeBlock>
          </Card>
        </div>

        <div className="lg:col-span-1">
          <Card className="!p-5 !gap-4">
            <Text weight="strong">Applicable Components</Text>
            {isSandboxPolicy ? (
              <div className="flex items-center gap-2">
                <Icon variant="ShippingContainerIcon" size={16} />
                <Text variant="subtext">Sandbox</Text>
              </div>
            ) : isAllComponents ? (
              <div className="flex items-center gap-2">
                <Icon variant="CardsIcon" size={16} />
                <Text variant="subtext">All components</Text>
              </div>
            ) : (
              <div className="flex flex-col gap-2">
                {policyComponents.map((comp) => {
                  const componentId = componentNameToId[comp]
                  return componentId ? (
                    <Link
                      key={comp}
                      to={`/${org?.id}/apps/${app?.id}/components/${componentId}`}
                      className="flex items-center gap-2 rounded px-3 py-2 text-sm border border-cool-grey-200 dark:border-dark-grey-600 hover:bg-grey-50 dark:hover:bg-dark-grey-800 transition-colors"
                    >
                      <Icon variant="CardsIcon" size={14} />
                      <Text variant="body">{comp}</Text>
                      <Icon
                        variant="ArrowSquareOutIcon"
                        size={12}
                        className="ml-auto text-grey-400"
                      />
                    </Link>
                  ) : (
                    <div
                      key={comp}
                      className="flex items-center gap-2 rounded px-3 py-2 text-sm border border-cool-grey-200 dark:border-dark-grey-600"
                    >
                      <Icon variant="CardsIcon" size={14} />
                      <Text variant="body">{comp}</Text>
                    </div>
                  )
                })}
              </div>
            )}
          </Card>
        </div>
      </div>
    </PageSection>
  )
}
