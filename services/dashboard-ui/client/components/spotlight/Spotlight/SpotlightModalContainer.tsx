import { useNavigate } from 'react-router'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { SpotlightModal } from './SpotlightModal'
import type { IModal } from '@/components/surfaces/Modal'

export const SpotlightModalContainer = ({ ...props }: IModal) => {
  const { removeModal, addModal } = useSurfaces()
  const { org } = useOrg()
  const navigate = useNavigate()
  const orgId = org?.id ?? ''

  return (
    <SpotlightModal
      orgId={orgId}
      onClose={() => removeModal(props.modalId)}
      onNavigate={(path) => navigate(path)}
      onAddModal={addModal}
      orgFeatures={org?.features as Record<string, boolean> | undefined}
      {...props}
    />
  )
}
