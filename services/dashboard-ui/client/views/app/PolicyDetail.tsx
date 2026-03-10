import { Link, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { CodeBlock } from '@/components/common/CodeBlock'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
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

  const policyComponents = policy?.components ?? []
  const isAllComponents =
    policyComponents.length === 0 ||
    (policyComponents.length === 1 && policyComponents[0] === '*')

  return (
    <PageSection isScrollable>
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
          <span className="flex items-center gap-3">
            <Icon variant="ShieldCheck" size={24} />
            <Text variant="base" weight="strong">
              {policy?.name}
            </Text>
          </span>
          {policy?.id && <ID>{policy.id}</ID>}
        </HeadingGroup>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        {policy?.type && (
          <Badge theme="default" size="md">
            {formatPolicyType(policy.type)}
          </Badge>
        )}
        {policy?.engine && (
          <Badge theme={policy.engine === 'kyverno' ? 'brand' : 'info'} size="md">
            {policy.engine.toUpperCase()}
          </Badge>
        )}
        {policy?.created_at && (
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
                textToCopy={policy?.contents ?? ''}
                className="w-fit"
              />
            </div>
            <div className="p-4">
              <CodeBlock
                language={getCodeLanguage(policy?.engine ?? '')}
                className="max-h-[600px]"
                showLineNumbers
              >
                {policy?.contents ?? ''}
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
                        to={`/${org?.id}/apps/${app?.id}/components/${componentId}`}
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
