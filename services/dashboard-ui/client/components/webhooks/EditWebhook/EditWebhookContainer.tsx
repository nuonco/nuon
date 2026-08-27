import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { updateCurrentOrgWebhook } from '@/lib'
import type { TWebhook } from '@/types'
import { WebhookFormModal } from '@/components/webhooks/WebhookForm'
import type { WebhookFormOutput } from '@/components/webhooks/WebhookForm/schema'

const EditWebhookModalContainer = ({
  webhook,
  ...props
}: { webhook: TWebhook } & Record<string, any>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: (input: WebhookFormOutput) =>
      updateCurrentOrgWebhook({
        body: {
          ...(input.webhookSecret !== undefined
            ? { webhook_secret: input.webhookSecret }
            : {}),
          match: input.match ?? null,
          interests: input.interests,
        },
        orgId: org.id,
        webhookId: webhook.id ?? '',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks', org.id] })
      addToast(
        <Toast heading="Webhook updated" theme="success">
          <Text>Filter, scope, and secret changes are live.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  const friendlyError =
    error && error.status === 409
      ? {
          ...error,
          error:
            'Another webhook for this URL already uses this scope. Pick a different scope.',
        }
      : error

  return (
    <WebhookFormModal
      mode="edit"
      webhook={webhook}
      isPending={isPending}
      error={friendlyError}
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
      Edit webhook
      <Icon variant="PencilSimpleIcon" />
    </Button>
  )
}
