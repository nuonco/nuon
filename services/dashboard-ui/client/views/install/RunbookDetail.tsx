import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Markdown } from '@/components/common/Markdown'
import { StatusWithDescription } from '@/components/common/StatusWithDescription'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RunRunbookButton } from '@/components/runbooks/RunRunbook/RunRunbook'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallRunbook, getInstallRunbookRuns } from '@/lib'

export const RunbookDetail = () => {
  const { runbookId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()

  const { data: runbook } = useQuery({
    queryKey: ['install-runbook', org?.id, install?.id, runbookId],
    queryFn: () =>
      getInstallRunbook({
        orgId: org!.id,
        installId: install!.id,
        runbookId: runbookId!,
      }),
    enabled: !!org?.id && !!install?.id && !!runbookId,
  })

  const { data: runsResult } = useQuery({
    queryKey: ['install-runbook-runs', org?.id, install?.id, runbookId],
    queryFn: () =>
      getInstallRunbookRuns({
        orgId: org!.id,
        installId: install!.id,
        limit: 10,
        offset: 0,
      }),
    enabled: !!org?.id && !!install?.id,
    refetchInterval: 20000,
  })

  const steps = runbook?.steps ?? []
  const runs = (runsResult?.data ?? []).filter((r) => r.runbook_id === runbookId)

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

      <div className="@container flex flex-col flex-1">
        <header className="p-6 border-b flex flex-wrap items-start gap-4 justify-between w-full">
          <HeadingGroup>
            <BackLink className="mb-4" />
            <Text variant="h3" weight="strong">
              {runbook?.name}
            </Text>
            {runbookId ? <ID>{runbookId}</ID> : null}
            {runbook?.description ? (
              <Text variant="subtext" theme="neutral">
                {runbook.description}
              </Text>
            ) : null}
          </HeadingGroup>

          {runbook ? (
            <RunRunbookButton runbook={runbook} variant="primary" />
          ) : null}
        </header>

        <div className="grid grid-cols-1 @5xl:grid-cols-12 flex-1">
          <div className="@5xl:col-span-8 flex flex-col gap-6">
            {runbook?.readme ? (
              <PageSection>
                <Markdown content={runbook.readme} mode="app" />
              </PageSection>
            ) : null}

            <PageSection className="flex flex-col gap-4">
              <Text variant="base" weight="strong">
                Steps
              </Text>
              {steps.length ? (
                <div className="grid grid-cols-1 gap-4">
                  {steps.map((step, i) => (
                    <div
                      key={step.id ?? i}
                      className="border rounded-lg p-4 flex flex-col gap-1"
                    >
                      <Text variant="body" weight="strong">
                        {i + 1}. {step.name}
                      </Text>
                      {step.description ? (
                        <Text variant="subtext" theme="neutral">
                          {step.description}
                        </Text>
                      ) : null}
                      {step.type ? (
                        <Text variant="subtext" theme="neutral">
                          Type: {step.type}
                        </Text>
                      ) : null}
                    </div>
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
                {runs.map((run) => (
                  <div key={run.id} className="border rounded-lg p-3 flex flex-col gap-1">
                    <div className="flex items-center justify-between gap-2">
                      <StatusWithDescription
                        statusProps={{ status: run.status_v2?.status }}
                        tooltipProps={{
                          position: 'top',
                          tipContent: run.status_v2?.status_human_description,
                        }}
                      />
                      <Time
                        variant="subtext"
                        time={run.created_at}
                        format="relative"
                        shouldTick
                      />
                    </div>
                    <ID>{run.id}</ID>
                  </div>
                ))}
              </div>
            ) : (
              <Text variant="subtext" theme="neutral">
                No runs yet.
              </Text>
            )}
          </PageSection>
        </div>
      </div>
    </PageSection>
  )
}
