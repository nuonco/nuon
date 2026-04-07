import { useSurfaces } from '@/hooks/use-surfaces'
import { AdminOrgFeaturesPanel } from '../AdminOrgFeaturesPanel'
import { AdminFeatureToggleCard } from './AdminFeatureToggleCard'
import type { TOrg } from '@/types'

interface IAdminFeatureToggleCardContainer {
  org: TOrg
  orgId: string
}

export const AdminFeatureToggleCardContainer = ({ org, orgId }: IAdminFeatureToggleCardContainer) => {
  const { addPanel } = useSurfaces()

  const handleOpenPanel = () => {
    const panel = <AdminOrgFeaturesPanel org={org} orgId={orgId} />
    addPanel(panel)
  }

  return <AdminFeatureToggleCard onOpenPanel={handleOpenPanel} />
}
