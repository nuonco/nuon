import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createCurrentOrgWebhook } from '@/lib'
import type { TAPIError } from '@/types'
import { CreateWebhookModal } from './CreateWebhook'

const CreateWebhookModalContainer = (props: Record<string, any>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: ({
      webhookUrl,
      webhookSecret,
    }: {
      webhookUrl: string
      webhookSecret: string
    }) =>
      createCurrentOrgWebhook({
        body: {
          webhook_url: webhookUrl,
          ...(webhookSecret ? { webhook_secret: webhookSecret } : {}),
        },
        orgId: org.id,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['webhooks', org.id] })
      addToast(
        <Toast heading="Webhook created" theme="success">
          <Text>The webhook will receive future workflow lifecycle events.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      const heading =
        err?.status === 409
          ? 'Webhook URL already exists'
          : 'Unable to create webhook'
      addToast(
        <Toast heading={heading} theme="error">
          <Text>
            {err?.description || err?.error || 'Please try again.'}
          </Text>
        </Toast>
      )
    },
  })

  const errorWithFriendlyConflict: TAPIError | null = error
    ? error.status === 409
      ? {
          ...error,
          error:
            'A webhook with this URL already exists for this org. Delete the existing webhook to recreate it.',
        }
      : error
    : null

  return (
    <CreateWebhookModal
      isPending={isPending}
      error={errorWithFriendlyConflict}
      onSubmit={({ webhookUrl, webhookSecret }) =>
        mutate({ webhookUrl, webhookSecret })
      }
      {...props}
    />
  )
}

export const CreateWebhookButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <CreateWebhookModalContainer />

  return (
    <Button variant="primary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="Plus" />
      Create webhook
    </Button>
  )
}
