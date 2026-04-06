import { useMutation } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { resendOrgInvite } from '@/lib'
import type { TOrgInvite } from '@/types'

interface IResendOrgInvite {
  invite: TOrgInvite
}

export const ResendOrgInviteModal = ({
  invite,
  ...props
}: IResendOrgInvite & IModal) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()

  const { mutate, isPending: isLoading, error } = useMutation({
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
    <Modal
      heading={
        <Text flex className="gap-4" variant="h3" weight="strong">
          <Icon variant="Envelope" size="24" />
          Resend invite
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Resending...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="Envelope" />
            Resend invite
          </span>
        ),
        disabled: isLoading,
        onClick: () => mutate(),
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to resend invite'}
          </Banner>
        ) : null}
        <Text variant="base">
          Resend the invitation email to <strong>{invite.email}</strong>?
        </Text>
      </div>
    </Modal>
  )
}

export const ResendOrgInviteButton = ({
  invite,
  ...props
}: IResendOrgInvite & IButtonAsButton) => {
  const { addModal } = useSurfaces()
  const modal = <ResendOrgInviteModal invite={invite} />

  return (
    <Button onClick={() => addModal(modal)} {...props}>
      Resend
      {props?.isMenuButton ? <Icon variant="Envelope" /> : null}
    </Button>
  )
}
