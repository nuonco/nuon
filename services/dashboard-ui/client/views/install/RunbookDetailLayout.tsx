import { Navigate, Outlet, useLocation, useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/common/Badge'
import { LabelBadge } from '@/components/common/LabelBadge'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { RunRunbookButton } from '@/components/runbooks/RunRunbook'
import { RemovedFromAppConfigBanner } from '@/components/installs/RemovedFromAppConfig'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { TabNav } from '@/components/navigation/TabNav'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallRunbook, getInstallRunbooks } from '@/lib'

export const RunbookDetailLayout = () => {
  const { runbookId } = useParams()
  const { pathname } = useLocation()
  const { org } = useOrg()
  const { install, labelColors } = useInstall()

  const { data: installRunbook, isLoading } = useQuery({
    queryKey: ['install-runbook', org?.id, install?.id, runbookId],
    queryFn: () =>
      getInstallRunbook({
        orgId: org!.id,
        installId: install!.id,
        runbookId: runbookId!,
      }),
    enabled: !!org?.id && !!install?.id && !!runbookId,
    refetchInterval: 10000,
  })

  const { data: removedResult } = useQuery({
    queryKey: ['install-runbooks-removed', org?.id, install?.id],
    queryFn: () =>
      getInstallRunbooks({
        orgId: org!.id,
        installId: install!.id,
        limit: 100,
        offset: 0,
        synced: false,
      }),
    enabled: !!org?.id && !!install?.id,
  })
  const removed = (removedResult?.data ?? []).some(
    (r) => r.runbook_id === runbookId || r.id === runbookId
  )

  const runbook = installRunbook?.runbook
  const latestConfig = runbook?.configs?.[0]
  const steps =
    latestConfig?.steps?.slice().sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)) ??
    []
  const runs = installRunbook?.runs ?? []
  const basePath = `/${org?.id}/installs/${install?.id}/runbooks/${runbookId}`

  const isIndexRoute = pathname === basePath || pathname === `${basePath}/`

  if (!isLoading && isIndexRoute) {
    return <Navigate to={`${basePath}/readme`} replace />
  }

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
              path: basePath,
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
          </header>
          <PageSection className="flex flex-col gap-4">
            <Skeleton height="40px" width="300px" />
            <Skeleton height="200px" width="100%" />
          </PageSection>
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
            path: basePath,
            text: runbook?.name,
          },
        ]}
      />

      <div className="@container flex flex-col flex-1">
        <header className="p-6 border-b flex flex-col gap-6">
          <div className="flex flex-wrap items-start gap-4 justify-between w-full">
            <HeadingGroup>
              <BackLink className="mb-4" />
              <span className="flex flex-wrap items-center gap-3">
                <Text variant="h3" weight="strong">
                  {runbook?.name}
                </Text>
                {runbook?.labels && Object.keys(runbook.labels).length > 0 ? (
                  <span className="flex flex-wrap items-center gap-1">
                    {Object.keys(runbook.labels)
                      .sort()
                      .map((k) => (
                        <LabelBadge key={k} labelKey={k} labelValue={runbook.labels[k]} size="sm" customColor={labelColors?.[k]} />
                      ))}
                  </span>
                ) : null}
              </span>
              {runbook?.description ? (
                <Text variant="subtext">
                  {runbook.description}
                </Text>
              ) : null}
              {runbookId ? <ID>{runbookId}</ID> : null}
            </HeadingGroup>

            {installRunbook ? (
              removed ? (
                <Button
                  variant="primary"
                  disabled
                  tooltipProps={{
                    position: 'left',
                    tipContent:
                      "This runbook is no longer in the install's app config version.",
                  }}
                >
                  Run runbook
                  <Icon variant="PlayIcon" />
                </Button>
              ) : (
                <RunRunbookButton
                  installRunbook={installRunbook}
                  variant="primary"
                />
              )
            ) : null}
          </div>
        </header>

        {removed ? (
          <div className="px-6 pt-6">
            <RemovedFromAppConfigBanner kind="runbook" />
          </div>
        ) : null}

        <PageSection>
          <TabNav
            basePath={basePath}
            tabs={[
              { path: '/readme', text: 'Readme' },
              {
                path: '/steps',
                text: (
                  <>
                    Steps <Badge size="sm">{steps.length}</Badge>
                  </>
                ),
              },
              { path: '/history', text: 'Run history' },
            ]}
          />
          <Outlet context={{ installRunbook }} />
        </PageSection>
      </div>
    </PageSection>
  )
}
