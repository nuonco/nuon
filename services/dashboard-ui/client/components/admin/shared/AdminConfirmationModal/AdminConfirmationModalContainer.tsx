import { useSurfaces } from '@/hooks/use-surfaces'
import { AdminConfirmationModal } from './AdminConfirmationModal'
import type { IModal } from '@/components/surfaces/Modal'

interface IAdminConfirmationModalContainer extends IModal {
  title: string
  message: string
  onConfirm: () => void
  variant?: 'default' | 'warning' | 'danger'
  requiresInput?: boolean
  inputText?: string
}

export const AdminConfirmationModalContainer = ({
  ...props
}: IAdminConfirmationModalContainer) => {
  const { removeModal } = useSurfaces()

  const handleCancel = () => {
    removeModal(props.modalId)
  }

  return <AdminConfirmationModal {...props} onCancel={handleCancel} />
}
