import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { createRole, updateRole } from '@/lib'
import type { TRoleInfo } from '@/types'
import { entriesFromRole } from '../permissions'
import { RoleFormModal } from './RoleForm'
import type { RoleFormValues } from './schema'

const RoleFormModalContainer = ({
  role,
  ...props
}: { role?: TRoleInfo } & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const mode = role ? 'edit' : 'create'

  const { mutate, isPending, error } = useMutation({
    mutationFn: (values: RoleFormValues) => {
      const body = {
        title: values.title,
        description: values.description,
        contexts: values.contexts,
        permissions: values.permissions,
      }

      return role?.id
        ? updateRole({ roleId: role.id, body, orgId: org.id })
        : createRole({ body, orgId: org.id })
    },
    onSuccess: (_data, values) => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      queryClient.invalidateQueries({ queryKey: ['role', org?.id] })
      addToast(
        <Toast
          heading={mode === 'create' ? 'Role created' : 'Role updated'}
          theme="success"
        >
          <Text>
            {values.title} is ready to assign in {org.name}.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
  })

  return (
    <RoleFormModal
      mode={mode}
      orgId={org?.id ?? ''}
      initialValues={
        role
          ? {
              title: role.title ?? '',
              description: role.description ?? '',
              contexts: (role.applies_to ?? []) as RoleFormValues['contexts'],
              permissions: entriesFromRole(role.policies),
            }
          : undefined
      }
      isPending={isPending}
      error={error}
      onSubmit={(values) => mutate(values)}
      {...props}
    />
  )
}

export const CreateRoleButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <RoleFormModalContainer />

  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="PlusIcon" />
      Create role
    </Button>
  )
}

export const EditRoleButton = ({
  role,
  ...props
}: { role: TRoleInfo } & Omit<IButtonAsButton, 'children' | 'role'>) => {
  const { addModal } = useSurfaces()
  const modal = <RoleFormModalContainer role={role} />

  return (
    <Button
      variant="ghost"
      className="!p-2 w-full justify-between"
      onClick={() => addModal(modal)}
      {...props}
    >
      Edit role
      {props?.isMenuButton ? <Icon variant="PencilSimpleIcon" /> : null}
    </Button>
  )
}
