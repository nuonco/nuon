import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createTrigger } from '@/lib'
import { CreateTriggerModal } from './CreateTrigger'

const CreateTriggerModalContainer = (props: Omit<IModal, 'onSubmit'>) => {
  const navigate = useNavigate()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const queryClient = useQueryClient()
  const mutation = useMutation({
    mutationFn: (body: Parameters<typeof createTrigger>[0]['body']) =>
      createTrigger({ body, orgId: org!.id }),
    onSuccess: (response) => {
      void queryClient.invalidateQueries({
        queryKey: ['triggers', org?.id],
      })
      removeModal(props.modalId)
      if (response?.trigger?.id)
        navigate(`/${org?.id}/triggers/${response.trigger.id}`)
    },
  })
  return (
    <CreateTriggerModal
      error={mutation.error}
      isPending={mutation.isPending}
      onSubmit={mutation.mutate}
      {...props}
    />
  )
}

export const CreateTriggerButton = (
  props: Omit<IButtonAsButton, 'children'>
) => {
  const { addModal } = useSurfaces()
  const modal = <CreateTriggerModalContainer />
  return (
    <Button variant="primary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PlusIcon" />
      Create trigger
    </Button>
  )
}
