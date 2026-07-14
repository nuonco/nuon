import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { updateApp } from '@/lib'
import type { TAPIError } from '@/types/dashboard.types'

interface IResetLabelColorsModal extends IModal {
  orgId: string
  appId: string
  appName?: string
}

export const ResetLabelColorsModal = ({ orgId, appId, appName, ...props }: IResetLabelColorsModal) => {
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate: reset, isPending } = useMutation({
    mutationFn: () => updateApp({ orgId, appId, body: { label_colors: {} } }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['app-labels', orgId, appId] })
      queryClient.invalidateQueries({ queryKey: ['app', orgId, appId] })
      addToast(
        <Toast heading="Label colors reset" theme="success">
          <Text>Restored the automatic colors{appName ? ` for ${appName}` : ''}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Label update failed" theme="error">
          <Text>{err?.error || 'Unable to reset label colors.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading="Reset label colors?"
      primaryActionTrigger={{
        children: isPending ? 'Resetting...' : 'Reset to defaults',
        disabled: isPending,
        onClick: () => reset(),
        variant: 'danger',
      }}
      {...props}
    >
      <Text>
        This clears every color override{appName ? ` for ${appName}` : ''} and restores the automatic
        default colors.
      </Text>
    </Modal>
  )
}
