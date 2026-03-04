import { Suspense } from 'react'
import { InstallGroupsCanvas } from './install-groups-canvas'

interface IInstallGroupsPage {
  params: Promise<{
    'org-id': string
    'app-id': string
    'branch-id': string
  }>
}

export default async function InstallGroupsPage({ params }: IInstallGroupsPage) {
  const resolvedParams = await params
  const orgId = resolvedParams['org-id']
  const appId = resolvedParams['app-id']
  const branchId = resolvedParams['branch-id']

  return (
    <div className="p-6">
      <Suspense fallback={<div>Loading install groups...</div>}>
        <InstallGroupsCanvas
          appId={appId}
          branchId={branchId}
          orgId={orgId}
        />
      </Suspense>
    </div>
  )
}
