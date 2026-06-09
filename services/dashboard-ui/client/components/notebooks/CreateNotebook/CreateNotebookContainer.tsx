import { useNavigate } from 'react-router'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import type { IModal } from '@/components/surfaces/Modal'
import { Toast } from '@/components/surfaces/Toast'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import { createNotebook, type ICreateNotebookBody } from '@/lib'
import type { TAPIError } from '@/types'
import { CreateNotebookModal } from './CreateNotebook'

const CreateNotebookModalContainer = ({ onSubmit: _, ...props }: IModal) => {
  const navigate = useNavigate()
  const { org } = useOrg()
  const { install } = useInstall()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate, isPending, error } = useMutation({
    mutationFn: (body: ICreateNotebookBody) =>
      createNotebook({
        orgId: org!.id,
        installId: install!.id,
        body,
      }),
    onSuccess: (nb) => {
      queryClient.invalidateQueries({
        queryKey: ['notebooks', org?.id, install?.id],
      })
      addToast(
        <Toast heading="Notebook created" theme="success">
          <Text>Created {nb.name}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
      navigate(`/${org?.id}/installs/${install?.id}/notebooks/${nb.id}`)
    },
    onError: (err: TAPIError) => {
      addToast(
        <Toast heading="Notebook creation failed" theme="error">
          <Text>{err?.error || 'Unable to create the notebook.'}</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateNotebookModal
      isPending={isPending}
      error={error}
      onSubmit={(body) => mutate(body)}
      {...props}
    />
  )
}

export const CreateNotebookButton = (
  props: Omit<IButtonAsButton, 'children'>
) => {
  const { addModal } = useSurfaces()
  const modal = <CreateNotebookModalContainer />

  return (
    <Button
      variant="primary"
      onClick={() => addModal(modal)}
      {...props}
    >
      <Icon variant="PlusIcon" size={16} />
      Create notebook
    </Button>
  )
}
