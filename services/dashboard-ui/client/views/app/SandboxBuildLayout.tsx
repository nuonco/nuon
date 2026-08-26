import { Outlet, useParams } from 'react-router'
import { CompositeError } from '@/components/common/CompositeError'
import { DetailPage } from '@/components/layout/DetailPage'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { SandboxBuildHeader } from '@/components/sandbox/builds/SandboxBuildHeader'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSandboxBuild } from '@/hooks/use-sandbox-build'
import { SandboxBuildProvider } from '@/providers/sandbox-build-provider'

const SandboxBuildLayoutInner = () => {
  const { branchId, buildId } = useParams()
  const { build } = useSandboxBuild()
  const { org } = useOrg()
  const { app } = useApp()

  if (!build) return null

  const basePath = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}/sandbox/builds/${buildId}`
    : `/${org?.id}/apps/${app?.id}/sandbox/builds/${buildId}`

  return (
    <DetailPage
      header={<SandboxBuildHeader />}
      banners={
        build?.composite_error ? (
          <CompositeError error={build.composite_error} />
        ) : null
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

export const SandboxBuildLayout = () => {
  const { branchId, buildId } = useParams()
  const { org } = useOrg()
  const { app } = useApp()

  const appBase = branchId
    ? `/${org?.id}/apps/${app?.id}/branches/${branchId}`
    : `/${org?.id}/apps/${app?.id}`

  return (
    <>
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `${appBase}/sandbox`, text: 'Sandbox' },
          {
            path: `${appBase}/sandbox/builds/${buildId}`,
            text: 'Build',
          },
        ]}
      />
      <SandboxBuildProvider buildId={buildId!} shouldPoll>
        <SandboxBuildLayoutInner />
      </SandboxBuildProvider>
    </>
  )
}
