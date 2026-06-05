import { useMutation } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { AdminConfirmationModal } from '@/components/admin/shared/AdminConfirmationModal'
import { DevActionCard } from './DevActionCard'

export interface DevActionCardContainerProps {
  title: string
  description: string
  action: () => Promise<any>
  variant?: 'default' | 'warning' | 'danger'
  requiresConfirmation?: boolean
  confirmationText?: string
  requiresInput?: boolean
  inputText?: string
}

export const DevActionCardContainer = ({
  title,
  description,
  action,
  variant = 'default',
  requiresConfirmation = false,
  confirmationText,
  requiresInput = false,
  inputText = 'CONFIRM',
}: DevActionCardContainerProps) => {
  const { addModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate: execute, isPending: isLoading } = useMutation({
    mutationFn: action,
    onSuccess: () => {
      addToast(
        <Toast heading="Action complete" theme="success">
          <Text>{title} completed.</Text>
        </Toast>
      )
    },
    onError: () => {
      addToast(
        <Toast heading="Action failed" theme="error">
          <Text>Unable to {title.toLowerCase()}. Try again.</Text>
        </Toast>
      )
    },
  })

  const handleClick = () => {
    if (requiresConfirmation) {
      const confirmationModal = (
        <AdminConfirmationModal
          title={`Confirm: ${title}`}
          message={
            confirmationText ||
            `This will ${title.toLowerCase()}.`
          }
          action={action}
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
    <DevActionCard
      title={title}
      description={description}
      variant={variant}
      isLoading={isLoading}
      onClick={handleClick}
    />
  )
}
