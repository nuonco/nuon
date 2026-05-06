import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { updateCurrentOrgWebhook } from '@/lib'
import type { TAPIError, TWebhook } from '@/types'
import { EditWebhookModal, type EditWebhookFormInput } from './EditWebhook'

const EditWebhookModalContainer = ({
  webhook,
  ...props
}: { webhook: TWebhook } & Record<string, any>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: (input: EditWebhookFormInput) =>
      updateCurrentOrgWebhook({
        body: {
          ...(input.webhookSecret !== undefined
            ? { webhook_secret: input.webhookSecret }
            : {}),
          interests: input.interests,
        },
        orgId: org.id,
        webhookId: webhook.id ?? '',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks', org.id] })
      addToast(
        <Toast heading="Webhook updated" theme="success">
          <Text>Filter and secret changes are live.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to update webhook" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <EditWebhookModal
      webhook={webhook}
      isPending={isPending}
      error={error}
      onSubmit={(input) => mutate(input)}
      {...props}
    />
  )
}

export const EditWebhookButton = ({
  webhook,
  ...props
}: { webhook: TWebhook } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <EditWebhookModalContainer webhook={webhook} />

  return (
    <Button variant="ghost" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PencilSimple" />
      Edit
    </Button>
  )
}
