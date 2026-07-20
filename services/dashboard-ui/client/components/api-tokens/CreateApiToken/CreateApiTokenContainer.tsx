import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createStaticToken } from '@/lib'
import { CreateApiTokenModal } from './CreateApiToken'

const CreateApiTokenModalContainer = (props: Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const [createdToken, setCreatedToken] = useState<string | null>(null)

  const { mutate, isPending, error } = useMutation({
    mutationFn: ({
      name,
      duration,
      role,
    }: {
      name: string
      duration: string
      role: string
    }) => createStaticToken({ body: { name, duration, role }, orgId: org.id }),
    onSuccess: (data, { name }) => {
      queryClient.invalidateQueries({ queryKey: ['static-tokens', org?.id] })
      setCreatedToken(data?.api_token ?? null)
      addToast(
        <Toast heading="Token created" theme="success">
          <Text>API token {name} was created for {org.name}.</Text>
        </Toast>
      )
    },
    onError: (_err, { name }) => {
      addToast(
        <Toast heading="Create failed" theme="error">
          <Text>There was an error creating API token {name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateApiTokenModal
      isPending={isPending}
      error={error}
      createdToken={createdToken}
      onSubmit={({ name, duration, role }) => mutate({ name, duration, role })}
      onDone={() => removeModal(props.modalId)}
      {...props}
    />
  )
}

export const CreateApiTokenButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <CreateApiTokenModalContainer />

  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PlusIcon" />
      Create token
    </Button>
  )
}
