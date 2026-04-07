import { useMutation } from '@tanstack/react-query'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { resendOrgInvite } from '@/lib'
import type { TOrgInvite } from '@/types'
import { ResendOrgInviteModal } from './ResendOrgInvite'

interface IResendOrgInvite {
  invite: TOrgInvite
}

const ResendOrgInviteModalContainer = ({
  invite,
  ...props
}: IResendOrgInvite & Record<string, any>) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending, error } = useMutation({
    mutationFn: () => resendOrgInvite({ inviteId: invite.id, orgId: org.id }),
    onSuccess: () => {
      addToast(
        <Toast heading="Invite resent" theme="success">
          <Text>Invitation has been resent to {invite.email}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Resend failed" theme="error">
          <Text>Failed to resend invite to {invite.email}.</Text>
        </Toast>
      )
    },
  })

  return (
    <ResendOrgInviteModal
      email={invite.email || ''}
      isPending={isPending}
      error={error}
      onSubmit={() => mutate()}
      {...props}
    />
  )
}

export const ResendOrgInviteButton = ({
  invite,
  ...props
}: IResendOrgInvite & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ResendOrgInviteModalContainer invite={invite} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      Resend
      {props?.isMenuButton ? <Icon variant="Envelope" /> : null}
    </Button>
  )
}
