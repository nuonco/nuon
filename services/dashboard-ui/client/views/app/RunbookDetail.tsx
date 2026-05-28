import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Markdown } from '@/components/common/Markdown'
import { Text } from '@/components/common/Text'
import { RunbookStep } from '@/components/runbooks/RunbookStep'
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
    queryFn: () =>
      getRunbook({ orgId: org!.id, appId: app!.id, runbookId: runbookId! }),
    enabled: !!org?.id && !!app?.id && !!runbookId,
  })

  const latestConfig = runbook?.configs?.[0]
  const steps =
    latestConfig?.steps
      ?.slice()
      .sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)) ?? []

  return (
    <PageSection flush className="flex-1">
      <PageTitle title={`${runbook?.name ?? 'Runbook'} | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          {
            path: `/${org?.id}/apps/${app?.id}/runbooks`,
            text: 'Runbooks',
          },
          {
            path: `/${org?.id}/apps/${app?.id}/runbooks/${runbookId}`,
            text: runbook?.name,
          },
        ]}
      />

      <div className="@container flex flex-col flex-1">
        <header className="p-6 border-b flex flex-col gap-6">
          <div className="flex flex-wrap items-start gap-4 justify-between w-full">
            <HeadingGroup>
              <BackLink className="mb-4" />
              <Text variant="h3" weight="strong">
                {runbook?.name}
              </Text>
              <span className="flex flex-wrap items-center gap-4 mt-1">
                {runbookId ? <ID>{runbookId}</ID> : null}
                {runbook?.labels &&
                Object.keys(runbook.labels).length > 0 ? (
                  <span className="flex flex-wrap gap-1">
                    {Object.keys(runbook.labels)
                      .sort()
                      .map((k) => (
                        <Badge
                          key={k}
                          variant="code"
                          size="sm"
                          theme="neutral"
                        >
                          {k}: {runbook.labels[k]}
                        </Badge>
                      ))}
                  </span>
                ) : null}
              </span>
              {runbook?.description ? (
                <Text variant="subtext" theme="neutral">
                  {runbook.description}
                </Text>
              ) : null}
            </HeadingGroup>
          </div>

          <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
            <LabeledValue label="Steps">
              <Text variant="subtext">{steps.length}</Text>
            </LabeledValue>
          </div>
        </header>

        <div className="flex flex-col gap-6">
          {latestConfig?.readme ? (
            <PageSection className="flex flex-col gap-4">
              <Text variant="base" weight="strong">
                Readme
              </Text>
              <Markdown content={latestConfig.readme} mode="app" />
            </PageSection>
          ) : null}

          <PageSection className="flex flex-col gap-4">
            <Text variant="base" weight="strong">
              Steps
            </Text>
            {steps.length ? (
              <div className="grid grid-cols-1 gap-4">
                {steps.map((step, i) => (
                  <RunbookStep key={step.id ?? i} index={i} step={step} actionBasePath={`/${org?.id}/apps/${app?.id}`} />
                ))}
              </div>
            ) : (
              <Text theme="neutral">No steps configured.</Text>
            )}
          </PageSection>
        </div>
      </div>
    </PageSection>
  )
}
