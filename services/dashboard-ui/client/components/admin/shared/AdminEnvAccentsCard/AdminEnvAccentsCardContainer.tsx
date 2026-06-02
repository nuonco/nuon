import { useSurfaces } from '@/hooks/use-surfaces'
import { AdminEnvAccentsPanel } from '../AdminEnvAccentsPanel'
import { AdminEnvAccentsCard } from './AdminEnvAccentsCard'

interface IAdminEnvAccentsCardContainer {
  orgId: string
}

export const AdminEnvAccentsCardContainer = ({
  orgId,
}: IAdminEnvAccentsCardContainer) => {
  const { addPanel } = useSurfaces()

  const handleOpenPanel = () => {
    const panel = <AdminEnvAccentsPanel orgId={orgId} />
    addPanel(panel)
  }

  return <AdminEnvAccentsCard onOpenPanel={handleOpenPanel} />
}
