import { Outlet, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Text } from '@/components/common/Text'
import { DetailHeader } from '@/components/layout/DetailHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { getRunbook } from '@/lib'

export const RunbookDetailLayout = () => {
  const { runbookId, branchId } = useParams()
  const { org } = useOrg()
  const { app, labelColors } = useApp()

  const { data: runbook, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['runbook', org?.id, app?.id, runbookId],
    queryFn: () =>
      getRunbook({ orgId: org!.id, appId: app!.id, runbookId: runbookId! }),
    enabled: !!org?.id && !!app?.id && !!runbookId,
  })

  const latestConfig = runbook?.configs?.[0]
  const steps =
    latestConfig?.steps?.slice().sort((a, b) => (a.idx ?? 0) - (b.idx ?? 0)) ??
    []

  const appBase = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}`
    : `/${org?.id}/apps/${app?.id}`
  const basePath = `${appBase}/runbooks/${runbookId}`

  const labelKeys = Object.keys(runbook?.labels ?? {}).sort()

  return (
    <>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          {
            path: `${appBase}/runbooks`,
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
          />
        }
        tabNav={{
          basePath,
          tabs: [
            { path: '/', text: 'Readme' },
            {
              path: '/steps',
              text: (
                <>
                  Steps <Badge size="sm">{steps.length}</Badge>
                </>
              ),
            },
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
          <Outlet context={{ runbook }} />
        )}
      </DetailPage>
    </>
  )
}
