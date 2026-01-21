'use client'

import { useState } from 'react'
import { usePathname } from 'next/navigation'
import { inviteUser } from '@/actions/orgs/invite-user'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'

export const InviteUserModal = ({ ...props }: IModal) => {
  const path = usePathname()
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const [email, setEmail] = useState('')

  const { data, error, isLoading, execute } = useServerAction({
    action: inviteUser,
  })

  useServerActionToast({
    data,
    error,
    errorContent: (
      <>
        <Text>
          There was an error while trying to invite {email} to {org.name}.
        </Text>
        <Text>{error?.error || 'Unknown error occurred.'}</Text>
      </>
    ),
    errorHeading: `Invite failed`,
    onSuccess: () => {
      removeModal(props.modalId)
    },
    successContent: <Text>An invitation has been sent to {email}.</Text>,
    successHeading: `Invitation sent`,
  })

  const handleInvite = () => {
    if (!email) return
    execute({
      body: { email },
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
          <Icon variant="UserPlus" size="24" />
          Invite team member
        </Text>
      }
      primaryActionTrigger={{
        children: isLoading ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" /> Inviting...
          </span>
        ) : (
          <span className="flex items-center gap-2">
            <Icon variant="UserPlus" />
            Invite user
          </span>
        ),
        disabled: !email || isLoading,
        onClick: handleInvite,
        variant: 'primary',
      }}
      {...props}
    >
      <div className="flex flex-col gap-6">
        {error ? (
          <Banner theme="error">
            {error?.error || 'Unable to invite user to organization'}
          </Banner>
        ) : null}

        <div className="flex flex-col gap-2">
          <Label htmlFor="invite-email">
            Email address of the user you want to invite
          </Label>
          <Input
            id="invite-email"
            placeholder="user@email.com"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </div>
      </div>
    </Modal>
  )
}

export const InviteUserButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <InviteUserModal />

  return (
    <Button
      variant="secondary"
      onClick={() => {
        addModal(modal)
      }}
      {...props}
    >
      {!props?.isMenuButton ? <Icon variant="UserPlus" /> : null}
      Invite user
      {props?.isMenuButton ? <Icon variant="UserPlus" /> : null}
    </Button>
  )
}
