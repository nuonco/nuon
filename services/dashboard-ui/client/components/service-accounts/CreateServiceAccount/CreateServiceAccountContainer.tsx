import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createServiceAccount, listRoles } from '@/lib'
import { CreateServiceAccountModal } from './CreateServiceAccount'

const CreateServiceAccountModalContainer = (props: Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { data: roles } = useQuery({
    queryKey: ['roles', org.id],
    queryFn: () => listRoles({ orgId: org.id }),
  })

  const roleOptions = (roles ?? [])
    .filter((role) => role.applies_to?.includes('service_account'))
    .map((role) => ({ value: role.role_type, label: role.title }))

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
    onError: () => {
      addToast(
        <Toast heading="Create failed" theme="error">
          <Text>There was an error creating the service account.</Text>
        </Toast>
      )
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
