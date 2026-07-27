import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getOrgInvites, inviteUser, resendOrgInvite } from '@/lib'
import { InviteUserModal } from './InviteUser'

const InviteUserModalContainer = (props: Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { data: invites } = useQuery({
    queryKey: ['org-invites', org?.id],
    queryFn: () => getOrgInvites({ orgId: org.id }),
    enabled: !!org?.id,
  })

  const { mutate, isPending, error } = useMutation({
    mutationFn: ({ email, roleType }: { email: string; roleType: string }) =>
      inviteUser({
        body: { email, role_type: roleType },
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

  const { mutate: resend, isPending: isResendPending } = useMutation({
    mutationFn: ({ inviteId }: { inviteId: string }) =>
      resendOrgInvite({ inviteId, orgId: org.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-invites', org?.id] })
      addToast(
        <Toast heading="Invite resent" theme="success">
          <Text>The invitation has been resent.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Resend failed" theme="error">
          <Text>Failed to resend the invite.</Text>
        </Toast>
      )
    },
  })

  return (
    <InviteUserModal
      isPending={isPending}
      isResendPending={isResendPending}
      error={error}
      invites={invites}
      onSubmit={({ email, roleType }) => mutate({ email, roleType })}
      onResend={(inviteId) => resend({ inviteId })}
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
