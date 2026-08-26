import { Outlet, useParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { CompositeError } from '@/components/common/CompositeError'
import { Text } from '@/components/common/Text'
import { BuildHeader } from '@/components/builds/BuildHeader'
import { DetailPage } from '@/components/layout/DetailPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { useApp } from '@/hooks/use-app'
import { useBuild } from '@/hooks/use-build'
import { useOrg } from '@/hooks/use-org'
import { getComponent } from '@/lib'
import { BuildProvider } from '@/providers/build-provider'
import type { TComponent } from '@/types'

const BuildLayoutInner = ({
  component,
}: {
  component: TComponent | undefined
}) => {
  const { branchId, componentId, buildId } = useParams()
  const { build } = useBuild()
  const { org } = useOrg()
  const { app } = useApp()

  if (!build) return null

  const basePath = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}/components/${componentId}/builds/${buildId}`
    : `/${org?.id}/apps/${app?.id}/components/${componentId}/builds/${buildId}`

  return (
    <DetailPage
      header={<BuildHeader component={component as TComponent} />}
      banners={
        <>
          {build?.composite_error ? (
            <CompositeError error={build.composite_error} />
          ) : null}
          {build?.status_v2?.metadata?.duplicate_build ? (
            <Banner theme="warn">
              <div className="flex flex-col">
                <Text weight="strong" variant="base">
                  Duplicate build
                </Text>
                <Text theme="neutral">
                  This build was triggered against the same commit and config as
                  a previous build. Push new changes to your branch to create a
                  unique build.
                </Text>
              </div>
            </Banner>
          ) : null}
        </>
      }
      tabNav={{
        basePath,
        tabs: [
          { path: '/', text: 'Summary' },
          { path: '/logs', text: 'Logs' },
          ...(org?.features?.['trace-view']
            ? [{ path: '/trace', text: 'Trace' }]
            : []),
        ],
      }}
    >
      <Outlet />
    </DetailPage>
  )
}

export const BuildLayout = () => {
  const { branchId, componentId, buildId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()

  const appBase = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}`
    : `/${org?.id}/apps/${app?.id}`

  const { data: component } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['component', org?.id, app?.id, componentId],
    queryFn: () => getComponent({ orgId: org.id, componentId: componentId! }),
    enabled: !!org?.id && !!app?.id && !!componentId,
  })

  return (
    <>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          {
            path: `${appBase}/components`,
            text: 'Components',
          },
          {
            path: `${appBase}/components/${componentId}`,
            text: component?.name,
          },
          {
            path: `${appBase}/components/${componentId}/builds/${buildId}`,
            text: 'Build',
          },
        ]}
      />
      <BuildProvider
        buildId={buildId!}
        componentId={componentId!}
        componentName={component?.name}
        shouldPoll
      >
        <BuildLayoutInner component={component} />
      </BuildProvider>
    </>
  )
}
