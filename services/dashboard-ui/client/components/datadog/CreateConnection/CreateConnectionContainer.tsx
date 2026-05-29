import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { createDatadogConnection } from '@/lib'
import type { TAPIError } from '@/types'
import {
  CreateConnectionModal,
  type CreateConnectionInput,
} from './CreateConnection'

const CreateConnectionModalContainer = (props: Record<string, any>) => {
  const { org } = useOrg()
  const queryClient = useQueryClient()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: (input: CreateConnectionInput) =>
      createDatadogConnection({ orgId: org.id, body: input }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['datadog-connections', org.id],
      })
      addToast(
        <Toast heading="Datadog connection created" theme="success">
          <Text>
            Subscribe an event scope below to start streaming events into
            this tenant.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (err: TAPIError) => {
      // Create surfaces 400 when DD rejects the API key — keep the
      // dashboard toast in lockstep with the validator wording so users
      // see the same phrasing here and on the connection's Test action.
      addToast(
        <Toast heading="Unable to create connection" theme="error">
          <Text>{err?.description || err?.error || 'Please try again.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateConnectionModal
      isPending={isPending}
      error={error}
      onSubmit={mutate}
      {...props}
    />
  )
}

export const CreateConnectionButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <CreateConnectionModalContainer />

  return (
    <Button variant="primary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PlusIcon" />
      Connect Datadog
    </Button>
  )
}
