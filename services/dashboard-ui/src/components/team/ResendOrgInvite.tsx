'use client'

import { usePathname } from 'next/navigation'
import { resendOrgInvite } from '@/actions/orgs/resend-org-invite'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import type { TOrgInvite } from '@/types'

interface IResendOrgInvite {
  invite: TOrgInvite
}

export const ResendOrgInviteModal = ({
  invite,
  ...props
}: IResendOrgInvite & IModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()

  const { data, error, isLoading, execute } = useServerAction({
    action: resendOrgInvite,
  })

  useServerActionToast({
    data,
    error,
    errorContent: (
      <>
        <Text>Failed to resend invite to {invite.email}.</Text>
        <Text>{error?.error || 'Unknown error occurred.'}</Text>
      </>
    ),
    errorHeading: 'Resend failed',
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: <Text>Invitation has been resent to {invite.email}.</Text>,
    successHeading: 'Invite resent',
  })

  const handleResend = () => {
    execute({
      inviteId: invite.id,
      orgId: org.id,
      path,
    })
  }

  return (
    <Modal
      heading={
        <Text
          className="inline-flex gap-4 items-center"
          variant="h3"
          weight="strong"
        >
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
        onClick: handleResend,
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
    <Button
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >    
      Resend
      {props?.isMenuButton ? <Icon variant="Envelope" /> : null}
    </Button>
  )
}
