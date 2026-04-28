import { useMutation } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { AdminConfirmationModal } from './AdminConfirmationModal'
import type { IModal } from '@/components/surfaces/Modal'

interface IAdminConfirmationModalContainer extends IModal {
  title: string
  message: string
  action: () => Promise<any>
  onConfirm?: () => void
  variant?: 'default' | 'warning' | 'danger'
  requiresInput?: boolean
  inputText?: string
  successMessage?: string
  errorMessage?: string
}

export const AdminConfirmationModalContainer = ({
  action,
  onConfirm,
  title,
  successMessage,
  errorMessage,
  ...props
}: IAdminConfirmationModalContainer) => {
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate: execute, isPending } = useMutation({
    mutationFn: action,
    onSuccess: () => {
      addToast(
        <Toast heading="Action complete" theme="success">
          <Text>{successMessage ?? `${title} completed successfully`}</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: any) => {
      const message = err?.error || err?.description || err?.message || 'Unknown error'
      addToast(
        <Toast heading="Action failed" theme="error">
          <Text>{errorMessage ?? `Failed to ${title.toLowerCase()}: ${message}`}</Text>
        </Toast>
      )
    },
  })

  const handleConfirm = () => {
    onConfirm?.()
    execute()
  }

  const handleCancel = () => {
    removeModal(props.modalId)
  }

  return (
    <AdminConfirmationModal
      {...props}
      title={title}
      onConfirm={handleConfirm}
      onCancel={handleCancel}
      isPending={isPending}
    />
  )
}
