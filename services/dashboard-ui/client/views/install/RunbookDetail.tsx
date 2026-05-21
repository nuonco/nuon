import { Link, useParams } from 'react-router'
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
import { Time } from '@/components/common/Time'
import { RunRunbookButton } from '@/components/runbooks/RunRunbook/RunRunbook'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallRunbook } from '@/lib'

export const RunbookDetail = () => {
  const { runbookId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: installRunbook } = useQuery({
    queryKey: ['install-runbook', org?.id, install?.id, runbookId],
    queryFn: () =>
      getInstallRunbook({
        orgId: org!.id,
        installId: install!.id,
        runbookId: runbookId!,
      }),
    enabled: !!org?.id && !!install?.id && !!runbookId,
    refetchInterval: 20000,
  })

  const runbook = installRunbook?.runbook
  const latestConfig = runbook?.configs?.[0]
  const steps = latestConfig?.steps?.slice().sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)) ?? []
  const runs = installRunbook?.runs ?? []

  return (
    <PageSection flush className="flex-1">
      <PageTitle title={`${runbook?.name ?? 'Runbook'} | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/runbooks`,
            text: 'Runbooks',
          },
          {
            path: `/${org?.id}/installs/${install?.id}/runbooks/${runbookId}`,
            text: runbook?.name,
          },
        ]}
      />

      <div className="p-6 border-b flex flex-wrap items-start gap-4 justify-between w-full">
        <HeadingGroup>
          <BackLink className="mb-4" />
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
          {installRunbook ? (
            <RunRunbookButton installRunbook={installRunbook} variant="primary" />
          ) : null}
        </div>
      </div>

      <div className="@container grid grid-cols-1 @5xl:grid-cols-12 flex-1">
        <div className="@5xl:col-span-8 flex flex-col gap-6">
          {latestConfig?.readme ? (
            <PageSection>
              <Expand
                heading={<Text variant="base" weight="strong">README</Text>}
                id="runbook-readme"
                className="border rounded-md"
              >
                <div className="p-4 border-t">
                  <Markdown content={latestConfig.readme} mode="install" />
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
        </div>

        <PageSection className="hidden @5xl:flex flex-col @5xl:col-span-4 gap-4">
          <Text variant="base" weight="strong">
            Run history
          </Text>
          {runs.length ? (
            <div className="flex flex-col gap-3">
              {runs.map((run) => {
                const wfStatus = typeof run.install_workflow?.status === 'object'
                  ? (run.install_workflow.status as { status?: string })?.status
                  : run.install_workflow?.status
                const status = wfStatus ?? run.status ?? 'unknown'
                const workflowId = run.install_workflow_id ?? run.install_workflow?.id
                return (
                  <Link
                    key={run.id}
                    to={workflowId ? `/${org?.id}/installs/${install?.id}/workflows/${workflowId}` : '#'}
                    className="border rounded-lg p-3 flex flex-col gap-1 hover:bg-neutral-50 dark:hover:bg-neutral-800 transition-colors"
                  >
                    <div className="flex items-center justify-between gap-2">
                      <Badge variant="code" size="sm" theme={
                        status === 'completed' ? 'success' :
                        status === 'error' ? 'error' :
                        status === 'in-progress' ? 'info' :
                        'neutral'
                      }>
                        {status}
                      </Badge>
                      <Time
                        variant="subtext"
                        time={run.created_at}
                        format="relative"
                        shouldTick
                      />
                    </div>
                    <ID>{run.id}</ID>
                  </Link>
                )
              })}
            </div>
          ) : (
            <Text variant="subtext" theme="neutral">
              No runs yet.
            </Text>
          )}
        </PageSection>
      </div>
    </PageSection>
  )
}
