import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useRoleOptions } from '@/hooks/use-roles'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createStaticToken } from '@/lib'
import { CreateApiTokenModal } from './CreateApiToken'

const CreateApiTokenModalContainer = (props: Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const { roleOptions, isLoading } = useRoleOptions('api_token')
  const [createdToken, setCreatedToken] = useState<string | null>(null)

  const { mutate, isPending, error } = useMutation({
    mutationFn: ({
      name,
      duration,
      role,
      identity,
    }: {
      name: string
      duration: string
      role: string
      identity: 'personal' | 'service_account'
    }) =>
      createStaticToken({
        body:
          identity === 'personal'
            ? { name, duration, token_identity: 'personal' }
            : { name, duration, role },
        orgId: org.id,
      }),
    onSuccess: (data, { name }) => {
      queryClient.invalidateQueries({ queryKey: ['static-tokens', org?.id] })
      setCreatedToken(data?.api_token ?? null)
      addToast(
        <Toast heading="Token created" theme="success">
          <Text>API token {name} was created for {org.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <CreateApiTokenModal
      isPending={isPending}
      error={error}
      createdToken={createdToken}
      roleOptions={roleOptions}
      rolesLoading={isLoading}
      onSubmit={({ name, duration, role, identity }) =>
        mutate({ name, duration, role, identity })
      }
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
