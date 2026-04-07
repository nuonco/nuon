import { useSurfaces } from '@/hooks/use-surfaces'
import { AdminRunnersPanel } from '../AdminRunnersPanel'
import { AdminRunnersCard } from './AdminRunnersCard'

interface IAdminRunnersCardContainer {
  orgId: string
}

export const AdminRunnersCardContainer = ({ orgId }: IAdminRunnersCardContainer) => {
  const { addPanel } = useSurfaces()

  const handleOpenPanel = () => {
    const panel = <AdminRunnersPanel orgId={orgId} />
    addPanel(panel)
  }

  return <AdminRunnersCard onOpenPanel={handleOpenPanel} />
}
