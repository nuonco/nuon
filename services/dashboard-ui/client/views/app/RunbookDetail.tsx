import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Expand } from '@/components/common/Expand'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getRunbook } from '@/lib'

export const RunbookDetail = () => {
  const { runbookId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()

  const { data: runbook } = useQuery({
    queryKey: ['runbook', org?.id, app?.id, runbookId],
    queryFn: () => getRunbook({ orgId: org!.id, appId: app!.id, runbookId: runbookId! }),
    enabled: !!org?.id && !!app?.id && !!runbookId,
  })

  const latestConfig = runbook?.configs?.[0]
  const steps = latestConfig?.steps?.slice().sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)) ?? []

  return (
    <PageSection flush>
      <PageTitle title={`${runbook?.name ?? 'Runbook'} | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/runbooks`, text: 'Runbooks' },
          {
            path: `/${org?.id}/apps/${app?.id}/runbooks/${runbookId}`,
            text: runbook?.name,
          },
        ]}
      />

      <div className="p-6 border-b flex items-start justify-between gap-8">
        <HeadingGroup>
          <BackLink className="mb-6" />
          <Text variant="base" weight="strong">
            {runbook?.name}
          </Text>
          {runbookId ? <ID>{runbookId}</ID> : null}
          {runbook?.labels && Object.keys(runbook.labels).length > 0 ? (
            <span className="flex flex-wrap gap-1 mt-1">
              {Object.keys(runbook.labels)
                .sort()
                .map((k) => (
                  <Badge key={k} variant="code" size="sm" theme="neutral">
                    {k}: {runbook.labels[k]}
                  </Badge>
                ))}
            </span>
          ) : null}
          {runbook?.description ? (
            <Text variant="subtext" theme="neutral">
              {runbook.description}
            </Text>
          ) : null}
        </HeadingGroup>

        <div className="flex flex-row gap-6 items-start">
          <LabeledValue label="Steps">
            <Text variant="subtext">{steps.length}</Text>
          </LabeledValue>
        </div>
      </div>

      {latestConfig?.readme ? (
        <PageSection>
          <Expand
            heading={<Text variant="base" weight="strong">README</Text>}
            id="runbook-readme"
            className="border rounded-md"
          >
            <div className="p-4 border-t">
              <Markdown content={latestConfig.readme} mode="app" />
            </div>
          </Expand>
        </PageSection>
      ) : null}

      <PageSection>
        <Text variant="base" weight="strong">
          Steps
        </Text>
        {steps.length ? (
          <div className="grid grid-cols-1 gap-4">
            {steps.map((step, i) => (
              <Expand
                key={step.id ?? i}
                className="border rounded-md"
                heading={
                  <Text weight="strong">
                    {i + 1}. {step.name}
                  </Text>
                }
                id={`step-${i}`}
                isOpen
              >
                <div className="flex flex-col gap-4 p-4 border-t">
                  <div className="flex gap-4">
                    <LabeledValue label="Type">
                      <Badge variant="code" size="sm" theme="neutral">
                        {step.type}
                      </Badge>
                    </LabeledValue>
                    {step.component_name ? (
                      <LabeledValue label="Component">
                        <Text variant="subtext">{step.component_name}</Text>
                      </LabeledValue>
                    ) : null}
                    {step.type === 'deploy' ? (
                      <LabeledValue label="Deploy dependencies">
                        <Badge variant="code" size="sm" theme={step.deploy_dependencies ? 'info' : 'neutral'}>
                          {step.deploy_dependencies ? 'Yes' : 'No'}
                        </Badge>
                      </LabeledValue>
                    ) : null}
                    {step.role ? (
                      <LabeledValue label="Role">
                        <Text variant="subtext">{step.role}</Text>
                      </LabeledValue>
                    ) : null}
                  </div>
                  {step.env_vars && Object.keys(step.env_vars).length > 0 ? (
                    <div className="flex flex-col gap-2">
                      <Text weight="strong">Environment variables</Text>
                      <div className="flex flex-wrap gap-1">
                        {Object.entries(step.env_vars).sort(([a], [b]) => a.localeCompare(b)).map(([k, v]) => (
                          <Badge key={k} variant="code" size="sm" theme="neutral">
                            {k}={v}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  ) : null}
                  {step.command ? (
                    <div className="flex flex-col gap-2">
                      <Text weight="strong">Command</Text>
                      <CodeBlock language="bash">{step.command}</CodeBlock>
                    </div>
                  ) : null}
                  {step.inline_contents ? (
                    <div className="flex flex-col gap-2">
                      <Text weight="strong">Inline contents</Text>
                      <CodeBlock language="bash">{step.inline_contents}</CodeBlock>
                    </div>
                  ) : null}
                </div>
              </Expand>
            ))}
          </div>
        ) : (
          <Text theme="neutral">No steps configured.</Text>
        )}
      </PageSection>
    </PageSection>
  )
}
