import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { revokeOrgInvite } from '@/lib'
import type { TOrgInvite } from '@/types'
import { RevokeOrgInviteModal } from './RevokeOrgInvite'

interface IRevokeOrgInvite {
  invite: TOrgInvite
}

const RevokeOrgInviteModalContainer = ({
  invite,
  ...props
}: IRevokeOrgInvite & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const queryClient = useQueryClient()

  const { mutate, isPending, error } = useMutation({
    mutationFn: () => revokeOrgInvite({ inviteId: invite.id, orgId: org.id }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['org-invites', org?.id] })
      addToast(
        <Toast heading="Invite revoked" theme="success">
          <Text>Invitation for {invite.email} has been revoked.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Revoke failed" theme="error">
          <Text>Failed to revoke invite for {invite.email}.</Text>
        </Toast>
      )
    },
  })

  return (
    <RevokeOrgInviteModal
      email={invite.email || ''}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const RevokeOrgInviteButton = ({
  invite,
  ...props
}: IRevokeOrgInvite & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <RevokeOrgInviteModalContainer invite={invite} />

  return (
    <Button variant="danger" onClick={() => addModal(modal)} {...props}>
      Revoke
      {props?.isMenuButton ? <Icon variant="ProhibitIcon" /> : null}
    </Button>
  )
}
