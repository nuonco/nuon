import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import type { IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { createApp } from '@/lib'
import type { TAPIError } from '@/types'
import { CreateAppModal } from './CreateAppModal'

type ICreateAppModalContainer = Omit<IModal, 'onSubmit'>

export const CreateAppModalContainer = ({
  ...props
}: ICreateAppModalContainer) => {
  const navigate = useNavigate()
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()
  const queryClient = useQueryClient()

  const { mutate, isPending: isSubmitting } = useMutation({
    mutationFn: (body: { name: string }) =>
      createApp({ orgId: org.id, body }),
    onSuccess: (app) => {
      addToast(
        <Toast heading="App created" theme="success">
          <Text>Created app {app.name}.</Text>
        </Toast>
      )
      queryClient.invalidateQueries({ queryKey: ['apps'] })
      removeModal(props.modalId)
      navigate(`/${org.id}/apps/${app.id}/branches`)
    },
    onError: (error: TAPIError) => {
      addToast(
        <Toast heading="App creation failed" theme="error">
          <Text>{error.error || 'Failed to create app.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateAppModal
      isSubmitting={isSubmitting}
      onSubmit={(body) => mutate(body)}
      onCancel={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const CreateAppButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <CreateAppModalContainer />
  return (
    <Button variant="primary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PlusIcon" size={16} />
      Create app
    </Button>
  )
}
