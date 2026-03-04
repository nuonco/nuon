import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useToast } from '@/hooks/use-toast'
import { useSurfaces } from '@/hooks/use-surfaces'
import { inviteUser } from '@/lib'

export const InviteUserModal = ({ ...props }: IModal) => {
  const { org } = useOrg()
  const { removeModal } = useSurfaces()
  const { addToast } = useToast()
  const [email, setEmail] = useState('')

  const { mutate, isPending: isLoading, error } = useMutation({
    mutationFn: () => inviteUser({ body: { email }, orgId: org.id }),
    onSuccess: () => {
      addToast(
        <Toast heading="Invitation sent" theme="success">
          <Text>An invitation has been sent to {email}.</Text>
        </Toast>
      )
      removeModal(props.modalId)
    },
    onError: () => {
      addToast(
        <Toast heading="Invite failed" theme="error">
          <Text>There was an error inviting {email} to {org.name}.</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading={
        <Text className="inline-flex gap-4 items-center" variant="h3" weight="strong">
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
        onClick: () => mutate(),
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
      onClick={() => addModal(modal)}
      {...props}
    >
      {!props?.isMenuButton ? <Icon variant="UserPlus" /> : null}
      Invite user
      {props?.isMenuButton ? <Icon variant="UserPlus" /> : null}
    </Button>
  )
}
