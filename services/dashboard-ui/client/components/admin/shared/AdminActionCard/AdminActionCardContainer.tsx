import { useMutation } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { AdminConfirmationModal } from '../AdminConfirmationModal'
import { AdminActionCard } from './AdminActionCard'

export interface AdminActionCardContainerProps {
  title: string
  description: string
  action: () => Promise<any>
  variant?: 'default' | 'warning' | 'danger'
  requiresConfirmation?: boolean
  confirmationText?: string
  requiresInput?: boolean
  inputText?: string
}

export const AdminActionCardContainer = ({
  title,
  description,
  action,
  variant = 'default',
  requiresConfirmation = false,
  confirmationText,
  requiresInput = false,
  inputText = 'CONFIRM',
}: AdminActionCardContainerProps) => {
  const { addModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate: execute, isPending: isLoading } = useMutation({
    mutationFn: action,
    onSuccess: () => {
      addToast(
        <Toast heading="Action Complete" theme="success">
          <Text>{title} completed successfully</Text>
        </Toast>
      )
    },
    onError: (err: any) => {
      const message = err?.error || err?.description || err?.message || 'Unknown error'
      addToast(
        <Toast heading="Action Failed" theme="error">
          <Text>Failed to {title.toLowerCase()}: {message}</Text>
        </Toast>
      )
    },
  })

  const handleClick = () => {
    if (requiresConfirmation) {
      const modalId = `admin-confirm-${title.toLowerCase().replace(/\s+/g, '-')}-${Date.now()}`
      const confirmationModal = (
        <AdminConfirmationModal
          modalId={modalId}
          title={`Confirm: ${title}`}
          message={confirmationText || `Are you sure you want to ${title.toLowerCase()}?`}
          onConfirm={() => execute()}
          variant={variant}
          requiresInput={requiresInput}
          inputText={inputText}
        />
      )
      addModal(confirmationModal)
    } else {
      execute()
    }
  }

  return (
    <AdminActionCard
      title={title}
      description={description}
      variant={variant}
      isLoading={isLoading}
      onClick={handleClick}
    />
  )
}
