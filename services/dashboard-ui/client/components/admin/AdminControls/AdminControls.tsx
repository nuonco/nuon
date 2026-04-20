import { AdminOrgSection } from '../sections/AdminOrgSection'
import { AdminAppSection } from '../sections/AdminAppSection'
import { AdminInstallSection } from '../sections/AdminInstallSection'
import { AdminSandboxSection } from '../sections/AdminSandboxSection'
import { DevOrgSection } from '../dev/sections/DevOrgSection'
import { DevInstallSection } from '../dev/sections/DevInstallSection'

interface IAdminControls {
  isNuonEmployee: boolean
  isDev: boolean
  orgId: string
  appId?: string
  installId?: string
}

export const AdminControls = ({
  isNuonEmployee,
  isDev,
  orgId,
  appId,
  installId,
}: IAdminControls) => {
  if (!isNuonEmployee && !isDev) {
    return null
  }

  return (
    <div className="flex flex-col gap-8 h-full overflow-y-auto">
      {isNuonEmployee && <AdminSandboxSection />}
      {isNuonEmployee && <AdminOrgSection orgId={orgId} />}
      {isNuonEmployee && appId && <AdminAppSection orgId={orgId} appId={appId} />}
      {isNuonEmployee && installId && (
        <AdminInstallSection orgId={orgId} installId={installId} />
      )}
      {isDev && <DevOrgSection orgId={orgId} />}
      {isDev && <DevInstallSection orgId={orgId} installId={installId} />}
    </div>
  )
}
