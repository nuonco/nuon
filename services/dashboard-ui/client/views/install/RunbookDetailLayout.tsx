import { Navigate, Outlet, useLocation, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { RunRunbookButton } from '@/components/runbooks/RunRunbook'
import { RemovedFromAppConfigBanner } from '@/components/installs/RemovedFromAppConfig'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallRunbook, getInstallRunbooks } from '@/lib'

export const RunbookDetailLayout = () => {
  const { runbookId } = useParams()
  const { pathname } = useLocation()
  const { org } = useOrg()
  const { install, labelColors } = useInstall()

  const { data: installRunbook, isLoading } = useQuery({
    placeholderData: keepPreviousData,
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
    placeholderData: keepPreviousData,
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
  const basePath = `/${org?.id}/installs/${install?.id}/runbooks/${runbookId}`

  const isIndexRoute = pathname === basePath || pathname === `${basePath}/`

  if (!isLoading && isIndexRoute) {
    return <Navigate to={`${basePath}/readme`} replace />
  }

  const labelKeys = Object.keys(runbook?.labels ?? {}).sort()

  return (
    <>
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

      <DetailPage
        header={
          <DetailHeader
            title={runbook?.name}
            loading={isLoading}
            loadingWidth={20}
            description={runbook?.description}
            id={runbookId}
            identity={
              labelKeys.length ? (
                <span className="flex flex-wrap gap-1">
                  {labelKeys.map((k) => (
                    <LabelBadge
                      key={k}
                      labelKey={k}
                      labelValue={runbook?.labels?.[k]}
                      size="sm"
                      customColor={labelColors?.[k]}
                    />
                  ))}
                </span>
              ) : null
            }
            actions={
              installRunbook ? (
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
              ) : null
            }
          />
        }
        banners={removed ? <RemovedFromAppConfigBanner kind="runbook" /> : null}
        tabNav={{
          basePath,
          tabs: [
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
          ],
        }}
      >
        {isLoading ? (
          <div className="flex flex-col gap-3">
            <Text variant="body" loading loadingWidth={40} />
            <Text variant="body" loading loadingWidth={60} />
            <Text variant="body" loading loadingWidth={30} />
          </div>
        ) : (
          <Outlet context={{ installRunbook }} />
        )}
      </DetailPage>
    </>
  )
}
