import { useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Link } from '@/components/common/Link'
import { Card } from '@/components/common/Card'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
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
  const { policyId, branchId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()
  const appBase = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}`
    : `/${org?.id}/apps/${app?.id}`

  const { data: policyResult } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['app-policy', org?.id, app?.id, policyId],
    queryFn: () =>
      getAppPolicy({ orgId: org.id, appId: app.id, policyId: policyId! }),
    enabled: !!org?.id && !!app?.id && !!policyId,
  })

  const { data: componentsResult } = useQuery({
    placeholderData: keepPreviousData,
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
  const isAllComponents = !isSandboxPolicy && policyComponents.includes('*')
  const hasNoComponents =
    !isSandboxPolicy && !isAllComponents && policyComponents.length === 0

  return (
    <>
      <PageTitle segments={[policy?.name ?? 'Policy', app?.name]} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `${appBase}/policies`, text: 'Policies' },
          {
            path: `${appBase}/policies/${policyId}`,
            text: policy?.name,
          },
        ]}
      />

      <DetailPage
        header={
          <DetailHeader
            title={policy?.name}
            id={policy?.id}
            metadata={
              <>
                {policy?.type ? (
                  <LabeledValue label="Type">
                    <Text variant="subtext">
                      {formatPolicyType(policy.type)}
                    </Text>
                  </LabeledValue>
                ) : null}
                {policy?.engine ? (
                  <LabeledValue label="Engine">
                    <Text variant="subtext">{policy.engine}</Text>
                  </LabeledValue>
                ) : null}
                {policy?.created_at ? (
                  <LabeledValue label="Created">
                    <Time
                      variant="subtext"
                      time={policy.created_at}
                      format="relative"
                    />
                  </LabeledValue>
                ) : null}
              </>
            }
          />
        }
      >
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
              ) : hasNoComponents ? (
                <div className="flex items-center gap-2">
                  <Icon variant="ProhibitIcon" size={16} />
                  <Text variant="subtext">No components</Text>
                </div>
              ) : (
                <div className="flex flex-col gap-2">
                  {policyComponents.map((comp) => {
                    const componentId = componentNameToId[comp]
                    return componentId ? (
                      <Link
                        key={comp}
                        href={`${appBase}/components/${componentId}`}
                        className="flex items-center gap-2 rounded px-3 py-2 border border-cool-grey-200 dark:border-dark-grey-600 hover:bg-grey-50 dark:hover:bg-dark-grey-800 transition-colors"
                      >
                        <Icon variant="CardsIcon" size={14} />
                        <Text variant="body">{comp}</Text>
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
      </DetailPage>
    </>
  )
}
