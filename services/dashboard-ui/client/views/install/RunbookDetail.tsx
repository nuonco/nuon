import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabeledStatus } from '@/components/common/LabeledStatus'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Markdown } from '@/components/common/Markdown'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { RunbookStep } from '@/components/runbooks/RunbookStep'
import { RunRunbookButton } from '@/components/runbooks/RunRunbook/RunRunbook'
import { RunbookRunTimeline } from '@/components/runbooks/RunbookRunTimeline'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { Panel } from '@/components/surfaces/Panel'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getInstallRunbook } from '@/lib'

export const RunbookDetail = () => {
  const { runbookId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const { addPanel } = useSurfaces()

  const { data: installRunbook, isLoading } = useQuery({
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
  const steps =
    latestConfig?.steps
      ?.slice()
      .sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)) ?? []
  const runs = installRunbook?.runs ?? []
  const basePath = `/${org?.id}/installs/${install?.id}`

  const lastRun = runs[0]
  const lastRunStatus = (() => {
    if (!lastRun) return undefined
    const wfStatus =
      typeof lastRun.install_workflow?.status === 'object'
        ? (lastRun.install_workflow.status as { status?: string })?.status
        : lastRun.install_workflow?.status
    return wfStatus ?? lastRun.status
  })()

  if (isLoading) {
    return (
      <PageSection flush className="flex-1">
        <PageTitle title={`Runbook | ${install?.name}`} />
        <Breadcrumbs
          breadcrumbs={[
            { path: `/${org?.id}`, text: org?.name },
            { path: `/${org?.id}/installs`, text: 'Installs' },
            {
              path: `/${org?.id}/installs/${install?.id}`,
              text: install?.name,
            },
            {
              path: `/${org?.id}/installs/${install?.id}/runbooks`,
              text: 'Runbooks',
            },
            {
              path: `/${org?.id}/installs/${install?.id}/runbooks/${runbookId}`,
              text: undefined,
            },
          ]}
        />
        <div className="@container flex flex-col flex-1">
          <header className="p-6 border-b flex flex-col gap-6">
            <div className="flex flex-wrap items-start gap-4 justify-between w-full">
              <HeadingGroup>
                <BackLink className="mb-4" />
                <Skeleton height="28px" width="200px" />
                <span className="flex items-center gap-4 mt-1">
                  <Skeleton height="20px" width="240px" />
                </span>
              </HeadingGroup>
              <Skeleton height="36px" width="100px" />
            </div>
            <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
              <Skeleton height="40px" width="120px" />
              <Skeleton height="40px" width="120px" />
            </div>
          </header>

          <div className="grid grid-cols-1 @5xl:grid-cols-12 flex-1">
            <div className="@5xl:col-span-8 flex flex-col gap-6">
              <PageSection className="flex flex-col gap-4">
                <Skeleton height="20px" width="60px" />
                <Skeleton height="120px" width="100%" />
              </PageSection>
            </div>
            <PageSection className="hidden @5xl:flex flex-col @5xl:col-span-4 gap-4">
              <Skeleton height="20px" width="100px" />
              <TimelineSkeleton eventCount={3} />
            </PageSection>
          </div>
        </div>
      </PageSection>
    )
  }

  return (
    <PageSection flush className="flex-1">
      <PageTitle
        title={`${runbook?.name ?? 'Runbook'} | ${install?.name}`}
      />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          {
            path: `/${org?.id}/installs/${install?.id}`,
            text: install?.name,
          },
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

            <div className="flex items-center gap-4">
              <div className="@5xl:hidden">
                <Button
                  variant="secondary"
                  onClick={() =>
                    addPanel(
                      <Panel heading="Run history">
                        <RunbookRunTimeline
                          runs={runs}
                          runbookName={runbook?.name ?? ''}
                          basePath={basePath}
                        />
                      </Panel>
                    )
                  }
                >
                  <Icon variant="ClockCounterClockwiseIcon" size={16} />
                  Run history
                </Button>
              </div>
              {installRunbook ? (
                <RunRunbookButton
                  installRunbook={installRunbook}
                  variant="primary"
                />
              ) : null}
            </div>
          </div>

          <div className="flex flex-wrap gap-x-8 gap-y-4 items-start">
            {lastRun ? (
              <LabeledStatus
                label="Last status"
                statusProps={{ status: lastRunStatus }}
                tooltipProps={{ tipContent: lastRunStatus }}
              />
            ) : null}
            <LabeledValue label="Steps">
              <Text variant="subtext">{steps.length}</Text>
            </LabeledValue>
            {lastRun ? (
              <LabeledValue label="Last run">
                <Time
                  variant="subtext"
                  time={lastRun.created_at}
                  format="relative"
                  shouldTick
                />
              </LabeledValue>
            ) : null}
          </div>
        </header>

        <div className="grid grid-cols-1 @5xl:grid-cols-12 flex-1">
          <div className="@5xl:col-span-8 flex flex-col gap-6">
            {latestConfig?.readme ? (
              <PageSection className="flex flex-col gap-4">
                <Text variant="base" weight="strong">
                  Readme
                </Text>
                <Markdown content={latestConfig.readme} mode="install" />
              </PageSection>
            ) : null}

            <PageSection className="flex flex-col gap-4">
              <Text variant="base" weight="strong">
                Steps
              </Text>
              {steps.length ? (
                <div className="grid grid-cols-1 gap-4">
                  {steps.map((step, i) => (
                    <RunbookStep key={step.id ?? i} index={i} step={step} actionBasePath={basePath} />
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
            <RunbookRunTimeline
              runs={runs}
              runbookName={runbook?.name ?? ''}
              basePath={basePath}
            />
          </PageSection>
        </div>
      </div>
    </PageSection>
  )
}
