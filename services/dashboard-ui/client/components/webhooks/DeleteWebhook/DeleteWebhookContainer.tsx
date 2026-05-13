import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deleteCurrentOrgWebhook } from '@/lib'
import type { TAPIError, TWebhook } from '@/types'
import { DeleteWebhookModal } from './DeleteWebhook'

const DeleteWebhookModalContainer = ({
  webhook,
  ...props
}: { webhook: TWebhook } & Record<string, any>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: () =>
      deleteCurrentOrgWebhook({
        orgId: org.id,
        webhookId: webhook.id ?? '',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks', org.id] })
      addToast(
        <Toast heading="Webhook deleted" theme="success">
          <Text>The webhook will no longer receive events.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Unable to delete webhook" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <DeleteWebhookModal
      webhookUrl={webhook.webhook_url ?? ''}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const DeleteWebhookButton = ({
  webhook,
  ...props
}: { webhook: TWebhook } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <DeleteWebhookModalContainer webhook={webhook} />

  return (
    <Button
      variant="ghost"
      className="!text-red-800 dark:!text-red-500"
      onClick={() => addModal(modal)}
      {...props}
    >
      <Icon variant="TrashIcon" />
      Delete
    </Button>
  )
}
