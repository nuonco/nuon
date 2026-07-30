import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { resetInstallHealthBaseline } from '@/lib'
import type { TAPIError } from '@/types'

interface IResetHealthWindowModal extends IModal {
  installId: string
}

export const ResetHealthWindowModal = ({
  installId,
  ...props
}: IResetHealthWindowModal) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate: reset, isPending } = useMutation({
    mutationFn: () =>
      resetInstallHealthBaseline({ orgId: org!.id, installId }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['install-health-timeline'] })
      queryClient.invalidateQueries({
        queryKey: ['install-component-health-timeline'],
      })
      addToast(
        <Toast heading="Health window reset" theme="success">
          <Text>Uptime and the health timeline now start from this moment.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Reset failed" theme="error">
          <Text>{err?.error || 'Unable to reset the health window.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading="Reset health window?"
      primaryActionTrigger={{
        children: isPending ? 'Resetting...' : 'Reset window',
        disabled: isPending,
        onClick: () => reset(),
        variant: 'danger',
      }}
      {...props}
    >
      <Text>
        Uptime and the health timeline will start counting from now — useful
        after initial provisioning or planned maintenance. Past observations
        stay recorded but no longer count toward uptime.
      </Text>
    </Modal>
  )
}

export const ResetHealthWindowButton = ({ installId }: { installId: string }) => {
  const { addModal } = useSurfaces()
  const modal = <ResetHealthWindowModal installId={installId} />

  return (
    <Button variant="ghost" size="xs" onClick={() => addModal(modal)}>
      <Icon variant="ClockCounterClockwiseIcon" size={14} />
      Reset window
    </Button>
  )
}
