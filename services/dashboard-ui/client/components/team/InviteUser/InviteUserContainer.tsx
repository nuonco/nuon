import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { inviteUser } from '@/lib'
import { InviteUserModal } from './InviteUser'

const InviteUserModalContainer = (props: Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const hasSupportRole = !!org?.features?.['support-role']

  const { mutate, isPending, error } = useMutation({
    mutationFn: ({ email, roleType }: { email: string; roleType: string }) =>
      inviteUser({
        body: { email, ...(hasSupportRole ? { role_type: roleType } : {}) },
        orgId: org.id,
      }),
    onSuccess: (_data, { email }) => {
      queryClient.invalidateQueries({ queryKey: ['org-invites', org?.id] })
      addToast(
        <Toast heading="Invitation sent" theme="success">
          <Text>An invitation has been sent to {email}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: (_err, { email }) => {
      addToast(
        <Toast heading="Invite failed" theme="error">
          <Text>There was an error inviting {email} to {org.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <InviteUserModal
      hasSupportRole={hasSupportRole}
      isPending={isPending}
      error={error}
      onSubmit={({ email, roleType }) => mutate({ email, roleType })}
      {...props}
    />
  )
}

export const InviteUserButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <InviteUserModalContainer />

  return (
    <Button
      variant="secondary"
      onClick={() => addModal(modal)}
      {...props}
    >
      {!props?.isMenuButton ? <Icon variant="UserPlusIcon" /> : null}
      Invite user
      {props?.isMenuButton ? <Icon variant="UserPlusIcon" /> : null}
    </Button>
  )
}
