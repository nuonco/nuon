import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { deleteRole } from '@/lib'
import type { TRoleInfo } from '@/types'
import { DeleteRoleModal } from './DeleteRole'

const roleLabel = (role: TRoleInfo) => role.title || 'This role'

const DeleteRoleModalContainer = ({
  role,
  ...props
}: { role: TRoleInfo } & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate, isPending, error } = useMutation({
    mutationFn: () => deleteRole({ roleId: role.id!, orgId: org.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] })
      addToast(
        <Toast heading="Role deleted" theme="success">
          <Text>
            {roleLabel(role)} was deleted and revoked from everyone holding it.
          </Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Delete failed" theme="error">
          <Text>There was an error deleting {roleLabel(role)}.</Text>
        </Toast>
      )
    },
  })

  return (
    <DeleteRoleModal
      roleTitle={roleLabel(role)}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const DeleteRoleButton = ({
  role,
  ...props
}: { role: TRoleInfo } & Omit<IButtonAsButton, 'children' | 'role'>) => {
  const { addModal } = useSurfaces()
  const modal = <DeleteRoleModalContainer role={role} />

  return (
    <Button
      variant="ghost"
      className="!text-red-800 dark:!text-red-500 !p-2 w-full justify-between"
      onClick={() => addModal(modal)}
      {...props}
    >
      Delete role
      {props?.isMenuButton ? <Icon variant="TrashIcon" /> : null}
    </Button>
  )
}
