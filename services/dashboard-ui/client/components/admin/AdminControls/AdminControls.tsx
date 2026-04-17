import { AdminOrgSection } from '../sections/AdminOrgSection'
import { AdminAppSection } from '../sections/AdminAppSection'
import { AdminInstallSection } from '../sections/AdminInstallSection'
import { AdminSandboxSection } from '../sections/AdminSandboxSection'

interface IAdminControls {
  isNuonEmployee: boolean
  orgId: string
  appId?: string
  installId?: string
}

export const AdminControls = ({
  isNuonEmployee,
  orgId,
  appId,
  installId,
}: IAdminControls) => {
  if (!isNuonEmployee) {
    return null
  }

  return (
    <div className="flex flex-col gap-8 h-full overflow-y-auto">
      <AdminSandboxSection />
      <AdminOrgSection orgId={orgId} />
      {appId && <AdminAppSection orgId={orgId} appId={appId} />}
      {installId && <AdminInstallSection orgId={orgId} installId={installId} />}
    </div>
  )
}
