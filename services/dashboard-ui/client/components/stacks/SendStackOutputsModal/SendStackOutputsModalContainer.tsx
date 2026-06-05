import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Text } from '@/components/common/Text'
import { postPhoneHome } from '@/lib'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { SendStackOutputsModal } from './SendStackOutputsModal'
import type { IModal } from '@/components/surfaces/Modal'
import type { TAPIError } from '@/types'

interface ISendStackOutputsModalContainer extends IModal {
  phoneHomeId: string
  versionId: string
}

export const SendStackOutputsModalContainer = ({
  phoneHomeId,
  versionId,
  ...props
}: ISendStackOutputsModalContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate, isPending, error } = useMutation({
    mutationFn: (body: Record<string, unknown>) =>
      postPhoneHome({
        installId: install!.id,
        orgId: org!.id,
        phoneHomeId,
        body,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['install-stack'] })
      removeModal(props.modalId)
      addToast(
        <Toast heading="Phone home triggered" theme="info">
          <Text>Stack outputs sent for processing.</Text>
        </Toast>
      )
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Phone home failed" theme="error">
          <Text>{err?.error ?? 'Unable to send stack outputs.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <SendStackOutputsModal
      phoneHomeId={phoneHomeId}
      versionId={versionId}
      onSend={(body) => mutate(body)}
      isPending={isPending}
      error={error as TAPIError | undefined}
      {...props}
    />
  )
}
