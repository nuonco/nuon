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
import { createServiceAccount, createToken } from '@/lib'
import { CreateApiTokenModal } from './CreateApiToken'

const CreateApiTokenModalContainer = (props: Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const { roleOptions, isLoading } = useRoleOptions('api_token')
  const [createdToken, setCreatedToken] = useState<string | null>(null)
  const [accountCreatedError, setAccountCreatedError] = useState<string | null>(
    null
  )

  const { mutate, isPending, error } = useMutation({
    mutationFn: async ({
      name,
      duration,
      role,
    }: {
      name: string
      duration: string
      role: string
    }) => {
      const account = await createServiceAccount({
        body: { name, role },
        orgId: org.id,
      })
      try {
        return await createToken({
          body: { account_id: account.id!, name, duration },
          orgId: org.id,
        })
      } catch (err) {
        setAccountCreatedError(
          `${name} was created as a service account, but creating a token for it failed.`
        )
        throw err
      }
    },
    onSuccess: (data, { name }) => {
      queryClient.invalidateQueries({ queryKey: ['tokens', org?.id] })
      queryClient.invalidateQueries({ queryKey: ['service-accounts', org?.id] })
      setCreatedToken(data?.token ?? null)
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
      error={accountCreatedError ? null : error}
      accountCreatedError={accountCreatedError}
      createdToken={createdToken}
      roleOptions={roleOptions}
      rolesLoading={isLoading}
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
