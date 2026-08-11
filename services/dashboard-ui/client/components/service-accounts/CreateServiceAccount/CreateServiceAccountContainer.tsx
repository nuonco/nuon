import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useRoleOptions } from '@/hooks/use-roles'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createServiceAccount } from '@/lib'
import { CreateServiceAccountModal } from './CreateServiceAccount'

const CreateServiceAccountModalContainer = (props: Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()
  const { roleOptions } = useRoleOptions('service_account')

  const { mutate, isPending, error } = useMutation({
    mutationFn: ({ name, role }: { name: string; role: string }) =>
      createServiceAccount({ body: { name, role }, orgId: org.id }),
    onSuccess: (_data, { name }) => {
      queryClient.invalidateQueries({ queryKey: ['service-accounts', org.id] })
      addToast(
        <Toast heading="Service account created" theme="success">
          <Text>Created {name} for {org.name}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  return (
    <CreateServiceAccountModal
      roleOptions={roleOptions}
      isPending={isPending}
      error={error}
      onSubmit={({ name, role }) => mutate({ name, role })}
      {...props}
    />
  )
}

export const CreateServiceAccountButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <CreateServiceAccountModalContainer />

  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PlusIcon" />
      Create service account
    </Button>
  )
}
