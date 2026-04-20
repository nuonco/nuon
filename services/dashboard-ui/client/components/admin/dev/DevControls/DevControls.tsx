import { DevOrgSection } from '../sections/DevOrgSection'
import { DevInstallSection } from '../sections/DevInstallSection'

interface IDevControls {
  orgId: string
  installId?: string
}

export const DevControls = ({ orgId, installId }: IDevControls) => {
  return (
    <div className="flex flex-col gap-8 h-full overflow-y-auto">
      <DevOrgSection orgId={orgId} />
      <DevInstallSection installId={installId} orgId={orgId} />
    </div>
  )
}
